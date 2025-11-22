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
	userIDKey         ctxKey = "user_id"
	userEmailKey      ctxKey = "user_email"
	userRoleKey       ctxKey = "user_role"
	organizationIDKey ctxKey = "organization_id"
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

		// Extract organization_id from token claims
		orgIDVal, _ := claims["organization_id"]
		var orgID int64
		switch v := orgIDVal.(type) {
		case float64:
			orgID = int64(v)
		case int64:
			orgID = v
		case json.Number:
			parsed, _ := v.Int64()
			orgID = parsed
		default:
			// Si no hay organization_id en el token, usar el user_id (para admins)
			orgID = uid
			slog.Warn("No organization_id in JWT token, using user_id as fallback", "user_id", uid, "email", email)
		}

		slog.Info("JWT Middleware", "user_id", uid, "email", email, "role", role, "organization_id", orgID, "path", r.URL.Path)

		// Inject user_id, email, role, and organization_id into context
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		ctx = context.WithValue(ctx, userEmailKey, email)
		ctx = context.WithValue(ctx, userRoleKey, role)
		ctx = context.WithValue(ctx, organizationIDKey, orgID)
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

// OrganizationIDFromContext retrieves the organization ID stored by JWTMiddleware.
func OrganizationIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(organizationIDKey)
	if v == nil {
		return 0, false
	}
	orgID, ok := v.(int64)
	return orgID, ok
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
					// Usuario sin suscripción → crear una GRATUITA automáticamente
					slog.Warn("Usuario sin suscripción, creando plan FREE automáticamente", "userID", userID)

					// Crear suscripción FREE (no expira)
					newSubscription := &models.Subscription{
						UserID:             userID,
						PlanID:             models.PlanFree,
						Status:             models.SubscriptionStatusActive,
						CurrentPeriodStart: nil, // FREE no tiene periodos
						CurrentPeriodEnd:   nil,
					}

					if err := sm.Create(newSubscription); err != nil {
						slog.Error("Error creando suscripción FREE", "error", err, "userID", userID)
						http.Error(w, "Error inicializando cuenta", http.StatusInternalServerError)
						return
					}

					// Usar la nueva suscripción
					subscription = newSubscription
					slog.Info("Suscripción FREE creada exitosamente", "userID", userID)
				} else {
					// Error al consultar DB
					slog.Error("Error obteniendo suscripción", "error", err, "userID", userID)
					http.Error(w, "Error verificando suscripción", http.StatusInternalServerError)
					return
				}
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
