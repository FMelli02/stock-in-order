package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/mercadopago"
	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/models"
)

// ============================================
// DTOs
// ============================================

// CreateCheckoutRequest DTO para crear checkout
type CreateCheckoutRequest struct {
	PlanID string `json:"plan_id" validate:"required"` // plan_basico, plan_pro, plan_enterprise
}

// CreateCheckoutResponse respuesta con URL de checkout
type CreateCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	InitPoint   string `json:"init_point"`
}

// SubscriptionStatusResponse estado de suscripción del usuario
type SubscriptionStatusResponse struct {
	UserID           int64               `json:"user_id"`
	PlanID           string              `json:"plan_id"`
	Status           string              `json:"status"`
	CurrentPeriodEnd *string             `json:"current_period_end,omitempty"`
	Features         models.PlanFeatures `json:"features"`
	AvailablePlans   []AvailablePlan     `json:"available_plans"`
}

// AvailablePlan plan disponible para upgrade
type AvailablePlan struct {
	PlanID      string              `json:"plan_id"`
	Name        string              `json:"name"`
	Price       float64             `json:"price"`
	Description string              `json:"description"`
	Features    models.PlanFeatures `json:"features"`
}

// ============================================
// HANDLERS
// ============================================

// CreateCheckoutHandler maneja POST /api/v1/subscriptions/create-checkout
// Crea un checkout de MercadoPago para el plan seleccionado
func CreateCheckoutHandler(db *pgxpool.Pool, mpClient *mercadopago.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Obtener usuario del JWT
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, ok := middleware.UserEmailFromContext(r.Context())
		if !ok {
			http.Error(w, "email not found in token", http.StatusUnauthorized)
			return
		}

		// 2. Parsear request
		var req CreateCheckoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Validar que el plan_id sea válido
		planID := models.PlanID(req.PlanID)
		if !isValidPlan(planID) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid plan_id. Valid values: plan_basico, plan_pro, plan_enterprise",
			})
			return
		}

		// No permitir checkout para plan gratuito
		if planID == models.PlanFree {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "cannot create checkout for free plan",
			})
			return
		}

		// 3. Verificar suscripción actual del usuario
		sm := &models.SubscriptionModel{DB: db}
		currentSub, err := sm.GetByUserID(organizationID)
		if err != nil {
			if err != models.ErrNotFound {
				slog.Error("Error getting subscription", "error", err, "organizationID", organizationID)
				http.Error(w, "error getting subscription", http.StatusInternalServerError)
				return
			}
		}

		// Si ya tiene el plan solicitado y está activo
		if currentSub != nil && currentSub.PlanID == planID && currentSub.Status == models.SubscriptionStatusActive {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "user already has this plan active",
			})
			return
		}

		// 4. Crear preferencia de pago en MercadoPago
		mpRequest := mercadopago.CreatePreferenceRequest{
			UserID:    organizationID,
			UserEmail: userEmail,
			PlanID:    planID,
		}

		mpResponse, err := mpClient.CreatePreference(mpRequest)
		if err != nil {
			slog.Error("Error creating MercadoPago preference", "error", err, "organizationID", organizationID)
			http.Error(w, "error creating payment link", http.StatusInternalServerError)
			return
		}

		// 5. Responder con la URL de checkout
		response := CreateCheckoutResponse{
			CheckoutURL: mpResponse.CheckoutURL,
			InitPoint:   mpResponse.InitPoint,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		slog.Info("Checkout created",
			"organizationID", organizationID,
			"plan", planID,
			"checkoutURL", mpResponse.CheckoutURL)
	}
}

// CreateRecurringSubscriptionHandler maneja POST /api/v1/subscriptions/create-recurring
// Crea una suscripción recurrente en MercadoPago
func CreateRecurringSubscriptionHandler(db *pgxpool.Pool, mpClient *mercadopago.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Obtener usuario del JWT
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, ok := middleware.UserEmailFromContext(r.Context())
		if !ok {
			http.Error(w, "email not found in token", http.StatusUnauthorized)
			return
		}

		// 2. Parsear request
		var req CreateCheckoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		planID := models.PlanID(req.PlanID)
		if !isValidPlan(planID) || planID == models.PlanFree {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid plan_id for recurring subscription",
			})
			return
		}

		// 3. Crear suscripción recurrente en MercadoPago
		mpRequest := mercadopago.CreateSubscriptionRequest{
			UserID:    organizationID,
			UserEmail: userEmail,
			PlanID:    planID,
		}

		mpResponse, err := mpClient.CreateSubscription(mpRequest)
		if err != nil {
			slog.Error("Error creating recurring subscription", "error", err, "organizationID", organizationID)
			http.Error(w, "error creating subscription", http.StatusInternalServerError)
			return
		}

		// 4. Responder con la URL de checkout
		response := CreateCheckoutResponse{
			CheckoutURL: mpResponse.CheckoutURL,
			InitPoint:   mpResponse.InitPoint,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		slog.Info("Recurring subscription created",
			"organizationID", organizationID,
			"plan", planID,
			"preapprovalID", mpResponse.PreapprovalID)
	}
}

