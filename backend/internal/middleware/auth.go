package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

// Context key for user ID
type ctxKey string

const userIDKey ctxKey = "user_id"

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

		ctx := context.WithValue(r.Context(), userIDKey, uid)
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
