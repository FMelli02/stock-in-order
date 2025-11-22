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
	"stock-in-order/backend/internal/mercadopago"
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
func SetupRouter(db *pgxpool.Pool, rabbit *rabbitmq.Client, auditRepo *repository.AuditRepository, mpClient *mercadopago.Client, cfg config.Config, logger *slog.Logger) http.Handler {
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

	// Initialize email service for password recovery
	emailService := services.NewEmailService()

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", handlers.Health()).Methods("GET")
	api.HandleFunc("/users/register", handlers.RegisterUser(db)).Methods("POST")
	api.HandleFunc("/users/login", handlers.LoginUser(db, cfg.JWTSecret)).Methods("POST")

	// Password recovery endpoints (public, no authentication required)
	api.HandleFunc("/users/forgot-password", handlers.ForgotPassword(db, emailService)).Methods("POST")
	api.HandleFunc("/users/reset-password", handlers.ResetPassword(db)).Methods("PUT")

	// Obtener usuario autenticado (protegido por JWT)
	api.Handle("/users/me",
		middleware.JWTMiddleware(
			http.HandlerFunc(handlers.GetCurrentUser(db)),
			cfg.JWTSecret,
		)).Methods("GET")

	// Helper function to combine JWT + Active Subscription middlewares (PAYWALL)
	withPaywall := func(handler http.Handler) http.Handler {
		return middleware.JWTMiddleware(
			middleware.RequireActiveSubscription(db)(handler),
			cfg.JWTSecret,
		)
	}

	// ============================================
	// ADMIN - Gestión de Usuarios con Auditoría
	// ============================================
	api.Handle("/admin/users",
		withPaywall(
			middleware.RequireRole("admin")(handlersApp.CreateUserByAdminV2()),
		),
	).Methods("POST")

	// RBAC Test endpoints (protected by JWT + Role middleware)
	api.Handle("/test/admin-only",
		withPaywall(
			middleware.RequireRole("admin")(http.HandlerFunc(handlers.AdminOnlyTest())),
		),
	).Methods("GET")

	api.Handle("/test/vendedor-only",
		withPaywall(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.VendedorOnlyTest())),
		),
	).Methods("GET")

	// ============================================
	// PRODUCTS - Con protección RBAC, Auditoría y PAYWALL
	// ============================================
	// Lectura: Todos los autenticados con suscripción activa
	api.Handle("/products",
		withPaywall(http.HandlerFunc(handlers.ListProducts(db)))).Methods("GET")
	api.Handle("/products/{id:[0-9]+}",
		withPaywall(http.HandlerFunc(handlers.GetProduct(db)))).Methods("GET")
	api.Handle("/products/{id:[0-9]+}/batches",
		withPaywall(http.HandlerFunc(handlers.GetProductBatches(db)))).Methods("GET")
	api.Handle("/products/{id:[0-9]+}/movements",
		withPaywall(http.HandlerFunc(handlers.GetProductMovements(db)))).Methods("GET")

	// Creación: Admin y Repositor (con auditoría y paywall)
	api.Handle("/products",
		withPaywall(
			middleware.RequireRole("admin", "repositor")(handlersApp.CreateProductV2()),
		)).Methods("POST")

	// Actualización: Admin y Repositor (con auditoría y paywall)
	api.Handle("/products/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("admin", "repositor")(handlersApp.UpdateProductV2()),
		)).Methods("PUT")

	// Ajuste de Stock: Admin y Repositor (con auditoría y paywall)
	api.Handle("/products/{id:[0-9]+}/adjust-stock",
		withPaywall(
			middleware.RequireRole("admin", "repositor")(handlersApp.AdjustProductStockV2()),
		)).Methods("POST")

	// Eliminación: Solo Admin (con auditoría y paywall)
	api.Handle("/products/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("admin")(handlersApp.DeleteProductV2()),
		)).Methods("DELETE")

	// ============================================
	// DASHBOARD - Con PAYWALL (requiere suscripción activa)
	// ============================================
	api.Handle("/dashboard/metrics",
		withPaywall(http.HandlerFunc(handlers.GetDashboardMetrics(db)))).Methods("GET")
	api.Handle("/dashboard/kpis",
		withPaywall(http.HandlerFunc(handlers.GetDashboardKPIs(db)))).Methods("GET")
	api.Handle("/dashboard/charts",
		withPaywall(http.HandlerFunc(handlers.GetDashboardCharts(db)))).Methods("GET")

	// ============================================
	// REPORTS - Con PAYWALL (requiere suscripción activa)
	// ============================================
	api.Handle("/reports/products/email",
		withPaywall(http.HandlerFunc(handlers.RequestProductsReportByEmail(db, rabbit)))).Methods("POST")
	api.Handle("/reports/customers/email",
		withPaywall(http.HandlerFunc(handlers.RequestCustomersReportByEmail(db, rabbit)))).Methods("POST")
	api.Handle("/reports/suppliers/email",
		withPaywall(http.HandlerFunc(handlers.RequestSuppliersReportByEmail(db, rabbit)))).Methods("POST")

	api.Handle("/reports/products/xlsx",
		withPaywall(http.HandlerFunc(handlers.ExportProductsXLSX(db)))).Methods("GET")
	api.Handle("/reports/customers/xlsx",
		withPaywall(http.HandlerFunc(handlers.ExportCustomersXLSX(db)))).Methods("GET")
	api.Handle("/reports/suppliers/xlsx",
		withPaywall(http.HandlerFunc(handlers.ExportSuppliersXLSX(db)))).Methods("GET")
	api.Handle("/reports/sales-orders/xlsx",
		withPaywall(http.HandlerFunc(handlers.ExportSalesOrdersXLSX(db)))).Methods("GET")
	api.Handle("/reports/purchase-orders/xlsx",
		withPaywall(http.HandlerFunc(handlers.ExportPurchaseOrdersXLSX(db)))).Methods("GET")

	// ============================================
	// ADMIN - Registro de Auditoría con PAYWALL
	// ============================================
	api.Handle("/admin/audit-logs",
		withPaywall(
			middleware.RequireRole("admin")(handlers.GetAuditLogs(db)),
		)).Methods("GET")

	// ============================================
	// SUPPLIERS - Con protección RBAC, Auditoría y PAYWALL
	// ============================================
	// Lectura: Todos los autenticados con suscripción activa
	api.Handle("/suppliers",
		withPaywall(http.HandlerFunc(handlers.ListSuppliers(db)))).Methods("GET")
	api.Handle("/suppliers/{id:[0-9]+}",
		withPaywall(http.HandlerFunc(handlers.GetSupplier(db)))).Methods("GET")

	// Creación: Admin y Repositor (con auditoría y paywall)
	api.Handle("/suppliers",
		withPaywall(
			middleware.RequireRole("admin", "repositor")(handlersApp.CreateSupplierV2()),
		)).Methods("POST")

	// Actualización: Admin y Repositor (con auditoría y paywall)
	api.Handle("/suppliers/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("admin", "repositor")(handlersApp.UpdateSupplierV2()),
		)).Methods("PUT")

	// Eliminación: Solo Admin (con auditoría y paywall)
	api.Handle("/suppliers/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("admin")(handlersApp.DeleteSupplierV2()),
		)).Methods("DELETE")

	// ============================================
	// CUSTOMERS - Con protección RBAC, Auditoría y PAYWALL
	// ============================================
	// Lectura: Admin y Vendedor con suscripción activa (repositor NO puede ver clientes)
	api.Handle("/customers",
		withPaywall(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.ListCustomers(db))),
		)).Methods("GET")
	api.Handle("/customers/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetCustomer(db))),
		)).Methods("GET")

	// Creación: Admin y Vendedor (con auditoría y paywall)
	api.Handle("/customers",
		withPaywall(
			middleware.RequireRole("vendedor")(handlersApp.CreateCustomerV2()),
		)).Methods("POST")

	// Actualización: Admin y Vendedor (con auditoría y paywall)
	api.Handle("/customers/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("vendedor")(handlersApp.UpdateCustomerV2()),
		)).Methods("PUT")

	// Eliminación: Solo Admin (con auditoría y paywall)
	api.Handle("/customers/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("admin")(handlersApp.DeleteCustomerV2()),
		)).Methods("DELETE")

	// ============================================
	// SALES ORDERS - Con protección RBAC, Auditoría y PAYWALL
	// ============================================
	// Creación y Lectura: Admin y Vendedor con suscripción activa
	api.Handle("/sales-orders",
		withPaywall(
			middleware.RequireRole("vendedor")(handlersApp.CreateSalesOrderV2()),
		)).Methods("POST")
	api.Handle("/sales-orders",
		withPaywall(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetSalesOrders(db))),
		)).Methods("GET")
	api.Handle("/sales-orders/{id:[0-9]+}",
		withPaywall(
			middleware.RequireRole("vendedor")(http.HandlerFunc(handlers.GetSalesOrderByID(db))),
		)).Methods("GET")

	// ============================================
	// PURCHASE ORDERS - Con protección RBAC, Auditoría y PAYWALL
	// ============================================
	// Lectura: Todos los autenticados con suscripción activa
	api.Handle("/purchase-orders",
		withPaywall(http.HandlerFunc(handlers.GetPurchaseOrders(db)))).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}",
		withPaywall(http.HandlerFunc(handlers.GetPurchaseOrderByID(db)))).Methods("GET")

	// Creación: Admin y Repositor (con auditoría y paywall)
	api.Handle("/purchase-orders",
		withPaywall(
			middleware.RequireRole("admin", "repositor")(handlersApp.CreatePurchaseOrderV2()),
		)).Methods("POST")

	// Actualización de estado: Admin y Repositor (con auditoría y paywall)
	api.Handle("/purchase-orders/{id:[0-9]+}/status",
		withPaywall(
			middleware.RequireRole("admin", "repositor")(handlersApp.UpdatePurchaseOrderStatusV2()),
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

	// Listar integraciones del usuario (protegido con paywall)
	api.Handle("/integrations",
		withPaywall(
			http.HandlerFunc(integrationHandlers.HandleListIntegrations),
		)).Methods("GET")

	// Eliminar integración (protegido con paywall)
	api.Handle("/integrations/{platform}",
		withPaywall(
			http.HandlerFunc(integrationHandlers.HandleDeleteIntegration),
		)).Methods("DELETE")

	// OAuth2 - Iniciar conexión con Mercado Libre (protegido con paywall)
	api.Handle("/integrations/mercadolibre/connect",
		withPaywall(
			http.HandlerFunc(integrationHandlers.HandleMercadoLibreConnect),
		)).Methods("GET")

	// OAuth2 - Callback de Mercado Libre (público, no requiere JWT)
	api.HandleFunc("/integrations/mercadolibre/callback",
		integrationHandlers.HandleMercadoLibreCallback).Methods("GET")

	// ============================================
	// SUBSCRIPTIONS - Gestión de suscripciones y pagos
	// ============================================
	// Obtener estado de suscripción actual (todos los autenticados)
	api.Handle("/subscriptions/status",
		middleware.JWTMiddleware(
			http.HandlerFunc(handlers.GetSubscriptionStatusHandler(db)),
			cfg.JWTSecret,
		)).Methods("GET")

	// Crear checkout de pago único (todos los autenticados)
	api.Handle("/subscriptions/create-checkout",
		middleware.JWTMiddleware(
			http.HandlerFunc(handlers.CreateCheckoutHandler(db, mpClient)),
			cfg.JWTSecret,
		)).Methods("POST")

	// Crear suscripción recurrente (todos los autenticados)
	api.Handle("/subscriptions/create-recurring",
		middleware.JWTMiddleware(
			http.HandlerFunc(handlers.CreateRecurringSubscriptionHandler(db, mpClient)),
			cfg.JWTSecret,
		)).Methods("POST")

	// Cancelar suscripción (todos los autenticados)
	api.Handle("/subscriptions/cancel",
		middleware.JWTMiddleware(
			http.HandlerFunc(handlers.CancelSubscriptionHandler(db, mpClient)),
			cfg.JWTSecret,
		)).Methods("POST")

	// ============================================
	// WEBHOOKS - Notificaciones de plataformas externas
	// ============================================
	webhookHandlers := handlers.NewMercadoLibreWebhookHandlers(rabbit)

	// Webhook de Mercado Libre (público, llamado por Meli)
	api.HandleFunc("/webhooks/mercadolibre",
		webhookHandlers.HandleMercadoLibreWebhook).Methods("POST")

	// Webhook de MercadoPago para pagos y suscripciones (público, llamado por MP)
	api.HandleFunc("/webhooks/mercadopago",
		handlers.HandleMercadoPagoWebhook(db, mpClient)).Methods("POST")

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
