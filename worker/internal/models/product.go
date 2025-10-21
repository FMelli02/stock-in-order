package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Product represents a product from the database.
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Description *string   `json:"description,omitempty"`
	Quantity    int       `json:"quantity"`
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProductModel wraps DB access for products.
type ProductModel struct {
	DB *pgxpool.Pool
}

// GetAllForUser returns all products for a given user.
func (m *ProductModel) GetAllForUser(userID int64) ([]Product, error) {
	const q = `
		SELECT id, name, sku, description, quantity, user_id, created_at
		FROM products
		WHERE user_id = $1
		ORDER BY id`

	rows, err := m.DB.Query(context.Background(), q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.Quantity, &p.UserID, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return products, nil
}
