package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/models"
)

// DTOs for creating a sales order

type OrderItemInput struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type CreateOrderInput struct {
	CustomerID int64            `json:"customer_id"`
	Items      []OrderItemInput `json:"items"`
}

// CreateSalesOrder handles POST /api/v1/sales-orders
func CreateSalesOrder(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var in CreateOrderInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if len(in.Items) == 0 {
			http.Error(w, "items required", http.StatusBadRequest)
			return
		}

		// Build model structs
		order := &models.SalesOrder{
			UserID: userID,
			Status: "pending",
		}
		if in.CustomerID > 0 {
			order.CustomerID.Int64 = in.CustomerID
			order.CustomerID.Valid = true
		}

		items := make([]models.OrderItem, 0, len(in.Items))
		for _, it := range in.Items {
			if it.Quantity <= 0 {
				http.Error(w, "quantity must be > 0", http.StatusBadRequest)
				return
			}
			items = append(items, models.OrderItem{
				ProductID: it.ProductID,
				Quantity:  it.Quantity,
				UnitPrice: 0,
			})
		}

		som := &models.SalesOrderModel{DB: db}
		if err := som.Create(order, items); err != nil {
			if err == models.ErrInsufficientStock {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "insufficient stock"})
				return
			}
			http.Error(w, "could not create order", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"order": order,
			"items": items,
		})
	}
}

// GetSalesOrders handles GET /api/v1/sales-orders
func GetSalesOrders(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		som := &models.SalesOrderModel{DB: db}
		orders, err := som.GetAllForUser(userID)
		if err != nil {
			http.Error(w, "could not fetch orders", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orders)
	}
}

// GetSalesOrderByID handles GET /api/v1/sales-orders/{id}
func GetSalesOrderByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		som := &models.SalesOrderModel{DB: db}
		order, items, err := som.GetByID(id, userID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch order", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"order": order,
			"items": items,
		})
	}
}
