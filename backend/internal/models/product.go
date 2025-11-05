package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Product represents a product belonging to a user.
// NOTE: Quantity ha sido removido de la tabla - ahora se calcula desde product_batches
type Product struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	SKU                string    `json:"sku"`
	Description        *string   `json:"description,omitempty"`
	CalculatedQuantity int       `json:"quantity"` // Calculado via SUM() de product_batches
	StockMinimo        int       `json:"stock_minimo"`
	Notificado         bool      `json:"notificado"`
	UserID             int64     `json:"user_id"`
	CreatedAt          time.Time `json:"created_at"`
}

// Errors for product operations
var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicateSKU  = errors.New("duplicate sku")
	ErrHasReferences = errors.New("cannot delete: record has references in other tables")
)

// ProductModel wraps DB access for products.
type ProductModel struct {
	DB *pgxpool.Pool
}

// Insert inserts a new product for a user and sets ID and CreatedAt.
// NOTE: Ya no inserta quantity - se manejará via product_batches
func (m *ProductModel) Insert(p *Product) error {
	const q = `
		INSERT INTO products (name, sku, description, stock_minimo, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, notificado`

	err := m.DB.QueryRow(context.Background(), q, p.Name, p.SKU, p.Description, p.StockMinimo, p.UserID).
		Scan(&p.ID, &p.CreatedAt, &p.Notificado)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation (user_id, sku)
			return ErrDuplicateSKU
		}
		return err
	}
	return nil
}

// GetByID returns a product by ID for a given user.
// Ahora calcula quantity desde product_batches con SUM()
func (m *ProductModel) GetByID(id int64, userID int64) (*Product, error) {
	const q = `
		SELECT 
			p.id, 
			p.name, 
			p.sku, 
			p.description, 
			p.stock_minimo, 
			p.notificado, 
			p.user_id, 
			p.created_at,
			COALESCE(SUM(pb.quantity), 0) AS calculated_quantity
		FROM products p
		LEFT JOIN product_batches pb ON p.id = pb.product_id
		WHERE p.id = $1 AND p.user_id = $2
		GROUP BY p.id, p.name, p.sku, p.description, p.stock_minimo, p.notificado, p.user_id, p.created_at`

	var p Product
	err := m.DB.QueryRow(context.Background(), q, id, userID).Scan(
		&p.ID, &p.Name, &p.SKU, &p.Description, &p.StockMinimo, &p.Notificado, &p.UserID, &p.CreatedAt, &p.CalculatedQuantity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GetAllForUser returns all products for a given user.
// Ahora calcula quantity desde product_batches con SUM()
func (m *ProductModel) GetAllForUser(userID int64) ([]Product, error) {
	const q = `
		SELECT 
			p.id, 
			p.name, 
			p.sku, 
			p.description, 
			p.stock_minimo, 
			p.notificado, 
			p.user_id, 
			p.created_at,
			COALESCE(SUM(pb.quantity), 0) AS calculated_quantity
		FROM products p
		LEFT JOIN product_batches pb ON p.id = pb.product_id
		WHERE p.user_id = $1
		GROUP BY p.id, p.name, p.sku, p.description, p.stock_minimo, p.notificado, p.user_id, p.created_at
		ORDER BY p.name`

	rows, err := m.DB.Query(context.Background(), q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{} // Initialize as empty slice instead of nil
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.StockMinimo, &p.Notificado, &p.UserID, &p.CreatedAt, &p.CalculatedQuantity); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return products, nil
}

// Update updates a product if it belongs to the user.
// NOTE: Ya no actualiza quantity - se maneja via product_batches
func (m *ProductModel) Update(id int64, userID int64, p *Product) error {
	const q = `
		UPDATE products
		SET name = $1, sku = $2, description = $3, stock_minimo = $4
		WHERE id = $5 AND user_id = $6`

	tag, err := m.DB.Exec(context.Background(), q, p.Name, p.SKU, p.Description, p.StockMinimo, id, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ErrDuplicateSKU
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a product if it belongs to the user.
func (m *ProductModel) Delete(id int64, userID int64) error {
	const q = `
		DELETE FROM products
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

// AdjustStock está DEPRECATED - ahora se debe usar product_batches
// Este método será eliminado en versiones futuras
// Para ajustar stock, crear/modificar lotes en product_batches
func (m *ProductModel) AdjustStock(productID int64, userID int64, quantityChange int, reason string) error {
	// DEPRECATED: La columna quantity ya no existe en products
	// El stock ahora se calcula desde product_batches
	return errors.New("AdjustStock is deprecated - use product_batches instead")
}
