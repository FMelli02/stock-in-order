package config

import (
	"os"
)

// Config holds worker service configuration.
// DB_DSN: Postgres connection string.
// RabbitMQ_URL: AMQP connection string.
// SendGrid_API_Key: API key for sending emails (future use).
// EncryptionKey: Key for decrypting integration tokens.
type Config struct {
	DB_DSN          string
	RabbitMQ_URL    string
	SendGrid_APIKey string
	EncryptionKey   string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() Config {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://user:pass@localhost:5432/stock_db?sslmode=disable"
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://user:pass@localhost:5672/"
	}

	sendgridKey := os.Getenv("SENDGRID_API_KEY")
	// SendGrid es opcional por ahora

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		// Usar la misma clave que el backend (debe estar en .env)
		encryptionKey = "12345678901234567890123456789012" // 32 bytes - CAMBIAR EN PRODUCCIÓN
	}

	return Config{
		DB_DSN:          dsn,
		RabbitMQ_URL:    rabbitURL,
		SendGrid_APIKey: sendgridKey,
		EncryptionKey:   encryptionKey,
	}
}
