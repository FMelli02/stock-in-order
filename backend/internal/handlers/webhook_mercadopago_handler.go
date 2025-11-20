package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/mercadopago"
	"stock-in-order/backend/internal/models"
)

// ============================================
// DTOs para Webhooks de MercadoPago
// ============================================

// MercadoPagoWebhookNotification estructura de notificación de MercadoPago
type MercadoPagoWebhookNotification struct {
	ID            int64  `json:"id"`
	LiveMode      bool   `json:"live_mode"`
	Type          string `json:"type"` // "payment", "subscription_preapproval", etc.
	DateCreated   string `json:"date_created"`
	ApplicationID int64  `json:"application_id"`
	organizationID        int64  `json:"user_id"`
	Version       int    `json:"version"`
	APIVersion    string `json:"api_version"`
	Action        string `json:"action"` // "payment.created", "payment.updated", etc.
	Data          struct {
		ID string `json:"id"` // ID del recurso (payment ID, preapproval ID, etc.)
	} `json:"data"`
}

// ============================================
// WEBHOOK HANDLER
// ============================================

// HandleMercadoPagoWebhook maneja las notificaciones de webhook de MercadoPago
// Endpoint PÚBLICO (no requiere autenticación JWT)
func HandleMercadoPagoWebhook(db *pgxpool.Pool, mpClient *mercadopago.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Leer el body completo
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error leyendo body del webhook", "error", err)
			http.Error(w, "error reading request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Log de la notificación recibida
		slog.Info("Webhook recibido de MercadoPago",
			"content-type", r.Header.Get("Content-Type"),
			"x-signature", r.Header.Get("X-Signature"),
			"x-request-id", r.Header.Get("X-Request-Id"))

		// 1. VALIDAR LA FIRMA (Crítico para seguridad)
		if !validateWebhookSignature(r, body, mpClient.GetAccessToken()) {
			slog.Warn("Webhook rechazado: firma inválida",
				"ip", r.RemoteAddr,
				"user-agent", r.Header.Get("User-Agent"))
			// Retornar 200 pero no procesar (para evitar reintentos de MercadoPago)
			w.WriteHeader(http.StatusOK)
			return
		}

		// 2. PARSEAR LA NOTIFICACIÓN
		var notification MercadoPagoWebhookNotification
		if err := json.Unmarshal(body, &notification); err != nil {
			slog.Error("Error parseando notificación JSON", "error", err, "body", string(body))
			w.WriteHeader(http.StatusOK) // Retornar 200 para evitar reintentos
			return
		}

		slog.Info("Notificación parseada",
			"type", notification.Type,
			"action", notification.Action,
			"data_id", notification.Data.ID)

		// 3. PROCESAR SEGÚN EL TIPO DE EVENTO
		switch notification.Type {
		case "payment":
			handlePaymentNotification(db, mpClient, notification)
		case "subscription_preapproval", "preapproval":
			handleSubscriptionNotification(db, mpClient, notification)
		default:
			slog.Info("Tipo de notificación no manejado", "type", notification.Type)
		}

		// 4. RESPONDER 200 OK INMEDIATAMENTE
		// MercadoPago espera respuesta rápida, si tardamos mucho reintentará
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// ============================================
// PROCESADORES DE EVENTOS
// ============================================

// handlePaymentNotification procesa notificaciones de pagos
func handlePaymentNotification(db *pgxpool.Pool, mpClient *mercadopago.Client, notification MercadoPagoWebhookNotification) {
	paymentID := notification.Data.ID

	slog.Info("Procesando notificación de pago", "paymentID", paymentID, "action", notification.Action)

	// Obtener detalles completos del pago desde MercadoPago
	payment, err := mpClient.GetPayment(paymentID)
	if err != nil {
		slog.Error("Error obteniendo detalles del pago", "error", err, "paymentID", paymentID)
		return
	}

	slog.Info("Detalles del pago obtenidos",
		"paymentID", paymentID,
		"status", payment.Status,
		"status_detail", payment.StatusDetail,
		"transaction_amount", payment.TransactionAmount,
		"external_reference", payment.ExternalReference)

	// Extraer user_id de metadata o external_reference
	organizationID, planID, err := extractUserAndPlanFromPayment(payment)
	if err != nil {
		slog.Error("Error extrayendo user_id del pago", "error", err, "paymentID", paymentID)
		return
	}

	// Procesar según el estado del pago
	sm := &models.SubscriptionModel{DB: db}
	phm := &models.PaymentHistoryModel{DB: db}

	switch payment.Status {
	case "approved":
		// Pago aprobado: activar suscripción
		slog.Info("Pago aprobado, activando suscripción",
			"organizationID", organizationID,
			"planID", planID,
			"amount", payment.TransactionAmount)

		// Obtener o crear suscripción
		sub, err := sm.GetByUserID(organizationID)
		if err != nil && err != models.ErrNotFound {
			slog.Error("Error obteniendo suscripción", "error", err, "organizationID", organizationID)
			return
		}

		// Si no existe, crear nueva suscripción
		if sub == nil {
			currentPeriodEnd := time.Now().AddDate(0, 1, 0) // +1 mes
			sub = &models.Subscription{
				UserID: organizationID,
				PlanID:           planID,
				Status:           models.SubscriptionStatusActive,
				CurrentPeriodEnd: &currentPeriodEnd,
			}
			if err := sm.Create(sub); err != nil {
				slog.Error("Error creando suscripción", "error", err, "organizationID", organizationID)
				return
			}
		} else {
			// Actualizar plan y activar
			if err := sm.UpdatePlan(sub.ID, planID); err != nil {
				slog.Error("Error actualizando plan", "error", err, "subscriptionID", sub.ID)
				return
			}
			if err := sm.Activate(sub.ID); err != nil {
				slog.Error("Error activando suscripción", "error", err, "subscriptionID", sub.ID)
				return
			}

			// Extender período si es necesario
			currentPeriodEnd := time.Now().AddDate(0, 1, 0)
			sub.CurrentPeriodEnd = &currentPeriodEnd
		}

		// Guardar el payment_id de MercadoPago
		if err := sm.UpdateMPPaymentID(sub.ID, paymentID); err != nil {
			slog.Error("Error guardando MP payment ID", "error", err, "subscriptionID", sub.ID)
		}

		// Registrar en historial de pagos
		paymentHistory := &models.PaymentHistory{
			SubscriptionID: sub.ID,
			UserID: organizationID,
			MPPaymentID:    paymentID,
			MPStatus:       "approved",
			MPStatusDetail: payment.StatusDetail,
			Amount:         payment.TransactionAmount,
			Currency:       payment.CurrencyID,
			Description:    fmt.Sprintf("Pago de %s", planID),
			PlanID:         planID,
			PaymentDate:    time.Now(),
		}
		if err := phm.Create(paymentHistory); err != nil {
			slog.Error("Error creando registro de pago", "error", err)
		}

		slog.Info("Suscripción activada exitosamente",
			"organizationID", organizationID,
			"planID", planID,
			"subscriptionID", sub.ID)

	case "rejected", "cancelled":
		// Pago rechazado o cancelado
		slog.Info("Pago rechazado/cancelado",
			"organizationID", organizationID,
			"status", payment.Status,
			"status_detail", payment.StatusDetail)

		// Registrar en historial
		sub, _ := sm.GetByUserID(organizationID)
		if sub != nil {
			paymentHistory := &models.PaymentHistory{
				SubscriptionID: sub.ID,
				UserID: organizationID,
				MPPaymentID:    paymentID,
				MPStatus:       payment.Status,
				MPStatusDetail: payment.StatusDetail,
				Amount:         payment.TransactionAmount,
				Currency:       payment.CurrencyID,
				Description:    fmt.Sprintf("Pago rechazado: %s", payment.StatusDetail),
				PlanID:         planID,
				PaymentDate:    time.Now(),
			}
			phm.Create(paymentHistory)
		}

	case "pending", "in_process":
		// Pago pendiente
		slog.Info("Pago pendiente de procesamiento",
			"organizationID", organizationID,
			"status", payment.Status)
		// Podríamos registrarlo en historial pero no activar

	default:
		slog.Info("Estado de pago no manejado", "status", payment.Status)
	}
}

// handleSubscriptionNotification procesa notificaciones de suscripciones (preapproval)
func handleSubscriptionNotification(db *pgxpool.Pool, mpClient *mercadopago.Client, notification MercadoPagoWebhookNotification) {
	preapprovalID := notification.Data.ID

	slog.Info("Procesando notificación de suscripción",
		"preapprovalID", preapprovalID,
		"action", notification.Action)

	// Obtener detalles completos de la suscripción desde MercadoPago
	preapproval, err := mpClient.GetSubscription(preapprovalID)
	if err != nil {
		slog.Error("Error obteniendo detalles de la suscripción", "error", err, "preapprovalID", preapprovalID)
		return
	}

	slog.Info("Detalles de suscripción obtenidos",
		"preapprovalID", preapprovalID,
		"status", preapproval.Status,
		"payer_email", preapproval.PayerEmail,
		"external_reference", preapproval.ExternalReference)

	// Extraer user_id y plan_id de external_reference
	// Formato esperado: "user_123_plan_plan_pro"
	organizationID, planID, err := extractUserAndPlanFromReference(preapproval.ExternalReference)
	if err != nil {
		slog.Error("Error extrayendo datos de external_reference", "error", err, "reference", preapproval.ExternalReference)
		return
	}

	sm := &models.SubscriptionModel{DB: db}

	// Procesar según el estado de la suscripción
	switch preapproval.Status {
	case "authorized", "active":
		// Suscripción autorizada/activa
		slog.Info("Suscripción autorizada, activando",
			"organizationID", organizationID,
			"planID", planID,
			"preapprovalID", preapprovalID)

		// Obtener o crear suscripción
		sub, err := sm.GetByUserID(organizationID)
		if err != nil && err != models.ErrNotFound {
			slog.Error("Error obteniendo suscripción", "error", err, "organizationID", organizationID)
			return
		}

		if sub == nil {
			// Crear nueva suscripción
			currentPeriodEnd := time.Now().AddDate(0, 1, 0)
			sub = &models.Subscription{
				UserID: organizationID,
				PlanID:           planID,
				Status:           models.SubscriptionStatusActive,
				CurrentPeriodEnd: &currentPeriodEnd,
			}
			if err := sm.Create(sub); err != nil {
				slog.Error("Error creando suscripción", "error", err, "organizationID", organizationID)
				return
			}
		} else {
			// Actualizar plan y activar
			if err := sm.UpdatePlan(sub.ID, planID); err != nil {
				slog.Error("Error actualizando plan", "error", err)
				return
			}
			if err := sm.Activate(sub.ID); err != nil {
				slog.Error("Error activando suscripción", "error", err)
				return
			}
		}

		// Guardar el preapproval_id de MercadoPago
		if err := sm.UpdateMPPreapprovalID(sub.ID, preapprovalID); err != nil {
			slog.Error("Error guardando MP preapproval ID", "error", err)
		}

		slog.Info("Suscripción recurrente activada",
			"organizationID", organizationID,
			"planID", planID,
			"subscriptionID", sub.ID)

	case "cancelled", "paused":
		// Suscripción cancelada o pausada
		slog.Info("Suscripción cancelada/pausada",
			"organizationID", organizationID,
			"status", preapproval.Status)

		sub, err := sm.GetByUserID(organizationID)
		if err != nil {
			slog.Error("Error obteniendo suscripción para cancelar", "error", err, "organizationID", organizationID)
			return
		}

		if err := sm.Cancel(sub.ID, "mercadopago_"+preapproval.Status); err != nil {
			slog.Error("Error cancelando suscripción", "error", err)
			return
		}

		// Downgrade a plan gratuito
		if err := sm.UpdatePlan(sub.ID, models.PlanFree); err != nil {
			slog.Error("Error downgrading a plan free", "error", err)
		}

		slog.Info("Suscripción cancelada en DB", "organizationID", organizationID)

	default:
		slog.Info("Estado de suscripción no manejado", "status", preapproval.Status)
	}
}

// ============================================
// VALIDACIÓN DE FIRMA
// ============================================

// validateWebhookSignature valida la firma X-Signature del webhook
// Documentación: https://www.mercadopago.com.ar/developers/es/docs/your-integrations/notifications/webhooks
func validateWebhookSignature(r *http.Request, body []byte, accessToken string) bool {
	// Obtener headers de firma
	xSignature := r.Header.Get("X-Signature")
	xRequestID := r.Header.Get("X-Request-Id")

	if xSignature == "" || xRequestID == "" {
		slog.Warn("Webhook sin firma o request ID")
		return false
	}

	// Parsear X-Signature (formato: "ts=123456789,v1=hash")
	parts := strings.Split(xSignature, ",")
	var ts, hash string
	for _, part := range parts {
		if strings.HasPrefix(part, "ts=") {
			ts = strings.TrimPrefix(part, "ts=")
		} else if strings.HasPrefix(part, "v1=") {
			hash = strings.TrimPrefix(part, "v1=")
		}
	}

	if ts == "" || hash == "" {
		slog.Warn("Formato de X-Signature inválido", "xSignature", xSignature)
		return false
	}

	// Construir el string a firmar
	// Formato: "id:<data.id>;request-id:<x-request-id>;ts:<timestamp>;"
	dataID := extractDataIDFromBody(body)
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", dataID, xRequestID, ts)

	// Calcular HMAC-SHA256
	// La secret key es el access token de MercadoPago
	h := hmac.New(sha256.New, []byte(accessToken))
	h.Write([]byte(manifest))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// Comparar
	isValid := hmac.Equal([]byte(hash), []byte(expectedHash))

	if !isValid {
		slog.Warn("Firma HMAC inválida",
			"expected", expectedHash,
			"received", hash,
			"manifest", manifest)
	}

	return isValid
}

// extractDataIDFromBody extrae el data.id del JSON body
func extractDataIDFromBody(body []byte) string {
	var notification struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &notification); err != nil {
		return ""
	}
	return notification.Data.ID
}

// ============================================
// HELPERS PARA EXTRACCIÓN DE DATOS
// ============================================

// extractUserAndPlanFromReference extrae user_id y plan_id de external_reference
// Formato: "user_123_plan_plan_pro"
func extractUserAndPlanFromReference(reference string) (int64, models.PlanID, error) {
	parts := strings.Split(reference, "_")
	if len(parts) < 4 {
		return 0, "", fmt.Errorf("formato de external_reference inválido: %s", reference)
	}

	// Extraer user_id
	organizationID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("user_id inválido en reference: %s", reference)
	}

	// Extraer plan_id (puede ser "plan_pro" o "plan_basico", etc.)
	// Buscar el índice de "plan_" y tomar el resto
	planIndex := strings.Index(reference, "plan_")
	if planIndex == -1 {
		return 0, "", fmt.Errorf("plan_id no encontrado en reference: %s", reference)
	}
	planID := models.PlanID(reference[planIndex:])

	return organizationID, planID, nil
}

