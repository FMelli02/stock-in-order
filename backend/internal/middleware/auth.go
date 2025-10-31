package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

// Context keys for user data
type ctxKey string

const (
	userIDKey   ctxKey = "user_id"
	userRoleKey ctxKey = "user_role"
)

// JWTMiddleware validates a Bearer token and injects user_id into request context.
func JWTMiddleware(next http.Handler, jwtSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

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

		// Inject both user_id and role into context
		ctx := context.WithValue(r.Context(), userIDKey, uid)
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

// RequireRole is a middleware that restricts access based on user role.
// It must be used AFTER JWTMiddleware.
func RequireRole(requiredRole string) func(http.Handler) http.Handler {
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

			// Check if user has the required role
			if role != requiredRole {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "No tienes permisos de " + requiredRole + " para esta acción",
				})
				return
			}

			// User has the required role, proceed
			next.ServeHTTP(w, r)
		})
	}
}
