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

// CreateProduct handles POST /api/v1/products
func CreateProduct(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Name        string `json:"name"`
			SKU         string `json:"sku"`
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		p := &models.Product{
			Name:        in.Name,
			SKU:         in.SKU,
			Description: in.Description,
			Quantity:    in.Quantity,
			UserID:      userID,
		}

		pm := &models.ProductModel{DB: db}
		if err := pm.Insert(p); err != nil {
			if err == models.ErrDuplicateSKU {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "sku already exists"})
				return
			}
			http.Error(w, "could not create product", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(p)
	}
}

// ListProducts handles GET /api/v1/products
func ListProducts(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		pm := &models.ProductModel{DB: db}
		items, err := pm.GetAllForUser(userID)
		if err != nil {
			http.Error(w, "could not fetch products", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
	}
}

// GetProduct handles GET /api/v1/products/{id}
func GetProduct(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		pm := &models.ProductModel{DB: db}
		p, err := pm.GetByID(id, userID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch product", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(p)
	}
}

// UpdateProduct handles PUT /api/v1/products/{id}
func UpdateProduct(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		var in struct {
			Name        string `json:"name"`
			SKU         string `json:"sku"`
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		p := &models.Product{
			Name:        in.Name,
			SKU:         in.SKU,
			Description: in.Description,
			Quantity:    in.Quantity,
		}

		pm := &models.ProductModel{DB: db}
		if err := pm.Update(id, userID, p); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			if err == models.ErrDuplicateSKU {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "sku already exists"})
				return
			}
			http.Error(w, "could not update product", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteProduct handles DELETE /api/v1/products/{id}
func DeleteProduct(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		pm := &models.ProductModel{DB: db}
		if err := pm.Delete(id, userID); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not delete product", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetProductMovements maneja GET /api/v1/products/{id}/movements
func GetProductMovements(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		smm := &models.StockMovementModel{DB: db}
		movements, err := smm.GetForProduct(id, userID)
		if err != nil {
			http.Error(w, "could not fetch movements", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(movements)
	}
}

// Constante estandarizada para razón de ajuste manual
const ReasonManualAdjustment = "MANUAL_ADJUSTMENT"

// StockAdjustmentInput DTO para ajuste manual
type StockAdjustmentInput struct {
	QuantityChange int    `json:"quantity_change" validate:"required,ne=0"`
	Reason         string `json:"reason" validate:"required"`
}

// AdjustProductStock maneja POST /api/v1/products/{id}/adjust-stock
func AdjustProductStock(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		var in StockAdjustmentInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := validate.Struct(in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "validation failed", "details": err.Error()})
			return
		}

		// Normalizamos la razón; si viene vacía la reemplazamos por la constante
		reason := in.Reason
		if reason == "" {
			reason = ReasonManualAdjustment
		}

		pm := &models.ProductModel{DB: db}
		if err := pm.AdjustStock(id, userID, in.QuantityChange, reason); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not adjust stock", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
