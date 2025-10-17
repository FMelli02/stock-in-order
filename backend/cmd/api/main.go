package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"stock-in-order/backend/internal/config"
	"stock-in-order/backend/internal/database"
	"stock-in-order/backend/internal/router"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present (development), ignore error if not found (production)
	_ = godotenv.Load()

	// Initialize structured logger (JSON format)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add custom formatting if needed
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	logger.Info("Iniciando servidor de Stock In Order...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Sentry for error tracking
	sentryDSN := os.Getenv("SENTRY_DSN")
	if sentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDSN,
			Environment:      getEnvironment(),
			Release:          "stock-in-order@1.0.0",
			TracesSampleRate: 1.0,
			AttachStacktrace: true,
		})
		if err != nil {
			logger.Error("Error inicializando Sentry", "error", err)
		} else {
			logger.Info("Sentry inicializado correctamente", "environment", getEnvironment())
			defer sentry.Flush(2 * time.Second)
		}
	} else {
		logger.Warn("SENTRY_DSN no configurado, monitoreo de errores deshabilitado")
	}

	// Connect to PostgreSQL
	pool, err := database.Connect(cfg.DB_DSN)
	if err != nil {
		logger.Error("Error conectando a la base de datos", "error", err)
		sentry.CaptureException(err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("Conexión a base de datos establecida")

	// Initialize router with routes
	r := router.SetupRouter(pool, cfg.JWTSecret, logger)

	// Start HTTP server
	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	logger.Info("Servidor HTTP iniciado", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Error iniciando el servidor HTTP", "error", err)
		sentry.CaptureException(err)
		os.Exit(1)
	}
}

// getEnvironment returns the current environment (development, staging, production)
func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}
