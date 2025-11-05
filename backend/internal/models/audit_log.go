package models

import "time"

// AuditLog representa una entrada en el libro de actas
type AuditLog struct {
ID         int64     `json:"id"`
UserID     *int64    `json:"user_id"`      // Nullable - puede ser una acción del sistema
UserEmail  string    `json:"user_email"`   // Email del usuario que realizó la acción
UserRole   string    `json:"user_role"`    // Rol del usuario en el momento de la acción
Action     string    `json:"action"`       // Tipo de acción (CREATE, UPDATE, DELETE, etc.)
EntityType string    `json:"entity_type"`  // Tipo de entidad afectada (PRODUCT, CUSTOMER, etc.)
EntityID   *int64    `json:"entity_id"`    // ID de la entidad afectada (nullable para acciones generales)
Details    string    `json:"details"`      // Detalles adicionales en formato JSON
Timestamp  time.Time `json:"timestamp"`    // Momento en que ocurrió la acción
}

// Constantes para tipos de acciones
const (
ActionCreate = "CREATE"
ActionUpdate = "UPDATE"
ActionDelete = "DELETE"
ActionLogin  = "LOGIN"
ActionLogout = "LOGOUT"
ActionExport = "EXPORT"
ActionImport = "IMPORT"
ActionSync   = "SYNC"
)

// Constantes para tipos de entidades
const (
EntityTypeProduct     = "PRODUCT"
EntityTypeCustomer    = "CUSTOMER"
EntityTypeOrder       = "ORDER"
EntityTypeUser        = "USER"
EntityTypeIntegration = "INTEGRATION"
EntityTypeStock       = "STOCK"
EntityTypePrice       = "PRICE"
EntityTypeCategory    = "CATEGORY"
)
