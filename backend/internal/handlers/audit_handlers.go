package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/models"
)

// GetAuditLogs maneja GET /api/v1/admin/audit-logs
// Devuelve los logs de auditoría con paginación (solo para admins)
func GetAuditLogs(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parsear parámetros de paginación
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := 50 // Por defecto 50 registros
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
				if limit > 200 {
					limit = 200 // Máximo 200 registros por página
				}
			}
		}

		offset := 0
		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Consultar los logs de auditoría
		query := `
			SELECT 
				id, 
				user_id, 
				user_email, 
				user_role, 
				action, 
				entity_type, 
				entity_id, 
				details, 
				timestamp
			FROM audit_logs
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Query(context.Background(), query, limit, offset)
		if err != nil {
			http.Error(w, "could not fetch audit logs", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		logs := make([]models.AuditLog, 0)
		for rows.Next() {
			var log models.AuditLog
			err := rows.Scan(
				&log.ID,
				&log.UserID,
				&log.UserEmail,
				&log.UserRole,
				&log.Action,
				&log.EntityType,
				&log.EntityID,
				&log.Details,
				&log.Timestamp,
			)
			if err != nil {
				http.Error(w, "error scanning audit log", http.StatusInternalServerError)
				return
			}
			logs = append(logs, log)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "error iterating audit logs", http.StatusInternalServerError)
			return
		}

		// Contar el total de registros para paginación
		var total int
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM audit_logs").Scan(&total)
		if err != nil {
			http.Error(w, "could not count audit logs", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"logs":   logs,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
	}
}
