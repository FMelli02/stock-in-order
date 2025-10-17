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
	Port      string
	DB_DSN    string
	JWTSecret string
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

	return Config{
		Port:      port,
		DB_DSN:    dsn,
		JWTSecret: jwt,
	}
}
