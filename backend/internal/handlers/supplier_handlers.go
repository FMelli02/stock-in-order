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

// CreateSupplier handles POST /api/v1/suppliers
func CreateSupplier(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Name          string `json:"name"`
			ContactPerson string `json:"contact_person"`
			Email         string `json:"email"`
			Phone         string `json:"phone"`
			Address       string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		s := &models.Supplier{
			Name:          in.Name,
			ContactPerson: in.ContactPerson,
			Email:         in.Email,
			Phone:         in.Phone,
			Address:       in.Address,
			UserID:        organizationID,
		}

		sm := &models.SupplierModel{DB: db}
		if err := sm.Insert(s); err != nil {
			http.Error(w, "could not create supplier", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(s)
	}
}

// ListSuppliers handles GET /api/v1/suppliers
func ListSuppliers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Leer query params para paginación
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		filters := models.NewFilters(page, pageSize)

		sm := &models.SupplierModel{DB: db}
		items, metadata, err := sm.GetAllForUserPaginated(organizationID, filters)
		if err != nil {
			http.Error(w, "could not fetch suppliers", http.StatusInternalServerError)
			return
		}

		slog.Info("Suppliers fetched with pagination",
			"page", filters.Page,
			"pageSize", filters.PageSize,
			"totalRecords", metadata.TotalRecords,
		)

		response := models.PaginatedResponse{
			Items:    items,
			Metadata: metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetSupplier handles GET /api/v1/suppliers/{id}
func GetSupplier(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		sm := &models.SupplierModel{DB: db}
		s, err := sm.GetByID(id, organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch supplier", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s)
	}
}

// UpdateSupplier handles PUT /api/v1/suppliers/{id}
func UpdateSupplier(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		var in struct {
			Name          string `json:"name"`
			ContactPerson string `json:"contact_person"`
			Email         string `json:"email"`
			Phone         string `json:"phone"`
			Address       string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		s := &models.Supplier{
			Name:          in.Name,
			ContactPerson: in.ContactPerson,
			Email:         in.Email,
			Phone:         in.Phone,
			Address:       in.Address,
		}

		sm := &models.SupplierModel{DB: db}
		if err := sm.Update(id, organizationID, s); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not update supplier", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteSupplier handles DELETE /api/v1/suppliers/{id}
func DeleteSupplier(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		sm := &models.SupplierModel{DB: db}
		if err := sm.Delete(id, organizationID); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			if err == models.ErrHasReferences {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "No se puede eliminar el proveedor porque tiene órdenes de compra asociadas",
				})
				return
			}
			slog.Error("DeleteSupplier failed", "error", err, "supplierID", id, "organizationID", organizationID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
