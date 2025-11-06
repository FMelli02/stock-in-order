package mercadopago

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preapproval"
	"github.com/mercadopago/sdk-go/pkg/preference"

	"stock-in-order/backend/internal/models"
)

// Client encapsula las operaciones con MercadoPago
type Client struct {
	accessToken string
	appURL      string // URL base de la aplicación para callbacks
}

// NewClient crea un nuevo cliente de MercadoPago
func NewClient() (*Client, error) {
	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("MP_ACCESS_TOKEN no configurado")
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:5173" // Valor por defecto para desarrollo
	}

	slog.Info("Cliente de MercadoPago inicializado", "appURL", appURL)

	return &Client{
		accessToken: accessToken,
		appURL:      appURL,
	}, nil
}

// ============================================
// PREFERENCIAS DE PAGO (Pago Único)
// ============================================

// CreatePreferenceRequest representa la solicitud para crear una preferencia
type CreatePreferenceRequest struct {
	UserID    int64
	UserEmail string
	PlanID    models.PlanID
}

// CreatePreferenceResponse respuesta con la URL de checkout
type CreatePreferenceResponse struct {
	PreferenceID string `json:"preference_id"`
	CheckoutURL  string `json:"checkout_url"`
	InitPoint    string `json:"init_point"`
}

// CreatePreference crea una preferencia de pago único en MercadoPago
func (c *Client) CreatePreference(req CreatePreferenceRequest) (*CreatePreferenceResponse, error) {
	ctx := context.Background()

	// Configurar el SDK
	cfg, err := config.New(c.accessToken)
	if err != nil {
		slog.Error("Error configurando MercadoPago SDK", "error", err)
		return nil, err
	}

	// Obtener detalles del plan
	planPrice := models.GetPlanPrice(req.PlanID)
	planName := c.getPlanName(req.PlanID)

	// Crear cliente de preferencias
	client := preference.NewClient(cfg)

	// Construir la preferencia
	request := preference.Request{
		Items: []preference.ItemRequest{
			{
				ID:          string(req.PlanID),
				Title:       planName,
				Description: fmt.Sprintf("Suscripción a %s de Stock In Order", planName),
				CategoryID:  "services",
				Quantity:    1,
				UnitPrice:   planPrice,
				CurrencyID:  "ARS",
			},
		},
		Payer: &preference.PayerRequest{
			Email: req.UserEmail,
		},
		BackURLs: &preference.BackURLsRequest{
			Success: fmt.Sprintf("%s/dashboard?payment=success", c.appURL),
			Failure: fmt.Sprintf("%s/dashboard?payment=failure", c.appURL),
			Pending: fmt.Sprintf("%s/dashboard?payment=pending", c.appURL),
		},
		AutoReturn:          "approved",
		NotificationURL:     fmt.Sprintf("%s/api/v1/webhooks/mercadopago", c.appURL),
		StatementDescriptor: "Stock In Order",
		ExternalReference:   fmt.Sprintf("user_%d_plan_%s", req.UserID, req.PlanID),
		Metadata: map[string]interface{}{
			"user_id": req.UserID,
			"plan_id": string(req.PlanID),
		},
	}

	// Crear la preferencia
	response, err := client.Create(ctx, request)
	if err != nil {
		slog.Error("Error creando preferencia de pago", "error", err, "userID", req.UserID, "plan", req.PlanID)
		return nil, fmt.Errorf("error creando preferencia: %w", err)
	}

	slog.Info("Preferencia de pago creada",
		"preferenceID", response.ID,
		"userID", req.UserID,
		"plan", req.PlanID,
		"amount", planPrice)

	return &CreatePreferenceResponse{
		PreferenceID: response.ID,
		CheckoutURL:  response.InitPoint,
		InitPoint:    response.InitPoint,
	}, nil
}

// ============================================
// SUSCRIPCIONES RECURRENTES (Plan)
// ============================================

// CreateSubscriptionRequest solicitud para crear suscripción recurrente
type CreateSubscriptionRequest struct {
	UserID    int64
	UserEmail string
	PlanID    models.PlanID
}

// CreateSubscriptionResponse respuesta con detalles de la suscripción
type CreateSubscriptionResponse struct {
	PreapprovalID string `json:"preapproval_id"`
	InitPoint     string `json:"init_point"`
	CheckoutURL   string `json:"checkout_url"`
}

