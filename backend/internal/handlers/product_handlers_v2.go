package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/models"
	"stock-in-order/backend/internal/repository"
)

// App holds dependencies for handlers
type App struct {
	DB        *pgxpool.Pool
	AuditRepo *repository.AuditRepository
}

// CreateProductV2 handles POST /api/v1/products with audit logging
func (app *App) CreateProductV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

		var in struct {
			Name        string `json:"name"`
			SKU         string `json:"sku"`
			Description string `json:"description"`
			StockMinimo int    `json:"stock_minimo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Validación: stock_minimo no puede ser negativo
		if in.StockMinimo < 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "stock_minimo cannot be negative"})
			return
		}

		p := &models.Product{
			Name:        in.Name,
			SKU:         in.SKU,
			Description: &in.Description,
			StockMinimo: in.StockMinimo,
			UserID:      organizationID,
		}

		pm := &models.ProductModel{DB: app.DB}
		if err := pm.Insert(p); err != nil {
			if err == models.ErrDuplicateSKU {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "sku already exists"})
				return
			}
			http.Error(w, "could not create product", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar creación de producto
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionCreate,
			EntityType: models.EntityTypeProduct,
			EntityID:   &p.ID,
			Details:    fmt.Sprintf("Usuario %s (%s) creó el producto '%s' (ID: %d, SKU: %s)", userEmail, userRole, p.Name, p.ID, p.SKU),
			Timestamp:  time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(p)
	}
}

// UpdateProductV2 handles PUT /api/v1/products/{id} with audit logging
func (app *App) UpdateProductV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		var in struct {
			Name        string `json:"name"`
			SKU         string `json:"sku"`
			Description string `json:"description"`
			StockMinimo int    `json:"stock_minimo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Validación: stock_minimo no puede ser negativo
		if in.StockMinimo < 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "stock_minimo cannot be negative"})
			return
		}

		p := &models.Product{
			Name:        in.Name,
			SKU:         in.SKU,
			Description: &in.Description,
			StockMinimo: in.StockMinimo,
		}

		pm := &models.ProductModel{DB: app.DB}
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

		// 🔔 AUDITORÍA: Registrar actualización de producto
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionUpdate,
			EntityType: models.EntityTypeProduct,
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) actualizó el producto '%s' (ID: %d, SKU: %s)", userEmail, userRole, in.Name, id, in.SKU),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteProductV2 handles DELETE /api/v1/products/{id} with audit logging
func (app *App) DeleteProductV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		// Obtener el producto antes de eliminarlo para guardar su nombre en el log
		pm := &models.ProductModel{DB: app.DB}
		product, err := pm.GetByID(id, organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch product", http.StatusInternalServerError)
			return
		}

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
			slog.Error("DeleteProduct failed", "error", err, "productID", id, "userID", organizationID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar eliminación de producto
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionDelete,
			EntityType: models.EntityTypeProduct,
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) eliminó el producto '%s' (ID: %d, SKU: %s)", userEmail, userRole, product.Name, id, product.SKU),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// AdjustProductStockV2 handles POST /api/v1/products/{id}/adjust-stock with audit logging
func (app *App) AdjustProductStockV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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

		// Obtener el producto para incluir su nombre en el log
		pm := &models.ProductModel{DB: app.DB}
		product, err := pm.GetByID(id, organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch product", http.StatusInternalServerError)
			return
		}

		if err := pm.AdjustStock(id, organizationID, in.QuantityChange, reason); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not adjust stock", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar ajuste de stock
		adjustAction := "ADJUST_STOCK"
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     adjustAction,
			EntityType: models.EntityTypeProduct,
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) ajustó el stock del producto '%s' en %+d unidades. Razón: %s", userEmail, userRole, product.Name, in.QuantityChange, reason),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// ============================================
// CUSTOMER HANDLERS WITH AUDIT
// ============================================

