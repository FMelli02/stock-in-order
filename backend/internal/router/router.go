package router

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	"stock-in-order/backend/internal/handlers"
	"stock-in-order/backend/internal/middleware"
)

// SetupRouter wires up HTTP routes.
func SetupRouter(db *pgxpool.Pool, jwtSecret string, logger *slog.Logger) http.Handler {
	r := mux.NewRouter()

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", handlers.Health()).Methods("GET")
	api.HandleFunc("/users/register", handlers.RegisterUser(db)).Methods("POST")
	api.HandleFunc("/users/login", handlers.LoginUser(db, jwtSecret)).Methods("POST")

	// Products endpoints (protected by JWT middleware)
	api.Handle("/products", middleware.JWTMiddleware(http.HandlerFunc(handlers.ListProducts(db)), jwtSecret)).Methods("GET")
	api.Handle("/products", middleware.JWTMiddleware(http.HandlerFunc(handlers.CreateProduct(db)), jwtSecret)).Methods("POST")
	api.Handle("/products/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProduct(db)), jwtSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.UpdateProduct(db)), jwtSecret)).Methods("PUT")
	api.Handle("/products/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.DeleteProduct(db)), jwtSecret)).Methods("DELETE")
	api.Handle("/products/{id:[0-9]+}/movements", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetProductMovements(db)), jwtSecret)).Methods("GET")
	api.Handle("/products/{id:[0-9]+}/adjust-stock", middleware.JWTMiddleware(http.HandlerFunc(handlers.AdjustProductStock(db)), jwtSecret)).Methods("POST")

	// Dashboard metrics (protected)
	api.Handle("/dashboard/metrics", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardMetrics(db)), jwtSecret)).Methods("GET")
	api.Handle("/dashboard/kpis", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardKPIs(db)), jwtSecret)).Methods("GET")
	api.Handle("/dashboard/charts", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetDashboardCharts(db)), jwtSecret)).Methods("GET")

	// Reports endpoints (protected)
	api.Handle("/reports/products/csv", middleware.JWTMiddleware(http.HandlerFunc(handlers.ExportProductsCSV(db)), jwtSecret)).Methods("GET")

	// Suppliers endpoints (protected by JWT middleware)
	api.Handle("/suppliers", middleware.JWTMiddleware(http.HandlerFunc(handlers.ListSuppliers(db)), jwtSecret)).Methods("GET")
	api.Handle("/suppliers", middleware.JWTMiddleware(http.HandlerFunc(handlers.CreateSupplier(db)), jwtSecret)).Methods("POST")
	api.Handle("/suppliers/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetSupplier(db)), jwtSecret)).Methods("GET")
	api.Handle("/suppliers/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.UpdateSupplier(db)), jwtSecret)).Methods("PUT")
	api.Handle("/suppliers/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.DeleteSupplier(db)), jwtSecret)).Methods("DELETE")

	// Customers endpoints (protected by JWT middleware)
	api.Handle("/customers", middleware.JWTMiddleware(http.HandlerFunc(handlers.ListCustomers(db)), jwtSecret)).Methods("GET")
	api.Handle("/customers", middleware.JWTMiddleware(http.HandlerFunc(handlers.CreateCustomer(db)), jwtSecret)).Methods("POST")
	api.Handle("/customers/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetCustomer(db)), jwtSecret)).Methods("GET")
	api.Handle("/customers/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.UpdateCustomer(db)), jwtSecret)).Methods("PUT")
	api.Handle("/customers/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.DeleteCustomer(db)), jwtSecret)).Methods("DELETE")

	// Sales orders (protected by JWT middleware)
	api.Handle("/sales-orders", middleware.JWTMiddleware(http.HandlerFunc(handlers.CreateSalesOrder(db)), jwtSecret)).Methods("POST")
	api.Handle("/sales-orders", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetSalesOrders(db)), jwtSecret)).Methods("GET")
	api.Handle("/sales-orders/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetSalesOrderByID(db)), jwtSecret)).Methods("GET")

	// Purchase orders (protected by JWT middleware)
	api.Handle("/purchase-orders", middleware.JWTMiddleware(http.HandlerFunc(handlers.CreatePurchaseOrder(db)), jwtSecret)).Methods("POST")
	api.Handle("/purchase-orders", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetPurchaseOrders(db)), jwtSecret)).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}", middleware.JWTMiddleware(http.HandlerFunc(handlers.GetPurchaseOrderByID(db)), jwtSecret)).Methods("GET")
	api.Handle("/purchase-orders/{id:[0-9]+}/status", middleware.JWTMiddleware(http.HandlerFunc(handlers.UpdatePurchaseOrderStatus(db)), jwtSecret)).Methods("PUT")

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
