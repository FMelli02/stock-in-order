package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/models"
)

// DTO de salida
// Se utiliza el mismo struct que en el modelo para simplicidad

// GetDashboardMetrics maneja GET /api/v1/dashboard/metrics
func GetDashboardMetrics(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		dm := &models.DashboardModel{DB: db}
		metrics, err := dm.GetMetrics(userID)
		if err != nil {
			http.Error(w, "could not fetch metrics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metrics)
	}
}
