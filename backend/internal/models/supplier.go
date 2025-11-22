package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Supplier represents a supplier belonging to a user.
type Supplier struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	ContactPerson string    `json:"contact_person,omitempty"`
	Email         string    `json:"email,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	Address       string    `json:"address,omitempty"`
	UserID        int64     `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// SupplierModel wraps DB access for suppliers.
type SupplierModel struct {
	DB *pgxpool.Pool
}

// Insert creates a new supplier for a user and sets ID and CreatedAt.
func (m *SupplierModel) Insert(s *Supplier) error {
	const q = `
		INSERT INTO suppliers (name, contact_person, email, phone, address, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return m.DB.QueryRow(context.Background(), q, s.Name, s.ContactPerson, s.Email, s.Phone, s.Address, s.UserID).
		Scan(&s.ID, &s.CreatedAt)
}

// GetByID returns a supplier by ID if it belongs to the organization.
func (m *SupplierModel) GetByID(id int64, organizationID int64) (*Supplier, error) {
	const q = `
		SELECT id, name, contact_person, email, phone, address, user_id, created_at
		FROM suppliers
		WHERE id = $1 AND user_id = $2`

	var s Supplier
	err := m.DB.QueryRow(context.Background(), q, id, organizationID).Scan(
		&s.ID, &s.Name, &s.ContactPerson, &s.Email, &s.Phone, &s.Address, &s.UserID, &s.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// GetAllForUser lists all suppliers for an organization (sin paginación - DEPRECATED).
func (m *SupplierModel) GetAllForUser(organizationID int64) ([]Supplier, error) {
	const q = `
		SELECT id, name, contact_person, email, phone, address, user_id, created_at
		FROM suppliers
		WHERE user_id = $1
		ORDER BY id`

	rows, err := m.DB.Query(context.Background(), q, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Supplier{} // Initialize as empty slice instead of nil
	for rows.Next() {
		var s Supplier
		if err := rows.Scan(&s.ID, &s.Name, &s.ContactPerson, &s.Email, &s.Phone, &s.Address, &s.UserID, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// GetAllForUserPaginated lists paginated suppliers for an organization.
func (m *SupplierModel) GetAllForUserPaginated(organizationID int64, filters Filters) ([]Supplier, Metadata, error) {
	// Contar total de registros
	const qCount = `SELECT COUNT(*) FROM suppliers WHERE user_id = $1`

	var totalRecords int
	err := m.DB.QueryRow(context.Background(), qCount, organizationID).Scan(&totalRecords)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// Query con paginación
	const q = `
		SELECT id, name, contact_person, email, phone, address, user_id, created_at
		FROM suppliers
		WHERE user_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`

	rows, err := m.DB.Query(context.Background(), q, organizationID, filters.PageSize, filters.Offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	out := []Supplier{}
	for rows.Next() {
		var s Supplier
		if err := rows.Scan(&s.ID, &s.Name, &s.ContactPerson, &s.Email, &s.Phone, &s.Address, &s.UserID, &s.CreatedAt); err != nil {
			return nil, Metadata{}, err
		}
		out = append(out, s)
	}
	if rows.Err() != nil {
		return nil, Metadata{}, rows.Err()
	}

	return out, metadata, nil
}

// Update updates a supplier if it belongs to the organization.
func (m *SupplierModel) Update(id int64, organizationID int64, s *Supplier) error {
	const q = `
		UPDATE suppliers
		SET name = $1, contact_person = $2, email = $3, phone = $4, address = $5
		WHERE id = $6 AND user_id = $7`

	tag, err := m.DB.Exec(context.Background(), q, s.Name, s.ContactPerson, s.Email, s.Phone, s.Address, id, organizationID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a supplier if it belongs to the organization.
func (m *SupplierModel) Delete(id int64, organizationID int64) error {
	const q = `
		DELETE FROM suppliers
		WHERE id = $1 AND user_id = $2`

	tag, err := m.DB.Exec(context.Background(), q, id, organizationID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			return ErrHasReferences
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
