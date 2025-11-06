package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/models"
)

// Context keys for user data
type ctxKey string

const (
	userIDKey    ctxKey = "user_id"
	userEmailKey ctxKey = "user_email"
	userRoleKey  ctxKey = "user_role"
)

// JWTMiddleware validates a Bearer token and injects user_id into request context.
// It accepts the token from Authorization header or from 'token' query parameter.
func JWTMiddleware(next http.Handler, jwtSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string

		// Try to get token from Authorization header first
		auth := r.Header.Get("Authorization")
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		} else {
			// If not in header, try to get from query parameter
			tokenStr = r.URL.Query().Get("token")
		}

		if tokenStr == "" {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		uidVal, ok := claims["user_id"]
		if !ok {
			http.Error(w, "user_id missing in token", http.StatusUnauthorized)
			return
		}

		// Accept numeric user_id (float64) from JSON numeric claims
		var uid int64
		switch v := uidVal.(type) {
		case float64:
			uid = int64(v)
		case int64:
			uid = v
		case json.Number:
			parsed, _ := v.Int64()
			uid = parsed
		default:
			http.Error(w, "invalid user_id type", http.StatusUnauthorized)
			return
		}

		// Extract role from token claims
		roleVal, _ := claims["role"]
		role, _ := roleVal.(string)

		// Extract email from token claims
		emailVal, _ := claims["email"]
		email, _ := emailVal.(string)

		// Inject user_id, email, and role into context
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		ctx = context.WithValue(ctx, userEmailKey, email)
		ctx = context.WithValue(ctx, userRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext retrieves the user ID stored by JWTMiddleware.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0, false
	}
	uid, ok := v.(int64)
	return uid, ok
}

// UserRoleFromContext retrieves the user role stored by JWTMiddleware.
func UserRoleFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userRoleKey)
	if v == nil {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

// UserEmailFromContext retrieves the user email stored by JWTMiddleware.
func UserEmailFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userEmailKey)
	if v == nil {
		return "", false
	}
	email, ok := v.(string)
	return email, ok
}

// RequireRole is a middleware that restricts access based on user role.
// It must be used AFTER JWTMiddleware.
// Accepts one or more roles. User needs to have ANY of the provided roles.
func RequireRole(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract role from context (injected by JWTMiddleware)
			role, ok := UserRoleFromContext(r.Context())
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "No se pudo determinar el rol del usuario",
				})
				return
			}

			// Check if user has ANY of the required roles
			hasPermission := false
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "Acceso denegado: requiere rol " + joinRoles(requiredRoles),
				})
				return
			}

			// User has the required role, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// joinRoles helper function to format roles for error message
func joinRoles(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	if len(roles) == 1 {
		return roles[0]
	}
	if len(roles) == 2 {
		return roles[0] + " o " + roles[1]
	}
	result := ""
	for i, role := range roles {
		if i == len(roles)-1 {
			result += " o " + role
		} else if i > 0 {
			result += ", " + role
		} else {
			result = role
		}
	}
	return result
}

// ============================================
// PAYWALL MIDDLEWARE - "EL PATOVICA 2.0"
// ============================================

// RequireActiveSubscription middleware verifica que el usuario tenga una suscripción activa.
// Debe usarse DESPUÉS de JWTMiddleware (requiere user_id en contexto).
// Si la suscripción no está activa, responde con 402 Payment Required.
func RequireActiveSubscription(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extraer user_id del contexto (inyectado por JWTMiddleware)
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "Usuario no autenticado",
				})
				return
			}

			// Obtener suscripción del usuario
			sm := &models.SubscriptionModel{DB: db}
			subscription, err := sm.GetByUserID(userID)
			if err != nil {
				if err == models.ErrNotFound {
					// Usuario sin suscripción → crear una gratuita automáticamente
					slog.Warn("Usuario sin suscripción, se requiere crear una", "userID", userID)
					w.WriteHeader(http.StatusPaymentRequired)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"error":       "No tienes una suscripción activa",
						"message":     "Para continuar usando la aplicación, necesitas una suscripción activa.",
						"action":      "upgrade",
						"upgrade_url": "/subscriptions/status",
					})
					return
				}

				// Error al consultar DB
				slog.Error("Error obteniendo suscripción", "error", err, "userID", userID)
				http.Error(w, "Error verificando suscripción", http.StatusInternalServerError)
				return
			}

			// Verificar que la suscripción esté activa
			if subscription.Status != models.SubscriptionStatusActive {
				slog.Info("Acceso denegado: suscripción no activa",
					"userID", userID,
					"status", subscription.Status,
					"plan", subscription.PlanID)

				w.WriteHeader(http.StatusPaymentRequired)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":       "Tu suscripción no está activa",
					"message":     "Tu suscripción está " + string(subscription.Status) + ". Para continuar, reactiva tu suscripción.",
					"status":      subscription.Status,
					"plan":        subscription.PlanID,
					"action":      "reactivate",
					"upgrade_url": "/subscriptions/status",
				})
				return
			}

			// Suscripción activa → permitir acceso
			slog.Debug("Suscripción activa verificada",
				"userID", userID,
				"plan", subscription.PlanID)

			// Opcional: Inyectar plan_id en el contexto para uso posterior
			ctx := context.WithValue(r.Context(), ctxKey("subscription"), subscription)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SubscriptionFromContext retrieves the subscription stored by RequireActiveSubscription
func SubscriptionFromContext(ctx context.Context) (*models.Subscription, bool) {
	v := ctx.Value(ctxKey("subscription"))
	if v == nil {
		return nil, false
	}
	sub, ok := v.(*models.Subscription)
	return sub, ok
}
