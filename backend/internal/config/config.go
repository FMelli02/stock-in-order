package config

import (
	"os"
)

// Config holds application configuration.
// Port: HTTP server address (e.g., ":8080").
// DB_DSN: Postgres connection string.
// Environment variables used: PORT, DB_DSN.
// Defaults: PORT=":8080", DB_DSN="postgres://user:pass@localhost:5432/stock_db?sslmode=disable"

type Config struct {
	Port          string
	DB_DSN        string
	JWTSecret     string
	EncryptionKey string

	// Mercado Libre OAuth2 Configuration
	MLClientID     string
	MLClientSecret string
	MLRedirectURI  string
}

// LoadConfig reads configuration from environment variables with sensible defaults.
func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://user:pass@localhost:5432/stock_db?sslmode=disable"
	}

	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		jwt = "dev-jwt-secret-change-me"
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		// Clave por defecto para desarrollo (32 bytes)
		// En producción, esta debe ser una clave aleatoria de 32 bytes
		encryptionKey = "dev-encryption-key-change-me32"
	}

	mlClientID := os.Getenv("ML_CLIENT_ID")
	if mlClientID == "" {
		mlClientID = "your_mercadolibre_app_id"
	}

	mlClientSecret := os.Getenv("ML_CLIENT_SECRET")
	if mlClientSecret == "" {
		mlClientSecret = "your_mercadolibre_app_secret"
	}

	mlRedirectURI := os.Getenv("ML_REDIRECT_URI")
	if mlRedirectURI == "" {
		mlRedirectURI = "http://localhost:8080/api/v1/integrations/mercadolibre/callback"
	}

	return Config{
		Port:          port,
		DB_DSN:        dsn,
		JWTSecret:     jwt,
		EncryptionKey: encryptionKey,

		MLClientID:     mlClientID,
		MLClientSecret: mlClientSecret,
		MLRedirectURI:  mlRedirectURI,
	}
}
