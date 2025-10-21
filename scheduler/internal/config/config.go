package config

import "os"

// Config holds scheduler service configuration
type Config struct {
	RabbitMQ_URL string
}

// LoadConfig reads configuration from environment variables
func LoadConfig() Config {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://user:pass@localhost:5672/"
	}

	return Config{
		RabbitMQ_URL: rabbitURL,
	}
}
