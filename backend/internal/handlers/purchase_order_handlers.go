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

// DTOs for creating a purchase order

type PurchaseOrderItemInput struct {
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitCost  float64 `json:"unit_cost"`
}

type CreatePurchaseOrderInput struct {
	SupplierID int64                    `json:"supplier_id"`
	Items      []PurchaseOrderItemInput `json:"items"`
}

// CreatePurchaseOrder handles POST /api/v1/purchase-orders
func CreatePurchaseOrder(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var in CreatePurchaseOrderInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if len(in.Items) == 0 {
			http.Error(w, "items required", http.StatusBadRequest)
			return
		}

		order := &models.PurchaseOrder{
			UserID: userID,
			Status: "pending",
		}
		if in.SupplierID > 0 {
			order.SupplierID.Int64 = in.SupplierID
			order.SupplierID.Valid = true
		}

		items := make([]models.PurchaseOrderItem, 0, len(in.Items))
		for _, it := range in.Items {
			if it.Quantity <= 0 {
				http.Error(w, "quantity must be > 0", http.StatusBadRequest)
				return
			}
			if it.UnitCost < 0 {
				http.Error(w, "unit_cost must be >= 0", http.StatusBadRequest)
				return
			}
			items = append(items, models.PurchaseOrderItem{
				ProductID: it.ProductID,
				Quantity:  it.Quantity,
				UnitCost:  it.UnitCost,
			})
		}

		pom := &models.PurchaseOrderModel{DB: db}
		if err := pom.Create(order, items); err != nil {
			http.Error(w, "could not create purchase order", http.StatusInternalServerError)
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

// GetPurchaseOrders handles GET /api/v1/purchase-orders
func GetPurchaseOrders(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		pom := &models.PurchaseOrderModel{DB: db}
		orders, err := pom.GetAllForUser(userID)
		if err != nil {
			http.Error(w, "could not fetch purchase orders", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orders)
	}
}

// GetPurchaseOrderByID handles GET /api/v1/purchase-orders/{id}
func GetPurchaseOrderByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		pom := &models.PurchaseOrderModel{DB: db}
		order, items, err := pom.GetByID(id, userID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch purchase order", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"order": order,
			"items": items,
		})
	}
}

// UpdatePurchaseOrderStatus handles PUT /api/v1/purchase-orders/{id}/status
func UpdatePurchaseOrderStatus(db *pgxpool.Pool) http.HandlerFunc {
	type statusInput struct {
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		var in statusInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if in.Status == "" {
			http.Error(w, "status required", http.StatusBadRequest)
			return
		}

		pom := &models.PurchaseOrderModel{DB: db}
		if err := pom.UpdateStatus(id, userID, in.Status); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			// Log the actual error for debugging
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
