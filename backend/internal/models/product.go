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
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Description *string   `json:"description,omitempty"`
	Quantity    int       `json:"quantity"`
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
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
func (m *ProductModel) Insert(p *Product) error {
	const q = `
		INSERT INTO products (name, sku, description, quantity, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := m.DB.QueryRow(context.Background(), q, p.Name, p.SKU, p.Description, p.Quantity, p.UserID).
		Scan(&p.ID, &p.CreatedAt)
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
func (m *ProductModel) GetByID(id int64, userID int64) (*Product, error) {
	const q = `
		SELECT id, name, sku, description, quantity, user_id, created_at
		FROM products
		WHERE id = $1 AND user_id = $2`

	var p Product
	err := m.DB.QueryRow(context.Background(), q, id, userID).Scan(
		&p.ID, &p.Name, &p.SKU, &p.Description, &p.Quantity, &p.UserID, &p.CreatedAt,
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

	products := []Product{} // Initialize as empty slice instead of nil
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

// Update updates a product if it belongs to the user.
func (m *ProductModel) Update(id int64, userID int64, p *Product) error {
	const q = `
		UPDATE products
		SET name = $1, sku = $2, description = $3, quantity = $4
		WHERE id = $5 AND user_id = $6`

	tag, err := m.DB.Exec(context.Background(), q, p.Name, p.SKU, p.Description, p.Quantity, id, userID)
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

// AdjustStock ajusta la cantidad de un producto y registra el movimiento de stock en una transacción.
func (m *ProductModel) AdjustStock(productID int64, userID int64, quantityChange int, reason string) error {
	ctx := context.Background()
	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Actualiza la cantidad del producto, validando que pertenezca al usuario
	const upd = `UPDATE products SET quantity = quantity + $1 WHERE id = $2 AND user_id = $3`
	tag, err := tx.Exec(ctx, upd, quantityChange, productID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Inserta el movimiento de stock; reference_id es NULL para ajuste manual
	const insertMovement = `
		INSERT INTO stock_movements (product_id, quantity_change, reason, reference_id, user_id)
		VALUES ($1, $2, $3, $4, $5)`
	var refID any = nil
	if _, err := tx.Exec(ctx, insertMovement, productID, quantityChange, reason, refID, userID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil
	return nil
}
