package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"stock-in-order/backend/internal/rabbitmq"
)

// MercadoLibreWebhookHandlers maneja los webhooks de Mercado Libre
type MercadoLibreWebhookHandlers struct {
	RabbitClient *rabbitmq.Client
}

// NewMercadoLibreWebhookHandlers crea una nueva instancia
func NewMercadoLibreWebhookHandlers(rabbitClient *rabbitmq.Client) *MercadoLibreWebhookHandlers {
	return &MercadoLibreWebhookHandlers{
		RabbitClient: rabbitClient,
	}
}

// MercadoLibreNotification representa la notificación de Mercado Libre
// Documentación: https://developers.mercadolibre.com.ar/es_ar/api-docs-es/notificaciones
type MercadoLibreNotification struct {
	ID            int64  `json:"_id"`
	Resource      string `json:"resource"` // Ej: "/orders/123456789"
	UserID        int64  `json:"user_id"`  // ID del usuario en Mercado Libre
	Topic         string `json:"topic"`    // Ej: "orders_v2", "items", "questions"
	ApplicationID int64  `json:"application_id"`
	Attempts      int    `json:"attempts"`
	Sent          string `json:"sent"`
	Received      string `json:"received"`
}

// HandleMercadoLibreWebhook recibe notificaciones de Mercado Libre
// POST /api/v1/webhooks/mercadolibre
// Este endpoint debe ser PÚBLICO (sin JWT) porque Mercado Libre lo llama
func (h *MercadoLibreWebhookHandlers) HandleMercadoLibreWebhook(w http.ResponseWriter, r *http.Request) {
	// Leer el body de la petición
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("HandleMercadoLibreWebhook: failed to read body", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validación rápida: parsear el JSON
	var notification MercadoLibreNotification
	if err := json.Unmarshal(body, &notification); err != nil {
		slog.Error("HandleMercadoLibreWebhook: failed to parse JSON", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Log de la notificación recibida
	slog.Info("HandleMercadoLibreWebhook: notification received",
		"topic", notification.Topic,
		"resource", notification.Resource,
		"user_id", notification.UserID,
		"notification_id", notification.ID)

	// Filtrar solo notificaciones de órdenes (ventas)
	// Topics relevantes: "orders_v2" (nuevas ventas)
	if notification.Topic != "orders_v2" {
		slog.Info("HandleMercadoLibreWebhook: ignoring non-order notification", "topic", notification.Topic)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Publicar en RabbitMQ para procesamiento asíncrono
	// NO procesamos nada aquí, solo encolamos
	err = h.RabbitClient.PublishMessage(r.Context(), "meli_sales_queue", body)
	if err != nil {
		slog.Error("HandleMercadoLibreWebhook: failed to publish to queue",
			"error", err,
			"notification_id", notification.ID)
		// Aún así respondemos 200 a Mercado Libre para que no reintente inmediatamente
		// El procesamiento se hará cuando RabbitMQ esté disponible nuevamente
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Info("HandleMercadoLibreWebhook: notification enqueued successfully",
		"queue", "meli_sales_queue",
		"notification_id", notification.ID)

	// Respuesta rápida a Mercado Libre (< 500ms recomendado)
	w.WriteHeader(http.StatusOK)
}
