package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Supplier represents a supplier from the database.
type Supplier struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// SupplierModel wraps DB access for suppliers.
type SupplierModel struct {
	DB *pgxpool.Pool
}

// GetAllForUser returns all suppliers for a given user.
func (m *SupplierModel) GetAllForUser(userID int64) ([]Supplier, error) {
	const q = `
		SELECT id, name, email, phone, address, user_id, created_at
		FROM suppliers
		WHERE user_id = $1
		ORDER BY id`

	rows, err := m.DB.Query(context.Background(), q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suppliers := []Supplier{}
	for rows.Next() {
		var s Supplier
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.Phone, &s.Address, &s.UserID, &s.CreatedAt); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, s)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return suppliers, nil
}