// CreateCustomerV2 handles POST /api/v1/customers with audit logging
func (app *App) CreateCustomerV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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

		cm := &models.CustomerModel{DB: app.DB}
		if err := cm.Insert(c); err != nil {
			http.Error(w, "could not create customer", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar creación de cliente
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionCreate,
			EntityType: models.EntityTypeCustomer,
			EntityID:   &c.ID,
			Details:    fmt.Sprintf("Usuario %s (%s) creó el cliente '%s' (ID: %d, Email: %s)", userEmail, userRole, c.Name, c.ID, c.Email),
			Timestamp:  time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(c)
	}
}

// UpdateCustomerV2 handles PUT /api/v1/customers/{id} with audit logging
func (app *App) UpdateCustomerV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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

		cm := &models.CustomerModel{DB: app.DB}
		if err := cm.Update(id, organizationID, c); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not update customer", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar actualización de cliente
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionUpdate,
			EntityType: models.EntityTypeCustomer,
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) actualizó el cliente '%s' (ID: %d)", userEmail, userRole, in.Name, id),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteCustomerV2 handles DELETE /api/v1/customers/{id} with audit logging
func (app *App) DeleteCustomerV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		// Obtener el cliente antes de eliminarlo para guardar su nombre en el log
		cm := &models.CustomerModel{DB: app.DB}
		customer, err := cm.GetByID(id, organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch customer", http.StatusInternalServerError)
			return
		}

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
			slog.Error("DeleteCustomer failed", "error", err, "customerID", id, "userID", organizationID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar eliminación de cliente
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionDelete,
			EntityType: models.EntityTypeCustomer,
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) eliminó el cliente '%s' (ID: %d)", userEmail, userRole, customer.Name, id),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// ============================================
// SUPPLIER HANDLERS WITH AUDIT
// ============================================

// CreateSupplierV2 handles POST /api/v1/suppliers with audit logging
func (app *App) CreateSupplierV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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

		sm := &models.SupplierModel{DB: app.DB}
		if err := sm.Insert(s); err != nil {
			http.Error(w, "could not create supplier", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar creación de proveedor
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionCreate,
			EntityType: "SUPPLIER",
			EntityID:   &s.ID,
			Details:    fmt.Sprintf("Usuario %s (%s) creó el proveedor '%s' (ID: %d, Email: %s)", userEmail, userRole, s.Name, s.ID, s.Email),
			Timestamp:  time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(s)
	}
}

// UpdateSupplierV2 handles PUT /api/v1/suppliers/{id} with audit logging
func (app *App) UpdateSupplierV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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

		sm := &models.SupplierModel{DB: app.DB}
		if err := sm.Update(id, organizationID, s); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not update supplier", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar actualización de proveedor
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionUpdate,
			EntityType: "SUPPLIER",
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) actualizó el proveedor '%s' (ID: %d)", userEmail, userRole, in.Name, id),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteSupplierV2 handles DELETE /api/v1/suppliers/{id} with audit logging
func (app *App) DeleteSupplierV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)

		// Obtener el proveedor antes de eliminarlo
		sm := &models.SupplierModel{DB: app.DB}
		supplier, err := sm.GetByID(id, organizationID)
		if err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "could not fetch supplier", http.StatusInternalServerError)
			return
		}

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
			slog.Error("DeleteSupplier failed", "error", err, "supplierID", id, "userID", organizationID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar eliminación de proveedor
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionDelete,
			EntityType: "SUPPLIER",
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) eliminó el proveedor '%s' (ID: %d)", userEmail, userRole, supplier.Name, id),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// ============================================
// SALES ORDER HANDLERS WITH AUDIT
// ============================================

