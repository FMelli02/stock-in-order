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

// GetByID returns a product by ID for a given organization.
// Ahora calcula quantity desde product_batches con SUM()
func (m *ProductModel) GetByID(id int64, organizationID int64) (*Product, error) {
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
	err := m.DB.QueryRow(context.Background(), q, id, organizationID).Scan(
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

// GetAllForUser returns all products for a given organization (sin paginación - DEPRECATED).
// Usar GetAllForUserPaginated en su lugar.
func (m *ProductModel) GetAllForUser(organizationID int64) ([]Product, error) {
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

	rows, err := m.DB.Query(context.Background(), q, organizationID)
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

// GetAllForUserPaginated returns paginated products for a given organization.
func (m *ProductModel) GetAllForUserPaginated(organizationID int64, filters Filters) ([]Product, Metadata, error) {
	// Query para obtener el total de registros
	const qCount = `
		SELECT COUNT(DISTINCT p.id)
		FROM products p
		WHERE p.user_id = $1`

	var totalRecords int
	err := m.DB.QueryRow(context.Background(), qCount, organizationID).Scan(&totalRecords)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Calcular metadata
	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// Query principal con paginación
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
		ORDER BY p.name
		LIMIT $2 OFFSET $3`

	rows, err := m.DB.Query(context.Background(), q, organizationID, filters.PageSize, filters.Offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.StockMinimo, &p.Notificado, &p.UserID, &p.CreatedAt, &p.CalculatedQuantity); err != nil {
			return nil, Metadata{}, err
		}
		products = append(products, p)
	}
	if rows.Err() != nil {
		return nil, Metadata{}, rows.Err()
	}

	return products, metadata, nil
}

// Update updates a product if it belongs to the organization.
// NOTE: Ya no actualiza quantity - se maneja via product_batches
func (m *ProductModel) Update(id int64, organizationID int64, p *Product) error {
	const q = `
		UPDATE products
		SET name = $1, sku = $2, description = $3, stock_minimo = $4
		WHERE id = $5 AND user_id = $6`

	tag, err := m.DB.Exec(context.Background(), q, p.Name, p.SKU, p.Description, p.StockMinimo, id, organizationID)
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

// Delete deletes a product if it belongs to the organization.
func (m *ProductModel) Delete(id int64, organizationID int64) error {
	const q = `
		DELETE FROM products
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

// AdjustStock está DEPRECATED - ahora se debe usar product_batches
// Este método será eliminado en versiones futuras
// Para ajustar stock, crear/modificar lotes en product_batches
func (m *ProductModel) AdjustStock(productID int64, organizationID int64, quantityChange int, reason string) error {
	// DEPRECATED: La columna quantity ya no existe en products
	// El stock ahora se calcula desde product_batches
	return errors.New("AdjustStock is deprecated - use product_batches instead")
}

// GetBatchesByProductID returns all active batches for a product (quantity > 0)
// ordered by expiry date (FEFO - First Expired First Out)
func (m *ProductModel) GetBatchesByProductID(productID int64, organizationID int64) ([]ProductBatch, error) {
	const q = `
		SELECT id, product_id, user_id, lote_number, quantity, expiry_date, created_at
		FROM product_batches
		WHERE product_id = $1 AND user_id = $2 AND quantity > 0
		ORDER BY expiry_date ASC NULLS LAST, created_at ASC`

	rows, err := m.DB.Query(context.Background(), q, productID, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []ProductBatch
	for rows.Next() {
		var b ProductBatch
		if err := rows.Scan(&b.ID, &b.ProductID, &b.UserID, &b.LoteNumber, &b.Quantity, &b.ExpiryDate, &b.CreatedAt); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	// Return empty slice instead of nil for consistency
	if batches == nil {
		batches = []ProductBatch{}
	}

	return batches, nil
}
