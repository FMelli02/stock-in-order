package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/worker/internal/config"
	"stock-in-order/worker/internal/consumer"
	"stock-in-order/worker/internal/email"
)

func main() {
	log.Println("🚀 Iniciando Worker Service...")

	// Cargar configuración
	cfg := config.LoadConfig()
	log.Printf("📝 Configuración cargada: DB_DSN=%s, RabbitMQ_URL=%s",
		maskConnectionString(cfg.DB_DSN),
		maskConnectionString(cfg.RabbitMQ_URL))

	// Conectar a PostgreSQL
	dbpool, err := pgxpool.New(context.Background(), cfg.DB_DSN)
	if err != nil {
		log.Fatalf("❌ No se pudo conectar a la base de datos: %v", err)
	}
	defer dbpool.Close()

	// Verificar conexión a la base de datos
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbpool.Ping(ctx); err != nil {
		log.Fatalf("❌ No se pudo hacer ping a la base de datos: %v", err)
	}
	log.Println("✅ Conectado a PostgreSQL")

	// Configurar cliente de SendGrid
	emailClient := email.NewClient(
		cfg.SendGrid_APIKey,
		"francoleproso1@gmail.com", // Email remitente
		"Stock in Order",           // Nombre remitente
	)
	log.Println("📧 Cliente de email configurado")

	// Canal para manejar señales de sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Iniciar consumidor en una goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := consumer.StartConsumer(cfg.RabbitMQ_URL, dbpool, emailClient, cfg.EncryptionKey); err != nil {
			errChan <- err
		}
	}()

	// Esperar por señal de terminación o error
	select {
	case sig := <-sigChan:
		log.Printf("🛑 Señal recibida: %v. Cerrando worker...", sig)
	case err := <-errChan:
		log.Printf("❌ Error fatal en consumer: %v", err)
	}

	log.Println("👋 Worker Service finalizado")
}

// maskConnectionString oculta credenciales en los logs
func maskConnectionString(s string) string {
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return "***"
}
