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

// CreateCustomer handles POST /api/v1/customers
func CreateCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Phone   string `json:"phone"`
			Address string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		c := &models.Customer{
			Name:    in.Name,
			Email:   in.Email,
			Phone:   in.Phone,
			Address: in.Address,
			UserID:  organizationID,
		}

		cm := &models.CustomerModel{DB: db}
		if err := cm.Insert(c); err != nil {
			http.Error(w, "could not create customer", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(c)
	}
}

// ListCustomers handles GET /api/v1/customers
func ListCustomers(db *pgxpool.Pool) http.HandlerFunc {
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

		cm := &models.CustomerModel{DB: db}
		items, metadata, err := cm.GetAllForUserPaginated(organizationID, filters)
		if err != nil {
			http.Error(w, "could not fetch customers", http.StatusInternalServerError)
			return
		}

		slog.Info("Customers fetched with pagination",
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

// GetCustomer handles GET /api/v1/customers/{id}
func GetCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		cm := &models.CustomerModel{DB: db}
		c, err := cm.GetByID(id, organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch customer", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(c)
	}
}

// UpdateCustomer handles PUT /api/v1/customers/{id}
func UpdateCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		var in struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Phone   string `json:"phone"`
			Address string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		c := &models.Customer{
			Name:    in.Name,
			Email:   in.Email,
			Phone:   in.Phone,
			Address: in.Address,
		}

		cm := &models.CustomerModel{DB: db}
		if err := cm.Update(id, organizationID, c); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not update customer", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteCustomer handles DELETE /api/v1/customers/{id}
func DeleteCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		cm := &models.CustomerModel{DB: db}
		if err := cm.Delete(id, organizationID); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			if err == models.ErrHasReferences {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "No se puede eliminar el cliente porque tiene órdenes de venta asociadas",
				})
				return
			}
			slog.Error("DeleteCustomer failed", "error", err, "customerID", id, "organizationID", organizationID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
