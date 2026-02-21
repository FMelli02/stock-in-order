package config

import "os"

// Config holds scheduler service configuration
type Config struct {
	RabbitMQ_URL string
	DB_DSN       string
}

// LoadConfig reads configuration from environment variables
func LoadConfig() Config {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://user:pass@localhost:5672/"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://user:pass@localhost:5432/stock_db?sslmode=disable"
	}

	return Config{
		RabbitMQ_URL: rabbitURL,
		DB_DSN:       dbDSN,
	}
}
