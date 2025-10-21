package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Customer represents a customer belonging to a user.
type Customer struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address,omitempty"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CustomerModel wraps DB access for customers.
type CustomerModel struct {
	DB *pgxpool.Pool
}

// Insert creates a new customer for a user and sets ID and CreatedAt.
func (m *CustomerModel) Insert(c *Customer) error {
	const q = `
		INSERT INTO customers (name, email, phone, address, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return m.DB.QueryRow(context.Background(), q, c.Name, c.Email, c.Phone, c.Address, c.UserID).
		Scan(&c.ID, &c.CreatedAt)
}

// GetByID returns a customer by ID if it belongs to the user.
func (m *CustomerModel) GetByID(id int64, userID int64) (*Customer, error) {
	const q = `
		SELECT id, name, email, phone, address, user_id, created_at
		FROM customers
		WHERE id = $1 AND user_id = $2`

	var c Customer
	err := m.DB.QueryRow(context.Background(), q, id, userID).Scan(
		&c.ID, &c.Name, &c.Email, &c.Phone, &c.Address, &c.UserID, &c.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// GetAllForUser lists all customers for a user.
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

	out := []Customer{} // Initialize as empty slice instead of nil
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Address, &c.UserID, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// Update updates a customer if it belongs to the user.
func (m *CustomerModel) Update(id int64, userID int64, c *Customer) error {
	const q = `
		UPDATE customers
		SET name = $1, email = $2, phone = $3, address = $4
		WHERE id = $5 AND user_id = $6`

	tag, err := m.DB.Exec(context.Background(), q, c.Name, c.Email, c.Phone, c.Address, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a customer if it belongs to the user.
func (m *CustomerModel) Delete(id int64, userID int64) error {
	const q = `
		DELETE FROM customers
		WHERE id = $1 AND user_id = $2`

	tag, err := m.DB.Exec(context.Background(), q, id, userID)
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
