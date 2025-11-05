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
	"stock-in-order/backend/internal/repository"
	"stock-in-order/backend/internal/services"
)

// Application holds dependencies for handlers
type Application struct {
	DB        *pgxpool.Pool
	Rabbit    *rabbitmq.Client
	AuditRepo *repository.AuditRepository
	Config    config.Config
	Logger    *slog.Logger
}

// SetupRouter wires up HTTP routes and receives the AuditRepository for handlers to use
func SetupRouter(db *pgxpool.Pool, rabbit *rabbitmq.Client, auditRepo *repository.AuditRepository, cfg config.Config, logger *slog.Logger) http.Handler {
	r := mux.NewRouter()

	// Create Application struct with all dependencies (for future use)
	_ = &Application{
		DB:        db,
		Rabbit:    rabbit,
		AuditRepo: auditRepo,
		Config:    cfg,
		Logger:    logger,
	}

	// Create handlers App for audited handlers
	handlersApp := &handlers.App{
		DB:        db,
		AuditRepo: auditRepo,
	}

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", handlers.Health()).Methods("GET")
	api.HandleFunc("/users/register", handlers.RegisterUser(db)).Methods("POST")
	api.HandleFunc("/users/login", handlers.LoginUser(db, cfg.JWTSecret)).Methods("POST")

	// ============================================
	// ADMIN - Gestión de Usuarios con Auditoría
	// ============================================
	api.Handle("/admin/users",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(handlersApp.CreateUserByAdminV2()),
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
	// PRODUCTS - Con protección RBAC y Auditoría
	// ============================================
	// Lectura: Todos los autenticados (admin, vendedor, repositor)
	api.Handle("/products",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ListProducts(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProduct(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}/movements",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProductMovements(db)), cfg.JWTSecret)).Methods("GET")

	// Creación: Admin y Repositor (con auditoría)
	api.Handle("/products",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin", "repositor")(handlersApp.CreateProductV2()),
			cfg.JWTSecret,
		)).Methods("POST")

	// Actualización: Admin y Repositor (con auditoría)
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin", "repositor")(handlersApp.UpdateProductV2()),
			cfg.JWTSecret,
		)).Methods("PUT")

	// Ajuste de Stock: Admin y Repositor (con auditoría)
	api.Handle("/products/{id:[0-9]+}/adjust-stock",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin", "repositor")(handlersApp.AdjustProductStockV2()),
			cfg.JWTSecret,
		)).Methods("POST")

	// Eliminación: Solo Admin (con auditoría)
	api.Handle("/products/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(handlersApp.DeleteProductV2()),
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
	// ADMIN - Registro de Auditoría
	// ============================================
	api.Handle("/admin/audit-logs",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(handlers.GetAuditLogs(db)),
			cfg.JWTSecret,
		)).Methods("GET")

	// ============================================
	// SUPPLIERS - Con protección RBAC y Auditoría
	// ============================================
	// Lectura: Todos los autenticados
	api.Handle("/suppliers",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.ListSuppliers(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetSupplier(db)), cfg.JWTSecret)).Methods("GET")

	// Creación: Admin y Repositor (con auditoría)
	api.Handle("/suppliers",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin", "repositor")(handlersApp.CreateSupplierV2()),
			cfg.JWTSecret,
		)).Methods("POST")

	// Actualización: Admin y Repositor (con auditoría)
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin", "repositor")(handlersApp.UpdateSupplierV2()),
			cfg.JWTSecret,
		)).Methods("PUT")

	// Eliminación: Solo Admin (con auditoría)
	api.Handle("/suppliers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(handlersApp.DeleteSupplierV2()),
			cfg.JWTSecret,
		)).Methods("DELETE")

	// ============================================
	// CUSTOMERS - Con protección RBAC y Auditoría
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

	// Creación: Admin y Vendedor (con auditoría)
	api.Handle("/customers",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(handlersApp.CreateCustomerV2()),
			cfg.JWTSecret,
		)).Methods("POST")

	// Actualización: Admin y Vendedor (con auditoría)
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(handlersApp.UpdateCustomerV2()),
			cfg.JWTSecret,
		)).Methods("PUT")

	// Eliminación: Solo Admin (con auditoría)
	api.Handle("/customers/{id:[0-9]+}",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin")(handlersApp.DeleteCustomerV2()),
			cfg.JWTSecret,
		)).Methods("DELETE")

	// ============================================
	// SALES ORDERS - Con protección RBAC y Auditoría
	// ============================================
	// Creación y Lectura: Admin y Vendedor
	api.Handle("/sales-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("vendedor")(handlersApp.CreateSalesOrderV2()),
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
	// PURCHASE ORDERS - Con protección RBAC y Auditoría
	// ============================================
	// Lectura: Todos los autenticados
	api.Handle("/purchase-orders",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetPurchaseOrders(db)), cfg.JWTSecret)).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}",
		middleware.JWTMiddleware(http.HandlerFunc(handlers.GetPurchaseOrderByID(db)), cfg.JWTSecret)).Methods("GET")

	// Creación: Admin y Repositor (con auditoría)
	api.Handle("/purchase-orders",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin", "repositor")(handlersApp.CreatePurchaseOrderV2()),
			cfg.JWTSecret,
		)).Methods("POST")

	// Actualización de estado: Admin y Repositor (con auditoría)
	api.Handle("/purchase-orders/{id:[0-9]+}/status",
		middleware.JWTMiddleware(
			middleware.RequireRole("admin", "repositor")(handlersApp.UpdatePurchaseOrderStatusV2()),
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
