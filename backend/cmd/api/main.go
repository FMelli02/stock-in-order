package main

import (
	"log"
	"net/http"

	"stock-in-order/backend/internal/config"
	"stock-in-order/backend/internal/database"
	"stock-in-order/backend/internal/router"
)

func main() {
	log.Println("Iniciando servidor de Stock In Order...")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to PostgreSQL
	pool, err := database.Connect(cfg.DB_DSN)
	if err != nil {
		log.Fatalf("error conectando a la base de datos: %v", err)
	}
	defer pool.Close()

	// Initialize router with routes
	r := router.SetupRouter(pool, cfg.JWTSecret)

	// Start HTTP server
	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	log.Printf("Servidor HTTP escuchando en %s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error iniciando el servidor HTTP: %v", err)
	}
}
