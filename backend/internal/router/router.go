package router

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	"stock-in-order/backend/internal/config"
	"stock-in-order/backend/internal/handlers"
	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/models"
	"stock-in-order/backend/internal/rabbitmq"
	"stock-in-order/backend/internal/services"
)

// SetupRouter wires up HTTP routes.
func SetupRouter(db *pgxpool.Pool, rabbit *rabbitmq.Client, cfg config.Config, logger *slog.Logger) http.Handler {
	r := mux.NewRouter()

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", handlers.Health()).Methods("GET")
	api.HandleFunc("/users/register", handlers.RegisterUser(db)).Methods("POST")
	api.HandleFunc("/users/login", handlers.LoginUser(db, cfg.JWTSecret)).Methods("POST")

	// ============================================
	// ADMIN - Gestión de Usuarios
	// ============================================
	api.Handle("/admin/users",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.CreateUserByAdmin(db))),
			cfg.JWTSecret,
		),
	).Methods("POST")

	// RBAC Test endpoints (protected by JWT + Role middleware)
	api.Handle("/test/admin-only",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.AdminOnlyTest())),
			cfg.JWTSecret,
		),
	).Methods("GET")

	api.Handle("/test/vendedor-only",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.VendedorOnlyTest())),
			cfg.JWTSecret,
		),
	).Methods("GET")

	// ============================================
	// PRODUCTS - Con protección RBAC
	// ============================================
	// Lectura: Todos los autenticados (admin, vendedor, repositor)
	api.Handle("/products",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ListProducts(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProduct(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}/movements",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProductMovements(db)), cfg.JWTSecret)).Methods("GET")

	// Creación: Admin y Repositor
	api.Handle("/products",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.CreateProduct(db))),
			cfg.JWTSecret,
		)).Methods("POST")

	// Actualización: Admin y Repositor
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.UpdateProduct(db))),
			cfg.JWTSecret,
		)).Methods("PUT")

	// Ajuste de Stock: Solo Repositor y Admin
	api.Handle("/products/{id:[0-9]+}/adjust-stock",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.AdjustProductStock(db))),
			cfg.JWTSecret,
		)).Methods("POST")

	// Eliminación: Solo Admin
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteProduct(db))),
			cfg.JWTSecret,
		)).Methods("DELETE")

	// ============================================
	// DASHBOARD - Todos los autenticados
	// ============================================
	api.Handle("/dashboard/metrics",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardMetrics(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/dashboard/kpis",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardKPIs(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/dashboard/charts",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardCharts(db)), cfg.JWTSecret)).Methods("GET")

	// ============================================
	// REPORTS - Todos los autenticados
	// ============================================
	api.Handle("/reports/products/email",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.RequestProductsReportByEmail(db, rabbit)), cfg.JWTSecret)).Methods("POST")
	api.Handle("/reports/customers/email",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.RequestCustomersReportByEmail(db, rabbit)), cfg.JWTSecret)).Methods("POST")
	api.Handle("/reports/suppliers/email",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.RequestSuppliersReportByEmail(db, rabbit)), cfg.JWTSecret)).Methods("POST")

	api.Handle("/reports/products/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportProductsXLSX(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/reports/customers/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportCustomersXLSX(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/reports/suppliers/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportSuppliersXLSX(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/reports/sales-orders/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportSalesOrdersXLSX(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/reports/purchase-orders/xlsx",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportPurchaseOrdersXLSX(db)), cfg.JWTSecret)).Methods("GET")

	// ============================================
	// SUPPLIERS - Con protección RBAC
	// ============================================
	// Lectura: Todos los autenticados
	api.Handle("/suppliers",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ListSuppliers(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetSupplier(db)), cfg.JWTSecret)).Methods("GET")

	// Creación: Admin y Repositor
	api.Handle("/suppliers",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.CreateSupplier(db))),
			cfg.JWTSecret,
		)).Methods("POST")

	// Actualización: Admin y Repositor
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.UpdateSupplier(db))),
			cfg.JWTSecret,
		)).Methods("PUT")

	// Eliminación: Solo Admin
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteSupplier(db))),
			cfg.JWTSecret,
		)).Methods("DELETE")

	// ============================================
	// CUSTOMERS - Con protección RBAC
	// ============================================
	// Lectura: Admin y Vendedor (repositor NO puede ver clientes)
	api.Handle("/customers",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.ListCustomers(db))),
			cfg.JWTSecret,
		)).Methods("GET")
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetCustomer(db))),
			cfg.JWTSecret,
		)).Methods("GET")

	// Creación: Admin y Vendedor
	api.Handle("/customers",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.CreateCustomer(db))),
			cfg.JWTSecret,
		)).Methods("POST")

	// Actualización: Admin y Vendedor
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.UpdateCustomer(db))),
			cfg.JWTSecret,
		)).Methods("PUT")

	// Eliminación: Solo Admin
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteCustomer(db))),
			cfg.JWTSecret,
		)).Methods("DELETE")

	// ============================================
	// SALES ORDERS - Con protección RBAC
	// ============================================
	// Creación y Lectura: Admin y Vendedor
	api.Handle("/sales-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.CreateSalesOrder(db))),
			cfg.JWTSecret,
		)).Methods("POST")
	api.Handle("/sales-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetSalesOrders(db))),
			cfg.JWTSecret,
		)).Methods("GET")
	api.Handle("/sales-orders/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetSalesOrderByID(db))),
			cfg.JWTSecret,
		)).Methods("GET")

	// ============================================
	// PURCHASE ORDERS - Con protección RBAC
	// ============================================
	// Creación y Gestión: Admin y Repositor
	api.Handle("/purchase-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.CreatePurchaseOrder(db))),
			cfg.JWTSecret,
		)).Methods("POST")
	api.Handle("/purchase-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.GetPurchaseOrders(db))),
			cfg.JWTSecret,
		)).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.GetPurchaseOrderByID(db))),
			cfg.JWTSecret,
		)).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}/status",
		middleware.JWTMiddleware(
			middleware.RequireRole("repositor")(http.HandlerFunc(handlers.UpdatePurchaseOrderStatus(db))),
			cfg.JWTSecret,
		)).Methods("PUT")

	// ============================================
	// INTEGRATIONS - OAuth2 y gestión de integraciones
	// ============================================
	// Inicializar modelos y servicios
	integrationModel := &models.IntegrationModel{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}
	mlService := services.NewMercadoLibreService(cfg.MLClientID, cfg.MLClientSecret, cfg.MLRedirectURI)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	integrationHandlers := handlers.NewIntegrationHandlers(integrationModel, mlService, frontendURL)

	// Listar integraciones del usuario (protegido)
	api.Handle("/integrations",
		middleware.JWTMiddleware(
			http.HandlerFunc(integrationHandlers.HandleListIntegrations),
			cfg.JWTSecret,
		)).Methods("GET")

	// Eliminar integración (protegido)
	api.Handle("/integrations/{platform}",
		middleware.JWTMiddleware(
			http.HandlerFunc(integrationHandlers.HandleDeleteIntegration),
			cfg.JWTSecret,
		)).Methods("DELETE")

	// OAuth2 - Iniciar conexión con Mercado Libre (protegido)
	api.Handle("/integrations/mercadolibre/connect",
		middleware.JWTMiddleware(
			http.HandlerFunc(integrationHandlers.HandleMercadoLibreConnect),
			cfg.JWTSecret,
		)).Methods("GET")

	// OAuth2 - Callback de Mercado Libre (público, no requiere JWT)
	api.HandleFunc("/integrations/mercadolibre/callback",
		integrationHandlers.HandleMercadoLibreCallback).Methods("GET")

	// ============================================
	// WEBHOOKS - Notificaciones de plataformas externas
	// ============================================
	webhookHandlers := handlers.NewMercadoLibreWebhookHandlers(rabbit)

	// Webhook de Mercado Libre (público, llamado por Meli)
	api.HandleFunc("/webhooks/mercadolibre",
		webhookHandlers.HandleMercadoLibreWebhook).Methods("POST")

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