// CreateSubscription crea una suscripción recurrente en MercadoPago
func (c *Client) CreateSubscription(req CreateSubscriptionRequest) (*CreateSubscriptionResponse, error) {
	ctx := context.Background()

	// Configurar el SDK
	cfg, err := config.New(c.accessToken)
	if err != nil {
		slog.Error("Error configurando MercadoPago SDK", "error", err)
		return nil, err
	}

	// Obtener detalles del plan
	planPrice := models.GetPlanPrice(req.PlanID)
	planName := c.getPlanName(req.PlanID)

	// Validar que no sea plan gratuito
	if req.PlanID == models.PlanFree {
		return nil, errors.New("no se puede crear suscripción para plan gratuito")
	}

	// Crear cliente de preapproval (suscripciones)
	client := preapproval.NewClient(cfg)

	// Calcular fechas (inicio ahora, sin fecha de fin para suscripción continua)
	now := time.Now()

	// Construir la suscripción
	request := preapproval.Request{
		Reason:            planName,
		ExternalReference: fmt.Sprintf("user_%d_plan_%s", req.UserID, req.PlanID),
		PayerEmail:        req.UserEmail,
		AutoRecurring: &preapproval.AutoRecurringRequest{
			Frequency:         1,
			FrequencyType:     "months",
			TransactionAmount: planPrice,
			CurrencyID:        "ARS",
			StartDate:         &now,
		},
		BackURL: fmt.Sprintf("%s/dashboard?subscription=success", c.appURL),
		Status:  "pending",
	}

	// Crear la suscripción
	response, err := client.Create(ctx, request)
	if err != nil {
		slog.Error("Error creando suscripción recurrente", "error", err, "userID", req.UserID, "plan", req.PlanID)
		return nil, fmt.Errorf("error creando suscripción: %w", err)
	}

	slog.Info("Suscripción recurrente creada",
		"preapprovalID", response.ID,
		"userID", req.UserID,
		"plan", req.PlanID,
		"amount", planPrice)

	return &CreateSubscriptionResponse{
		PreapprovalID: response.ID,
		InitPoint:     response.InitPoint,
		CheckoutURL:   response.InitPoint,
	}, nil
}

// ============================================
// GESTIÓN DE SUSCRIPCIONES
// ============================================

// CancelSubscription cancela una suscripción en MercadoPago
func (c *Client) CancelSubscription(preapprovalID string) error {
	ctx := context.Background()

	// Configurar el SDK
	cfg, err := config.New(c.accessToken)
	if err != nil {
		slog.Error("Error configurando MercadoPago SDK", "error", err)
		return err
	}

	// Crear cliente
	client := preapproval.NewClient(cfg)

	// Actualizar la suscripción a cancelada
	updateRequest := preapproval.UpdateRequest{
		Status: "cancelled",
	}

	_, err = client.Update(ctx, preapprovalID, updateRequest)
	if err != nil {
		slog.Error("Error cancelando suscripción en MercadoPago", "error", err, "preapprovalID", preapprovalID)
		return fmt.Errorf("error cancelando suscripción: %w", err)
	}

	slog.Info("Suscripción cancelada en MercadoPago", "preapprovalID", preapprovalID)
	return nil
}

// GetSubscription obtiene los detalles de una suscripción
func (c *Client) GetSubscription(preapprovalID string) (*preapproval.Response, error) {
	ctx := context.Background()

	// Configurar el SDK
	cfg, err := config.New(c.accessToken)
	if err != nil {
		slog.Error("Error configurando MercadoPago SDK", "error", err)
		return nil, err
	}

	// Crear cliente
	client := preapproval.NewClient(cfg)

	// Obtener la suscripción
	response, err := client.Get(ctx, preapprovalID)
	if err != nil {
		slog.Error("Error obteniendo suscripción de MercadoPago", "error", err, "preapprovalID", preapprovalID)
		return nil, fmt.Errorf("error obteniendo suscripción: %w", err)
	}

	return response, nil
}

// ============================================
// HELPERS
// ============================================

// getPlanName retorna el nombre legible del plan
func (c *Client) getPlanName(planID models.PlanID) string {
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

// GetPlanDescription retorna la descripción del plan
func (c *Client) GetPlanDescription(planID models.PlanID) string {
	features := models.GetPlanFeatures(planID)
	return fmt.Sprintf(
		"Hasta %d productos, %d órdenes/mes, %d usuarios",
		features.MaxProducts,
		features.MaxOrders,
		features.MaxUsers,
	)
}

// ============================================
// WEBHOOKS Y NOTIFICACIONES
// ============================================

// GetPayment obtiene los detalles de un pago por su ID
func (c *Client) GetPayment(paymentID string) (*payment.Response, error) {
	ctx := context.Background()

	// Configurar el SDK
	cfg, err := config.New(c.accessToken)
	if err != nil {
		slog.Error("Error configurando MercadoPago SDK", "error", err)
		return nil, err
	}

	// Convertir paymentID a int64
	paymentIDInt, err := strconv.ParseInt(paymentID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID format: %w", err)
	}

	// Crear cliente de pagos
	client := payment.NewClient(cfg)

	// Obtener el pago (el SDK espera int, no int64)
	response, err := client.Get(ctx, int(paymentIDInt))
	if err != nil {
		slog.Error("Error obteniendo pago de MercadoPago", "error", err, "paymentID", paymentID)
		return nil, fmt.Errorf("error obteniendo pago: %w", err)
	}

	return response, nil
}

// GetAccessToken retorna el access token (útil para validaciones)
func (c *Client) GetAccessToken() string {
	return c.accessToken
}
