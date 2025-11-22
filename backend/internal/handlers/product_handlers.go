package handlers

import (
	"encoding/json"
	"log/slog"
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
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Name        string `json:"name"`
			SKU         string `json:"sku"`
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
			StockMinimo int    `json:"stock_minimo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Validación: stock_minimo y quantity no pueden ser negativos
		if in.StockMinimo < 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "stock_minimo cannot be negative"})
			return
		}
		if in.Quantity < 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "quantity cannot be negative"})
			return
		}

		p := &models.Product{
			Name:        in.Name,
			SKU:         in.SKU,
			Description: &in.Description,
			StockMinimo: in.StockMinimo,
			UserID:      organizationID,
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

		// Si se proporciona una cantidad inicial, crear un lote automáticamente
		if in.Quantity > 0 {
			batchInsert := `
				INSERT INTO product_batches (product_id, user_id, quantity, lote_number)
				VALUES ($1, $2, $3, $4)`

			loteNumber := "INICIAL-" + p.SKU
			if _, err := db.Exec(r.Context(), batchInsert, p.ID, organizationID, in.Quantity, loteNumber); err != nil {
				slog.Error("Failed to create initial batch", "error", err, "productID", p.ID)
				// No fallar la creación del producto, solo loggear el error
			}
		}

		// Recargar el producto para obtener el quantity calculado
		updatedProduct, err := pm.GetByID(p.ID, organizationID)
		if err != nil {
			slog.Error("Failed to reload product after creation", "error", err, "productID", p.ID)
			// Retornar el producto original aunque no tenga quantity calculado
			updatedProduct = p
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(updatedProduct)
	}
}

// ListProducts handles GET /api/v1/products with pagination
func ListProducts(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			slog.Error("ListProducts: No organization_id in context")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Leer query params para paginación
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

		filters := models.NewFilters(page, pageSize)

		slog.Info("ListProducts called", "organizationID", organizationID, "page", filters.Page, "pageSize", filters.PageSize)

		pm := &models.ProductModel{DB: db}
		items, metadata, err := pm.GetAllForUserPaginated(organizationID, filters)
		if err != nil {
			slog.Error("ListProducts failed", "error", err, "organizationID", organizationID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info("ListProducts result", "organizationID", organizationID, "count", len(items), "totalRecords", metadata.TotalRecords)

		// Respuesta con estructura paginada
		response := models.PaginatedResponse{
			Items:    items,
			Metadata: metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetProduct handles GET /api/v1/products/{id}
func GetProduct(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		pm := &models.ProductModel{DB: db}
		p, err := pm.GetByID(id, organizationID)
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
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
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
			StockMinimo int    `json:"stock_minimo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Validación: stock_minimo y quantity no pueden ser negativos
		if in.StockMinimo < 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "stock_minimo cannot be negative"})
			return
		}
		if in.Quantity < 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "quantity cannot be negative"})
			return
		}

		pm := &models.ProductModel{DB: db}

		// Obtener el producto actual para comparar quantity
		currentProduct, err := pm.GetByID(id, organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch product", http.StatusInternalServerError)
			return
		}

		// Actualizar información básica del producto
		p := &models.Product{
			Name:        in.Name,
			SKU:         in.SKU,
			Description: &in.Description,
			StockMinimo: in.StockMinimo,
		}

		if err := pm.Update(id, organizationID, p); err != nil {
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

		// Si la quantity cambió, ajustar el lote principal
		if in.Quantity != currentProduct.CalculatedQuantity {
			quantityDiff := in.Quantity - currentProduct.CalculatedQuantity

			// Buscar si existe un lote INICIAL para este producto
			var batchID int64
			batchQuery := `
				SELECT id FROM product_batches 
				WHERE product_id = $1 AND user_id = $2 
				ORDER BY created_at ASC 
				LIMIT 1`

			err := db.QueryRow(r.Context(), batchQuery, id, organizationID).Scan(&batchID)
			if err != nil {
				// No existe un lote, crear uno nuevo
				batchInsert := `
					INSERT INTO product_batches (product_id, user_id, quantity, lote_number)
					VALUES ($1, $2, $3, $4)`

				loteNumber := "AJUSTE-" + in.SKU
				if _, err := db.Exec(r.Context(), batchInsert, id, organizationID, in.Quantity, loteNumber); err != nil {
					slog.Error("Failed to create adjustment batch", "error", err, "productID", id)
				}
			} else {
				// Actualizar el lote existente
				batchUpdate := `
					UPDATE product_batches 
					SET quantity = quantity + $1 
					WHERE id = $2 AND user_id = $3`

				if _, err := db.Exec(r.Context(), batchUpdate, quantityDiff, batchID, organizationID); err != nil {
					slog.Error("Failed to update batch quantity", "error", err, "batchID", batchID)
				}
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteProduct handles DELETE /api/v1/products/{id}
func DeleteProduct(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		pm := &models.ProductModel{DB: db}
		if err := pm.Delete(id, organizationID); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			if err == models.ErrHasReferences {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "No se puede eliminar el producto porque está siendo usado en órdenes de venta o compra",
				})
				return
			}
			slog.Error("DeleteProduct failed", "error", err, "productID", id, "organizationID", organizationID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetProductMovements maneja GET /api/v1/products/{id}/movements
func GetProductMovements(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		smm := &models.StockMovementModel{DB: db}
		movements, err := smm.GetForProduct(id, organizationID)
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
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
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
		if err := pm.AdjustStock(id, organizationID, in.QuantityChange, reason); err != nil {
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

// GetProductBatches handles GET /api/v1/products/{id}/batches
// Returns all active batches (quantity > 0) for a product, ordered by expiry date
func GetProductBatches(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		productID, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			http.Error(w, "invalid product ID", http.StatusBadRequest)
			return
		}

		pm := &models.ProductModel{DB: db}
		batches, err := pm.GetBatchesByProductID(productID, organizationID)
		if err != nil {
			slog.Error("Failed to get product batches",
				"productID", productID,
				"organizationID", organizationID,
				"error", err,
			)
			http.Error(w, "could not fetch batches", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(batches)
	}
}
