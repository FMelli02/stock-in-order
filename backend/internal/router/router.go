package router

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	"stock-in-order/backend/internal/handlers"
	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/rabbitmq"
)

// SetupRouter wires up HTTP routes.
func SetupRouter(db *pgxpool.Pool, rabbit *rabbitmq.Client, jwtSecret string, logger *slog.Logger) http.Handler {
	r := mux.NewRouter()

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", handlers.Health()).Methods("GET")
	api.HandleFunc("/users/register", handlers.RegisterUser(db)).Methods("POST")
	api.HandleFunc("/users/login", handlers.LoginUser(db, jwtSecret)).Methods("POST")

	// ============================================
	// ADMIN - Gestión de Usuarios
	// ============================================
	api.Handle("/admin/users",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.CreateUserByAdmin(db))),
			jwtSecret,
		),
	).Methods("POST")

	// RBAC Test endpoints (protected by JWT + Role middleware)
	api.Handle("/test/admin-only",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.AdminOnlyTest())),
			jwtSecret,
		),
	).Methods("GET")

	api.Handle("/test/vendedor-only",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.VendedorOnlyTest())),
			jwtSecret,
		),
	).Methods("GET")

	// ============================================
	// PRODUCTS - Con protección RBAC
	// ============================================
	// Lectura: Todos los autenticados (admin, vendedor, repositor)
	api.Handle("/products",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ListProducts(db)), jwtSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProduct(db)), jwtSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}/movements",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProductMovements(db)), jwtSecret)).Methods("GET")

	// Creación: Admin y Repositor
	api.Handle("/products",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.CreateProduct(db))),
			jwtSecret,
		)).Methods("POST")

	// Actualización: Admin y Repositor
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.UpdateProduct(db))),
			jwtSecret,
		)).Methods("PUT")

	// Ajuste de Stock: Solo Repositor y Admin
	api.Handle("/products/{id:[0-9]+}/adjust-stock",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.AdjustProductStock(db))),
			jwtSecret,
		)).Methods("POST")

	// Eliminación: Solo Admin
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteProduct(db))),
			jwtSecret,
		)).Methods("DELETE")

	// ============================================
	// DASHBOARD - Todos los autenticados
	// ============================================
	api.Handle("/dashboard/metrics",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardMetrics(db)), jwtSecret)).Methods("GET")
	api.Handle("/dashboard/kpis",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardKPIs(db)), jwtSecret)).Methods("GET")
	api.Handle("/dashboard/charts",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardCharts(db)), jwtSecret)).Methods("GET")

	// ============================================
	// REPORTS - Todos los autenticados
	// ============================================
	api.Handle("/reports/products/email",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.RequestProductsReportByEmail(db, rabbit)), jwtSecret)).Methods("POST")
	api.Handle("/reports/customers/email",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.RequestCustomersReportByEmail(db, rabbit)), jwtSecret)).Methods("POST")
	api.Handle("/reports/suppliers/email",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.RequestSuppliersReportByEmail(db, rabbit)), jwtSecret)).Methods("POST")

	api.Handle("/reports/products/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportProductsXLSX(db)), jwtSecret)).Methods("GET")
	api.Handle("/reports/customers/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportCustomersXLSX(db)), jwtSecret)).Methods("GET")
	api.Handle("/reports/suppliers/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportSuppliersXLSX(db)), jwtSecret)).Methods("GET")
	api.Handle("/reports/sales-orders/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportSalesOrdersXLSX(db)), jwtSecret)).Methods("GET")
	api.Handle("/reports/purchase-orders/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportPurchaseOrdersXLSX(db)), jwtSecret)).Methods("GET")

	// ============================================
	// SUPPLIERS - Con protección RBAC
	// ============================================
	// Lectura: Todos los autenticados
	api.Handle("/suppliers",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ListSuppliers(db)), jwtSecret)).Methods("GET")
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetSupplier(db)), jwtSecret)).Methods("GET")

	// Creación: Admin y Repositor
	api.Handle("/suppliers",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.CreateSupplier(db))),
			jwtSecret,
		)).Methods("POST")

	// Actualización: Admin y Repositor
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.UpdateSupplier(db))),
			jwtSecret,
		)).Methods("PUT")

	// Eliminación: Solo Admin
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteSupplier(db))),
			jwtSecret,
		)).Methods("DELETE")

	// ============================================
	// CUSTOMERS - Con protección RBAC
	// ============================================
	// Lectura: Admin y Vendedor (repositor NO puede ver clientes)
	api.Handle("/customers",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.ListCustomers(db))),
			jwtSecret,
		)).Methods("GET")
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetCustomer(db))),
			jwtSecret,
		)).Methods("GET")

	// Creación: Admin y Vendedor
	api.Handle("/customers",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.CreateCustomer(db))),
			jwtSecret,
		)).Methods("POST")

	// Actualización: Admin y Vendedor
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.UpdateCustomer(db))),
			jwtSecret,
		)).Methods("PUT")

	// Eliminación: Solo Admin
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteCustomer(db))),
			jwtSecret,
		)).Methods("DELETE")

	// ============================================
	// SALES ORDERS - Con protección RBAC
	// ============================================
	// Creación y Lectura: Admin y Vendedor
	api.Handle("/sales-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.CreateSalesOrder(db))),
			jwtSecret,
		)).Methods("POST")
	api.Handle("/sales-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetSalesOrders(db))),
			jwtSecret,
		)).Methods("GET")
	api.Handle("/sales-orders/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetSalesOrderByID(db))),
			jwtSecret,
		)).Methods("GET")

	// ============================================
	// PURCHASE ORDERS - Con protección RBAC
	// ============================================
	// Creación y Gestión: Admin y Repositor
	api.Handle("/purchase-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.CreatePurchaseOrder(db))),
			jwtSecret,
		)).Methods("POST")
	api.Handle("/purchase-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.GetPurchaseOrders(db))),
			jwtSecret,
		)).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.GetPurchaseOrderByID(db))),
			jwtSecret,
		)).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}/status",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.UpdatePurchaseOrderStatus(db))),
			jwtSecret,
		)).Methods("PUT")

	// Configure CORS for Vite dev server and common API usage
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Apply middlewares in order: Sentry (innermost) → Logging → CORS (outermost)
	// This ensures: CORS first, then logging captures the request, then Sentry catches panics, then routes
	handler := middleware.SentryMiddleware(r, logger)
	handler = middleware.LoggingMiddleware(logger)(handler)
	handler = c.Handler(handler)

	return handler
}
