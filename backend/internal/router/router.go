package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/handlers"
	"stock-in-order/backend/internal/middleware"
)

// SetupRouter wires up HTTP routes.
func SetupRouter(db *pgxpool.Pool, jwtSecret string) *mux.Router {
	r := mux.NewRouter()
	// Global middlewares
	r.Use(middleware.CORSMiddleware)

	// API v1
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", handlers.Health()).Methods("GET")
	api.HandleFunc("/users/register", handlers.RegisterUser(db)).Methods("POST")
	api.HandleFunc("/users/login", handlers.LoginUser(db, jwtSecret)).Methods("POST")

	// Products subrouter protected by JWT middleware
	products := api.PathPrefix("/products").Subrouter()
	products.Use(func(next http.Handler) http.Handler { return middleware.JWTMiddleware(next, jwtSecret) })
	products.HandleFunc("/", handlers.CreateProduct(db)).Methods("POST")
	products.HandleFunc("/", handlers.ListProducts(db)).Methods("GET")
	products.HandleFunc("/{id:[0-9]+}", handlers.GetProduct(db)).Methods("GET")
	products.HandleFunc("/{id:[0-9]+}", handlers.UpdateProduct(db)).Methods("PUT")
	products.HandleFunc("/{id:[0-9]+}", handlers.DeleteProduct(db)).Methods("DELETE")
	products.HandleFunc("/{id:[0-9]+}/movements", handlers.GetProductMovements(db)).Methods("GET")
	products.HandleFunc("/{id:[0-9]+}/adjust-stock", handlers.AdjustProductStock(db)).Methods("POST")

	// Dashboard metrics (protected)
	dashboard := api.PathPrefix("/dashboard").Subrouter()
	dashboard.Use(func(next http.Handler) http.Handler { return middleware.JWTMiddleware(next, jwtSecret) })
	dashboard.HandleFunc("/metrics", handlers.GetDashboardMetrics(db)).Methods("GET")

	// Suppliers subrouter protected by JWT middleware
	suppliers := api.PathPrefix("/suppliers").Subrouter()
	suppliers.Use(func(next http.Handler) http.Handler { return middleware.JWTMiddleware(next, jwtSecret) })
	suppliers.HandleFunc("/", handlers.CreateSupplier(db)).Methods("POST")
	suppliers.HandleFunc("/", handlers.ListSuppliers(db)).Methods("GET")
	suppliers.HandleFunc("/{id:[0-9]+}", handlers.GetSupplier(db)).Methods("GET")
	suppliers.HandleFunc("/{id:[0-9]+}", handlers.UpdateSupplier(db)).Methods("PUT")
	suppliers.HandleFunc("/{id:[0-9]+}", handlers.DeleteSupplier(db)).Methods("DELETE")

	// Customers subrouter protected by JWT middleware
	customers := api.PathPrefix("/customers").Subrouter()
	customers.Use(func(next http.Handler) http.Handler { return middleware.JWTMiddleware(next, jwtSecret) })
	customers.HandleFunc("/", handlers.CreateCustomer(db)).Methods("POST")
	customers.HandleFunc("/", handlers.ListCustomers(db)).Methods("GET")
	customers.HandleFunc("/{id:[0-9]+}", handlers.GetCustomer(db)).Methods("GET")
	customers.HandleFunc("/{id:[0-9]+}", handlers.UpdateCustomer(db)).Methods("PUT")
	customers.HandleFunc("/{id:[0-9]+}", handlers.DeleteCustomer(db)).Methods("DELETE")

	// Sales orders
	sales := api.PathPrefix("/sales-orders").Subrouter()
	sales.Use(func(next http.Handler) http.Handler { return middleware.JWTMiddleware(next, jwtSecret) })
	sales.HandleFunc("/", handlers.CreateSalesOrder(db)).Methods("POST")
	sales.HandleFunc("/", handlers.GetSalesOrders(db)).Methods("GET")
	sales.HandleFunc("/{id:[0-9]+}", handlers.GetSalesOrderByID(db)).Methods("GET")

	// Purchase orders
	purchase := api.PathPrefix("/purchase-orders").Subrouter()
	purchase.Use(func(next http.Handler) http.Handler { return middleware.JWTMiddleware(next, jwtSecret) })
	purchase.HandleFunc("/", handlers.CreatePurchaseOrder(db)).Methods("POST")
	purchase.HandleFunc("/", handlers.GetPurchaseOrders(db)).Methods("GET")
	purchase.HandleFunc("/{id:[0-9]+}", handlers.GetPurchaseOrderByID(db)).Methods("GET")
	purchase.HandleFunc("/{id:[0-9]+}/status", handlers.UpdatePurchaseOrderStatus(db)).Methods("PUT")

	return r
}
