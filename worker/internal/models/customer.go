package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Customer represents a customer from the database.
type Customer struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CustomerModel wraps DB access for customers.
type CustomerModel struct {
	DB *pgxpool.Pool
}

// GetAllForUser returns all customers for a given user.
func (m *CustomerModel) GetAllForUser(userID int64) ([]Customer, error) {
	const q = `
		SELECT id, name, email, phone, address, user_id, created_at
		FROM customers
		WHERE user_id = $1
		ORDER BY id`

	rows, err := m.DB.Query(context.Background(), q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := []Customer{}
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Address, &c.UserID, &c.CreatedAt); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return customers, nil
}
