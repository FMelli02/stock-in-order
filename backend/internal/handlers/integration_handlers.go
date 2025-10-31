package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"stock-in-order/backend/internal/models"
	"stock-in-order/backend/internal/services"
)

// IntegrationHandlers maneja las rutas de integraciones
type IntegrationHandlers struct {
	IntegrationModel    *models.IntegrationModel
	MercadoLibreService *services.MercadoLibreService
	FrontendURL         string
}

// NewIntegrationHandlers crea una nueva instancia de IntegrationHandlers
func NewIntegrationHandlers(
	integrationModel *models.IntegrationModel,
	mlService *services.MercadoLibreService,
	frontendURL string,
) *IntegrationHandlers {
	return &IntegrationHandlers{
		IntegrationModel:    integrationModel,
		MercadoLibreService: mlService,
		FrontendURL:         frontendURL,
	}
}

// HandleMercadoLibreConnect inicia el flujo OAuth2 con Mercado Libre
// Redirige al usuario a la página de autorización de Mercado Libre
// GET /api/v1/integrations/mercadolibre/connect
func (h *IntegrationHandlers) HandleMercadoLibreConnect(w http.ResponseWriter, r *http.Request) {
	// Obtener el user_id del contexto (viene del middleware de autenticación)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		slog.Error("HandleMercadoLibreConnect: user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Usar el user_id como state para rastrear quién inició el proceso
	state := fmt.Sprintf("%d", userID)

	// Generar la URL de autorización
	authURL := h.MercadoLibreService.GetAuthorizationURL(state)

	slog.Info("HandleMercadoLibreConnect: redirecting to MercadoLibre",
		"user_id", userID,
		"auth_url", authURL)

	// Redirigir al usuario a Mercado Libre
	http.Redirect(w, r, authURL, http.StatusFound) // 302
}

// HandleMercadoLibreCallback maneja el callback de Mercado Libre después de la autorización
// GET /api/v1/integrations/mercadolibre/callback?code=...&state=...
func (h *IntegrationHandlers) HandleMercadoLibreCallback(w http.ResponseWriter, r *http.Request) {
	// Obtener los parámetros de la query
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Verificar si el usuario rechazó la autorización
	if errorParam != "" {
		slog.Warn("HandleMercadoLibreCallback: user denied authorization", "error", errorParam)
		redirectURL := fmt.Sprintf("%s/integrations?success=false&error=denied", h.FrontendURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Validar que tenemos code y state
	if code == "" || state == "" {
		slog.Error("HandleMercadoLibreCallback: missing code or state")
		redirectURL := fmt.Sprintf("%s/integrations?success=false&error=invalid_params", h.FrontendURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Convertir el state (user_id) a int64
	userID, err := strconv.ParseInt(state, 10, 64)
	if err != nil {
		slog.Error("HandleMercadoLibreCallback: invalid state", "state", state, "error", err)
		redirectURL := fmt.Sprintf("%s/integrations?success=false&error=invalid_state", h.FrontendURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	slog.Info("HandleMercadoLibreCallback: processing callback",
		"user_id", userID,
		"code", code[:10]+"...") // Solo logueamos parte del código por seguridad

	// Intercambiar el código por tokens
	tokenResp, err := h.MercadoLibreService.ExchangeCodeForToken(code)
	if err != nil {
		slog.Error("HandleMercadoLibreCallback: failed to exchange code for token",
			"user_id", userID,
			"error", err)
		redirectURL := fmt.Sprintf("%s/integrations?success=false&error=token_exchange_failed", h.FrontendURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	slog.Info("HandleMercadoLibreCallback: successfully obtained tokens",
		"user_id", userID,
		"ml_user_id", tokenResp.UserID,
		"expires_in", tokenResp.ExpiresIn)

	// Calcular la fecha de expiración
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Convertir el ML UserID a string
	externalUserID := fmt.Sprintf("%d", tokenResp.UserID)

	// Crear la integración
	integration := &models.Integration{
		UserID:         userID,
		Platform:       "mercadolibre",
		ExternalUserID: &externalUserID,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		ExpiresAt:      expiresAt,
	}

	// Guardar o actualizar la integración (usando upsert)
	err = h.IntegrationModel.UpsertByUserAndPlatform(integration)
	if err != nil {
		slog.Error("HandleMercadoLibreCallback: failed to save integration",
			"user_id", userID,
			"error", err)
		redirectURL := fmt.Sprintf("%s/integrations?success=false&error=database_error", h.FrontendURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	slog.Info("HandleMercadoLibreCallback: integration saved successfully",
		"user_id", userID,
		"integration_id", integration.ID)

	// Redirigir al frontend con éxito
	redirectURL := fmt.Sprintf("%s/integrations?success=true", h.FrontendURL)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleListIntegrations devuelve todas las integraciones del usuario autenticado
// GET /api/v1/integrations
func (h *IntegrationHandlers) HandleListIntegrations(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	integrations, err := h.IntegrationModel.GetAllForUser(userID)
	if err != nil {
		slog.Error("HandleListIntegrations: failed to get integrations",
			"user_id", userID,
			"error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Preparar respuesta (sin tokens)
	type IntegrationResponse struct {
		ID             int64     `json:"id"`
		Platform       string    `json:"platform"`
		ExternalUserID *string   `json:"external_user_id,omitempty"`
		ExpiresAt      time.Time `json:"expires_at"`
		IsExpired      bool      `json:"is_expired"`
		CreatedAt      time.Time `json:"created_at"`
	}

	response := make([]IntegrationResponse, 0, len(integrations))
	for _, integration := range integrations {
		response = append(response, IntegrationResponse{
			ID:             integration.ID,
			Platform:       integration.Platform,
			ExternalUserID: integration.ExternalUserID,
			ExpiresAt:      integration.ExpiresAt,
			IsExpired:      integration.IsTokenExpired(),
			CreatedAt:      integration.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// HandleDeleteIntegration elimina una integración específica
// DELETE /api/v1/integrations/{platform}
func (h *IntegrationHandlers) HandleDeleteIntegration(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	platform := r.PathValue("platform")
	if platform == "" {
		http.Error(w, "Platform is required", http.StatusBadRequest)
		return
	}

	err := h.IntegrationModel.Delete(userID, platform)
	if err != nil {
		if err == models.ErrNotFound {
			http.Error(w, "Integration not found", http.StatusNotFound)
			return
		}
		slog.Error("HandleDeleteIntegration: failed to delete integration",
			"user_id", userID,
			"platform", platform,
			"error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("HandleDeleteIntegration: integration deleted",
		"user_id", userID,
		"platform", platform)

	w.WriteHeader(http.StatusNoContent)
}
