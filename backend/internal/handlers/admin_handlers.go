package handlers

import (
	"encoding/json"
	"net/http"

	"stock-in-order/backend/internal/middleware"
)

// AdminOnlyTest is a test endpoint that requires admin role.
func AdminOnlyTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user info from context
		userID, _ := middleware.UserIDFromContext(r.Context())
		role, _ := middleware.UserRoleFromContext(r.Context())

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "¡Bienvenido admin! Solo usuarios con rol 'admin' pueden ver esto.",
			"user_id": userID,
			"role":    role,
		})
	}
}

// VendedorOnlyTest is a test endpoint that requires vendedor role.
func VendedorOnlyTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user info from context
		userID, _ := middleware.UserIDFromContext(r.Context())
		role, _ := middleware.UserRoleFromContext(r.Context())

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "¡Bienvenido vendedor! Solo usuarios con rol 'vendedor' pueden ver esto.",
			"user_id": userID,
			"role":    role,
		})
	}
}
