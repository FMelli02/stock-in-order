package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryMiddleware wraps handlers to capture panics and errors
func SentryMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new Sentry hub for this request
		hub := sentry.CurrentHub().Clone()
		ctx := sentry.SetHubOnContext(r.Context(), hub)

		// Add request context to Sentry
		hub.Scope().SetRequest(r)
		hub.Scope().SetTag("method", r.Method)
		hub.Scope().SetTag("path", r.URL.Path)

		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recuperado",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)

				// Report to Sentry
				hub.RecoverWithContext(ctx, err)
				hub.Flush(2 * time.Second)

				// Return 500 error to client
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
