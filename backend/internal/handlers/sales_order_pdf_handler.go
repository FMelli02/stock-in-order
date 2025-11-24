package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"stock-in-order/backend/internal/models"
	"stock-in-order/backend/internal/rabbitmq"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AppContextPDF holds dependencies for PDF handler
type AppContextPDF struct {
	DB       *pgxpool.Pool
	RabbitMQ *rabbitmq.Client
}

// RequestSalesOrderPDFHandler maneja la solicitud para generar un PDF de orden de venta
func RequestSalesOrderPDFHandler(app *AppContextPDF) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener el ID de la orden de la URL
		vars := mux.Vars(r)
		orderIDStr := vars["id"]
		orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}

		// Obtener el user_id del contexto (del token JWT)
		userID := r.Context().Value("user_id").(int64)

		// Verificar que la orden pertenezca al usuario
		orderModel := &models.SalesOrderModel{DB: app.DB}
		order, _, err := orderModel.GetByID(orderID, userID)
		if err != nil {
			if err == models.ErrNotFound {
				http.Error(w, "Order not found", http.StatusNotFound)
				return
			}
			slog.Error("RequestSalesOrderPDFHandler: failed to get order", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Obtener el email del usuario
		userModel := &models.UserModel{DB: app.DB}
		user, err := userModel.GetByID(userID)
		if err != nil {
			slog.Error("RequestSalesOrderPDFHandler: failed to get user", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Publicar mensaje en RabbitMQ
		req := rabbitmq.ReportRequest{
			UserID:     userID,
			Email:      user.Email,
			ReportType: "sales_order_pdf",
			OrderID:    orderID,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = app.RabbitMQ.PublishReportRequest(ctx, req)
		if err != nil {
			slog.Error("RequestSalesOrderPDFHandler: failed to publish message", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		slog.Info("RequestSalesOrderPDFHandler: PDF request queued",
			"orderID", orderID,
			"userID", userID,
			"email", user.Email)

		// Responder con éxito
		response := map[string]interface{}{
			"success": true,
			"message": "PDF being generated. You will receive it by email shortly.",
			"order": map[string]interface{}{
				"id":         order.ID,
				"order_date": order.OrderDate,
				"status":     order.Status,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}
}