// GetSubscriptionStatusHandler maneja GET /api/v1/subscriptions/status
// Retorna el estado de la suscripción del usuario actual
func GetSubscriptionStatusHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener usuario del JWT
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener suscripción
		sm := &models.SubscriptionModel{DB: db}
		sub, err := sm.GetByUserID(organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			slog.Error("Error getting subscription", "error", err, "organizationID", organizationID)
			http.Error(w, "error getting subscription", http.StatusInternalServerError)
			return
		}

		// Construir respuesta
		features := models.GetPlanFeatures(sub.PlanID)

		var periodEnd *string
		if sub.CurrentPeriodEnd != nil {
			endStr := sub.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z07:00")
			periodEnd = &endStr
		}

		response := SubscriptionStatusResponse{
			UserID:           organizationID,
			PlanID:           string(sub.PlanID),
			Status:           string(sub.Status),
			CurrentPeriodEnd: periodEnd,
			Features:         features,
			AvailablePlans:   getAvailablePlans(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// CancelSubscriptionHandler maneja POST /api/v1/subscriptions/cancel
// Cancela la suscripción del usuario
func CancelSubscriptionHandler(db *pgxpool.Pool, mpClient *mercadopago.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener suscripción actual
		sm := &models.SubscriptionModel{DB: db}
		sub, err := sm.GetByUserID(organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			slog.Error("Error getting subscription", "error", err, "organizationID", organizationID)
			http.Error(w, "error getting subscription", http.StatusInternalServerError)
			return
		}

		// No se puede cancelar plan gratuito
		if sub.PlanID == models.PlanFree {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "cannot cancel free plan",
			})
			return
		}

		// Si tiene preapproval_id de MercadoPago, cancelar en MP también
		if sub.MPPreapprovalID.Valid && sub.MPPreapprovalID.String != "" {
			if err := mpClient.CancelSubscription(sub.MPPreapprovalID.String); err != nil {
				slog.Error("Error canceling subscription in MercadoPago", "error", err)
				// Continuamos de todas formas para cancelar en nuestra DB
			}
		}

		// Cancelar en nuestra base de datos
		if err := sm.Cancel(sub.ID, "user_requested"); err != nil {
			slog.Error("Error canceling subscription", "error", err, "organizationID", organizationID)
			http.Error(w, "error canceling subscription", http.StatusInternalServerError)
			return
		}

		// Downgrade a plan gratuito
		if err := sm.UpdatePlan(sub.ID, models.PlanFree); err != nil {
			slog.Error("Error downgrading to free plan", "error", err, "organizationID", organizationID)
		}

		w.WriteHeader(http.StatusNoContent)
		slog.Info("Subscription canceled", "organizationID", organizationID, "plan", sub.PlanID)
	}
}

// ============================================
// HELPERS
// ============================================

// isValidPlan valida si el plan existe
func isValidPlan(planID models.PlanID) bool {
	switch planID {
	case models.PlanFree, models.PlanBasico, models.PlanPro, models.PlanEnterprise:
		return true
	default:
		return false
	}
}

// getAvailablePlans retorna los planes disponibles
func getAvailablePlans() []AvailablePlan {
	plans := []models.PlanID{
		models.PlanFree,
		models.PlanBasico,
		models.PlanPro,
		models.PlanEnterprise,
	}

	var available []AvailablePlan
	for _, planID := range plans {
		features := models.GetPlanFeatures(planID)
		price := models.GetPlanPrice(planID)

		available = append(available, AvailablePlan{
			PlanID:      string(planID),
			Name:        getPlanName(planID),
			Price:       price,
			Description: getPlanDescription(planID),
			Features:    features,
		})
	}

	return available
}

// getPlanName retorna el nombre del plan
func getPlanName(planID models.PlanID) string {
	switch planID {
	case models.PlanFree:
		return "Plan Gratuito"
	case models.PlanBasico:
		return "Plan Básico"
	case models.PlanPro:
		return "Plan Profesional"
	case models.PlanEnterprise:
		return "Plan Empresarial"
	default:
		return "Plan Desconocido"
	}
}

// getPlanDescription retorna la descripción del plan
func getPlanDescription(planID models.PlanID) string {
	switch planID {
	case models.PlanFree:
		return "Ideal para empezar - Funcionalidades básicas"
	case models.PlanBasico:
		return "Para pequeñas empresas - Incluye reportes y lotes"
	case models.PlanPro:
		return "Para empresas en crecimiento - API y multi-almacén"
	case models.PlanEnterprise:
		return "Sin límites - Soporte prioritario e integraciones"
	default:
		return ""
	}
}