// extractUserAndPlanFromPayment extrae user_id y plan_id de un payment
func extractUserAndPlanFromPayment(payment interface{}) (int64, models.PlanID, error) {
	// El payment tiene external_reference con formato "user_123_plan_plan_pro"
	// Necesitamos acceder al campo ExternalReference

	// Type assertion para acceder a ExternalReference
	type PaymentWithReference interface {
		GetExternalReference() string
	}

	// Intentar extraer usando reflection o type assertion
	// Por ahora usaremos un approach simple con JSON marshal/unmarshal
	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return 0, "", err
	}

	var paymentData struct {
		ExternalReference string `json:"external_reference"`
		Metadata          struct {
			organizationID string `json:"user_id"`
			PlanID string `json:"plan_id"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(paymentJSON, &paymentData); err != nil {
		return 0, "", err
	}

	// Primero intentar desde metadata (más confiable)
	if paymentData.Metadata.organizationID != "" && paymentData.Metadata.PlanID != "" {
		organizationID, err := strconv.ParseInt(paymentData.Metadata.organizationID, 10, 64)
		if err != nil {
			return 0, "", err
		}
		return organizationID, models.PlanID(paymentData.Metadata.PlanID), nil
	}

	// Fallback: extraer de external_reference
	if paymentData.ExternalReference != "" {
		return extractUserAndPlanFromReference(paymentData.ExternalReference)
	}

	return 0, "", fmt.Errorf("no se pudo extraer user_id y plan_id del pago")
}
