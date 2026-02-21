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
		SELECT 
			p.id, 
			p.name, 
			p.sku, 
			p.description, 
			COALESCE(SUM(pb.quantity), 0) as quantity,
			p.user_id, 
			p.created_at
		FROM products p
		LEFT JOIN product_batches pb ON p.id = pb.product_id
		WHERE p.user_id = $1
		GROUP BY p.id, p.name, p.sku, p.description, p.user_id, p.created_at
		ORDER BY p.id`

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
