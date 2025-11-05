package repository

import (
	"context"
	"log/slog"
	"stock-in-order/backend/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository maneja las operaciones de auditoría
type AuditRepository struct {
	db *pgxpool.Pool
}

// NewAuditRepository crea una nueva instancia del repositorio de auditoría
func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{
		db: db,
	}
}

// Log registra una acción en el libro de actas de forma asíncrona
// No retorna errores porque no debe bloquear la operación principal
func (r *AuditRepository) Log(logEntry *models.AuditLog) {
	// Ejecutar en goroutine para no bloquear la operación principal
	go func() {
		ctx := context.Background()

		// Si no se especificó timestamp, usar el actual
		if logEntry.Timestamp.IsZero() {
			logEntry.Timestamp = time.Now()
		}

		query := `
            INSERT INTO audit_logs (
                user_id, user_email, user_role, action, 
                entity_type, entity_id, details, timestamp
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `

		_, err := r.db.Exec(
			ctx,
			query,
			logEntry.UserID,
			logEntry.UserEmail,
			logEntry.UserRole,
			logEntry.Action,
			logEntry.EntityType,
			logEntry.EntityID,
			logEntry.Details,
			logEntry.Timestamp,
		)

		if err != nil {
			// Solo loguear el error, no propagar
			// La auditoría no debe afectar el flujo normal
			slog.Error("Failed to write audit log",
				"error", err,
				"user_email", logEntry.UserEmail,
				"action", logEntry.Action,
				"entity_type", logEntry.EntityType,
			)
		}
	}()
}