// CreateSalesOrderV2 handles POST /api/v1/sales-orders with audit logging
func (app *App) CreateSalesOrderV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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
			UserID: organizationID,
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

		som := &models.SalesOrderModel{DB: app.DB}
		if err := som.Create(order, items); err != nil {
			if err == models.ErrInsufficientStock {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "insufficient stock"})
				return
			}
			http.Error(w, "could not create order", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar creación de orden de venta
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionCreate,
			EntityType: models.EntityTypeOrder,
			EntityID:   &order.ID,
			Details:    fmt.Sprintf("Usuario %s (%s) creó una orden de venta (ID: %d) con %d items", userEmail, userRole, order.ID, len(items)),
			Timestamp:  time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"order": order,
			"items": items,
		})
	}
}

// ============================================
// PURCHASE ORDER HANDLERS WITH AUDIT
// ============================================

// CreatePurchaseOrderV2 handles POST /api/v1/purchase-orders with audit logging
func (app *App) CreatePurchaseOrderV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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
			UserID: organizationID,
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

		pom := &models.PurchaseOrderModel{DB: app.DB}
		if err := pom.Create(order, items); err != nil {
			http.Error(w, "could not create purchase order", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar creación de orden de compra
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionCreate,
			EntityType: "PURCHASE_ORDER",
			EntityID:   &order.ID,
			Details:    fmt.Sprintf("Usuario %s (%s) creó una orden de compra (ID: %d) con %d items", userEmail, userRole, order.ID, len(items)),
			Timestamp:  time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"order": order,
			"items": items,
		})
	}
}

// UpdatePurchaseOrderStatusV2 handles PUT /api/v1/purchase-orders/{id}/status with audit logging
func (app *App) UpdatePurchaseOrderStatusV2() http.HandlerFunc {
	type statusInput struct {
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userEmail, _ := middleware.UserEmailFromContext(r.Context())
		userRole, _ := middleware.UserRoleFromContext(r.Context())

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

		pom := &models.PurchaseOrderModel{DB: app.DB}
		if err := pom.UpdateStatus(id, organizationID, in.Status); err != nil {
			if err == models.ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar cambio de estado de orden de compra
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &organizationID,
			UserEmail:  userEmail,
			UserRole:   userRole,
			Action:     models.ActionUpdate,
			EntityType: "PURCHASE_ORDER",
			EntityID:   &id,
			Details:    fmt.Sprintf("Usuario %s (%s) cambió el estado de la orden de compra %d a '%s'", userEmail, userRole, id, in.Status),
			Timestamp:  time.Now(),
		})

		w.WriteHeader(http.StatusNoContent)
	}
}

// ============================================
// USER HANDLERS WITH AUDIT
// ============================================

// CreateUserByAdminV2 handles POST /admin/users with audit logging
func (app *App) CreateUserByAdminV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener info del admin que está creando el usuario
		adminID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		adminEmail, _ := middleware.UserEmailFromContext(r.Context())
		adminRole, _ := middleware.UserRoleFromContext(r.Context())

		var in CreateUserByAdminInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if err := validate.Struct(in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "validation failed", "details": err.Error()})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "could not hash password", http.StatusInternalServerError)
			return
		}

		user := &models.User{
			Name:         in.Name,
			Email:        in.Email,
			PasswordHash: hash,
			Role:         in.Role,
		}

		um := &models.UserModel{DB: app.DB}
		if err := um.Insert(user); err != nil {
			if err == models.ErrDuplicateEmail {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "email already exists"})
				return
			}
			http.Error(w, "could not create user", http.StatusInternalServerError)
			return
		}

		// 🔔 AUDITORÍA: Registrar creación de usuario por admin
		app.AuditRepo.Log(&models.AuditLog{
			UserID:     &adminID,
			UserEmail:  adminEmail,
			UserRole:   adminRole,
			Action:     models.ActionCreate,
			EntityType: models.EntityTypeUser,
			EntityID:   &user.ID,
			Details:    fmt.Sprintf("Admin %s creó el usuario '%s' (Email: %s, Rol: %s)", adminEmail, user.Name, user.Email, user.Role),
			Timestamp:  time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		})
	}
}
