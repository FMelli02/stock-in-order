package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SalesOrderFilters contiene los filtros opcionales para exportar órdenes de venta
type SalesOrderFilters struct {
	DateFrom string
	DateTo   string
	Status   string
}

// SalesOrder represents the header of a sales order.
type SalesOrder struct {
	ID           int64           `json:"id"`
	CustomerID   sql.NullInt64   `json:"customer_id"`
	CustomerName string          `json:"customer_name,omitempty"`
	OrderDate    time.Time       `json:"order_date"`
	Status       string          `json:"status"`
	TotalAmount  sql.NullFloat64 `json:"total_amount"`
	UserID       int64           `json:"user_id"`
}

// OrderItem represents a product item belonging to a sales order.
type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// SalesOrderModel wraps DB access for sales orders.
type SalesOrderModel struct {
	DB *pgxpool.Pool
}

// ErrInsufficientStock is returned when available stock is not enough.
var ErrInsufficientStock = errors.New("insufficient stock")

// InsufficientStockError provides detailed information about stock shortage
type InsufficientStockError struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	Requested   int    `json:"requested"`
	Available   int    `json:"available"`
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for product %s (ID: %d): requested %d, available %d",
		e.ProductName, e.ProductID, e.Requested, e.Available)
}

// ValidateStockAvailability checks if there's enough total stock for all items BEFORE starting a transaction.
// This prevents unnecessary database operations and provides better error messages.
func (m *SalesOrderModel) ValidateStockAvailability(items []OrderItem, userID int64) error {
	ctx := context.Background()

	for _, item := range items {
		// Query total available stock across all batches
		const qTotalStock = `
			SELECT COALESCE(SUM(pb.quantity), 0), p.name
			FROM product_batches pb
			JOIN products p ON pb.product_id = p.id
			WHERE pb.product_id = $1 AND pb.user_id = $2 AND pb.quantity > 0
			GROUP BY p.name`

		var availableStock int
		var productName string
		err := m.DB.QueryRow(ctx, qTotalStock, item.ProductID, userID).Scan(&availableStock, &productName)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// No batches found, stock is 0
				// Try to get product name separately
				const qProductName = `SELECT name FROM products WHERE id = $1 AND user_id = $2`
				if nameErr := m.DB.QueryRow(ctx, qProductName, item.ProductID, userID).Scan(&productName); nameErr != nil {
					productName = fmt.Sprintf("Product #%d", item.ProductID)
				}
				return &InsufficientStockError{
					ProductID:   item.ProductID,
					ProductName: productName,
					Requested:   item.Quantity,
					Available:   0,
				}
			}
			slog.Error("ValidateStockAvailability: failed to query stock", "productID", item.ProductID, "error", err)
			return err
		}

		// Check if available stock is sufficient
		if availableStock < item.Quantity {
			slog.Warn("ValidateStockAvailability: insufficient stock detected",
				"productID", item.ProductID,
				"productName", productName,
				"requested", item.Quantity,
				"available", availableStock)

			return &InsufficientStockError{
				ProductID:   item.ProductID,
				ProductName: productName,
				Requested:   item.Quantity,
				Available:   availableStock,
			}
		}

		slog.Info("ValidateStockAvailability: stock check passed",
			"productID", item.ProductID,
			"productName", productName,
			"requested", item.Quantity,
			"available", availableStock)
	}

	return nil
}

// ConsumeStockFEFO implements First Expired First Out logic for stock consumption.
// It deducts the requested quantity from product batches, prioritizing those that expire first.
// This function MUST be called within a transaction.
func ConsumeStockFEFO(ctx context.Context, tx pgx.Tx, productID int64, userID int64, quantityToConsume int) error {
	slog.Info("ConsumeStockFEFO: starting FEFO consumption",
		"productID", productID,
		"userID", userID,
		"quantityNeeded", quantityToConsume)

	// Get batches ordered by expiry date (FEFO - First Expired, First Out)
	// NULL expiry dates go last (NULLS LAST)
	// If expiry dates are equal, use oldest batch first (created_at ASC)
	// FOR UPDATE locks the rows for this transaction
	const qBatches = `
		SELECT id, quantity, expiry_date, lote_number
		FROM product_batches
		WHERE product_id = $1 AND user_id = $2 AND quantity > 0
		ORDER BY expiry_date ASC NULLS LAST, created_at ASC
		FOR UPDATE`

	rows, err := tx.Query(ctx, qBatches, productID, userID)
	if err != nil {
		slog.Error("ConsumeStockFEFO: failed to query batches", "error", err)
		return err
	}
	defer rows.Close()

	type batch struct {
		id         int64
		quantity   int
		expiryDate sql.NullTime
		loteNumber string
	}

	var batches []batch
	for rows.Next() {
		var b batch
		if err := rows.Scan(&b.id, &b.quantity, &b.expiryDate, &b.loteNumber); err != nil {
			slog.Error("ConsumeStockFEFO: failed to scan batch", "error", err)
			return err
		}
		batches = append(batches, b)
	}
	rows.Close()

	if rows.Err() != nil {
		slog.Error("ConsumeStockFEFO: rows error", "error", rows.Err())
		return rows.Err()
	}

	slog.Info("ConsumeStockFEFO: found batches", "count", len(batches))

	// Iterate through batches and consume stock
	remaining := quantityToConsume
	const updateBatch = `UPDATE product_batches SET quantity = $1 WHERE id = $2`

	for _, b := range batches {
		if remaining <= 0 {
			break
		}

		expiryStr := "sin vencimiento"
		if b.expiryDate.Valid {
			expiryStr = b.expiryDate.Time.Format("2006-01-02")
		}

		slog.Info("ConsumeStockFEFO: processing batch",
			"batchID", b.id,
			"lote", b.loteNumber,
			"available", b.quantity,
			"needed", remaining,
			"expiry", expiryStr)

		if b.quantity >= remaining {
			// This batch has enough stock to fulfill the remaining quantity
			newQuantity := b.quantity - remaining
			if _, err := tx.Exec(ctx, updateBatch, newQuantity, b.id); err != nil {
				slog.Error("ConsumeStockFEFO: failed to update batch", "batchID", b.id, "error", err)
				return err
			}
			slog.Info("ConsumeStockFEFO: consumed from batch",
				"batchID", b.id,
				"consumed", remaining,
				"newQuantity", newQuantity)
			remaining = 0
			break
		} else {
			// This batch doesn't have enough, consume all of it
			if _, err := tx.Exec(ctx, updateBatch, 0, b.id); err != nil {
				slog.Error("ConsumeStockFEFO: failed to update batch", "batchID", b.id, "error", err)
				return err
			}
			slog.Info("ConsumeStockFEFO: batch depleted",
				"batchID", b.id,
				"consumed", b.quantity)
			remaining -= b.quantity
		}
	}

	// If we still have remaining quantity, there's insufficient stock
	if remaining > 0 {
		slog.Error("ConsumeStockFEFO: insufficient stock",
			"productID", productID,
			"needed", quantityToConsume,
			"missing", remaining)
		return ErrInsufficientStock
	}

	slog.Info("ConsumeStockFEFO: stock consumed successfully",
		"productID", productID,
		"quantityConsumed", quantityToConsume)
	return nil
}

// Create inserts a sales order with items and updates stock atomically.
func (m *SalesOrderModel) Create(order *SalesOrder, items []OrderItem) error {
	ctx := context.Background()

	// CRITICAL: Validate stock availability BEFORE starting the transaction
	// This prevents unnecessary DB operations and provides better error messages
	if err := m.ValidateStockAvailability(items, order.UserID); err != nil {
		slog.Error("Create: stock validation failed", "error", err)
		return err
	}

	slog.Info("Create: stock validation passed, starting transaction", "orderItems", len(items))

	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Insert order header
	const insertOrder = `
		INSERT INTO sales_orders (customer_id, order_date, status, total_amount, user_id)
		VALUES ($1, COALESCE($2, NOW()), COALESCE($3, 'pending'), $4, $5)
		RETURNING id, order_date`

	if err := tx.QueryRow(ctx, insertOrder,
		order.CustomerID, order.OrderDate, order.Status, order.TotalAmount, order.UserID,
	).Scan(&order.ID, &order.OrderDate); err != nil {
		return err
	}

	// Insert items and consume stock using FEFO logic
	const insertItem = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	for i := range items {
		items[i].OrderID = order.ID

		slog.Info("Create: processing order item",
			"orderID", order.ID,
			"productID", items[i].ProductID,
			"quantity", items[i].Quantity)

		// CRITICAL: Consume stock using FEFO logic BEFORE inserting the item
		// This ensures we fail fast if there's insufficient stock
		if err := ConsumeStockFEFO(ctx, tx, items[i].ProductID, order.UserID, items[i].Quantity); err != nil {
			slog.Error("Create: failed to consume stock",
				"productID", items[i].ProductID,
				"quantity", items[i].Quantity,
				"error", err)
			return err
		}

		// Insert order item
		if err := tx.QueryRow(ctx, insertItem, items[i].OrderID, items[i].ProductID, items[i].Quantity, items[i].UnitPrice).
			Scan(&items[i].ID); err != nil {
			slog.Error("Create: failed to insert order item", "error", err)
			return err
		}

		// Insert stock movement (negative for sales)
		const insertMovement = `
			INSERT INTO stock_movements (product_id, quantity_change, reason, reference_id, user_id)
			VALUES ($1, $2, $3, $4, $5)`
		if _, err := tx.Exec(ctx, insertMovement,
			items[i].ProductID,
			-items[i].Quantity,
			"SALES_ORDER",
			fmt.Sprintf("%d", order.ID),
			order.UserID,
		); err != nil {
			slog.Error("Create: failed to insert stock movement", "error", err)
			return err
		}

		slog.Info("Create: order item processed successfully",
			"itemID", items[i].ID,
			"productID", items[i].ProductID)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil
	return nil
}

// GetAllForUser returns all sales orders for the given user.
// GetAllForUser lists all sales orders for an organization (sin paginación - DEPRECATED).
func (m *SalesOrderModel) GetAllForUser(userID int64) ([]SalesOrder, error) {
	const q = `
		SELECT 
			so.id, so.customer_id, so.order_date, so.status, so.total_amount, so.user_id,
			c.name AS customer_name
		FROM sales_orders so
		LEFT JOIN customers c ON so.customer_id = c.id
		WHERE so.user_id = $1
		ORDER BY so.id DESC`

	rows, err := m.DB.Query(context.Background(), q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []SalesOrder{} // Initialize as empty slice instead of nil
	for rows.Next() {
		var o SalesOrder
		var customerName sql.NullString
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.OrderDate, &o.Status, &o.TotalAmount, &o.UserID, &customerName); err != nil {
			return nil, err
		}
		if customerName.Valid {
			o.CustomerName = customerName.String
		}
		out = append(out, o)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// GetAllForUserPaginated lists paginated sales orders for an organization.
func (m *SalesOrderModel) GetAllForUserPaginated(userID int64, filters Filters) ([]SalesOrder, Metadata, error) {
	// Contar total de registros
	const qCount = `
		SELECT COUNT(*) 
		FROM sales_orders so 
		WHERE so.user_id = $1`

	var totalRecords int
	err := m.DB.QueryRow(context.Background(), qCount, userID).Scan(&totalRecords)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// Query con paginación y JOIN con customers
	const q = `
		SELECT 
			so.id, so.customer_id, so.order_date, so.status, so.total_amount, so.user_id,
			c.name AS customer_name
		FROM sales_orders so
		LEFT JOIN customers c ON so.customer_id = c.id
		WHERE so.user_id = $1
		ORDER BY so.order_date DESC
		LIMIT $2 OFFSET $3`

	rows, err := m.DB.Query(context.Background(), q, userID, filters.PageSize, filters.Offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	out := []SalesOrder{}
	for rows.Next() {
		var o SalesOrder
		var customerName sql.NullString
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.OrderDate, &o.Status, &o.TotalAmount, &o.UserID, &customerName); err != nil {
			return nil, Metadata{}, err
		}
		if customerName.Valid {
			o.CustomerName = customerName.String
		}
		out = append(out, o)
	}
	if rows.Err() != nil {
		return nil, Metadata{}, rows.Err()
	}

	return out, metadata, nil
}

// GetAllForUserWithFilters returns sales orders for the user with optional filters
func (m *SalesOrderModel) GetAllForUserWithFilters(userID int64, filters SalesOrderFilters) ([]SalesOrder, error) {
	query := `
		SELECT 
			so.id, so.customer_id, so.order_date, so.status, so.total_amount, so.user_id,
			c.name AS customer_name
		FROM sales_orders so
		LEFT JOIN customers c ON so.customer_id = c.id
		WHERE so.user_id = $1`

	args := []interface{}{userID}
	argIndex := 2

	// Agregar filtros dinámicamente
	if filters.DateFrom != "" {
		query += fmt.Sprintf(" AND so.order_date >= $%d", argIndex)
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		query += fmt.Sprintf(" AND so.order_date <= $%d", argIndex)
		args = append(args, filters.DateTo)
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND so.status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	query += " ORDER BY so.id DESC"

	rows, err := m.DB.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []SalesOrder{}
	for rows.Next() {
		var o SalesOrder
		var customerName sql.NullString
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.OrderDate, &o.Status, &o.TotalAmount, &o.UserID, &customerName); err != nil {
			return nil, err
		}
		if customerName.Valid {
			o.CustomerName = customerName.String
		}
		out = append(out, o)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// GetByID returns a specific order for the user along with its items.
func (m *SalesOrderModel) GetByID(orderID int64, userID int64) (*SalesOrder, []OrderItem, error) {
	const qOrder = `
		SELECT 
			so.id, so.customer_id, so.order_date, so.status, so.total_amount, so.user_id,
			c.name AS customer_name
		FROM sales_orders so
		LEFT JOIN customers c ON so.customer_id = c.id
		WHERE so.id = $1 AND so.user_id = $2`

	var o SalesOrder
	var customerName sql.NullString
	err := m.DB.QueryRow(context.Background(), qOrder, orderID, userID).
		Scan(&o.ID, &o.CustomerID, &o.OrderDate, &o.Status, &o.TotalAmount, &o.UserID, &customerName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	if customerName.Valid {
		o.CustomerName = customerName.String
	}

	const qItems = `
		SELECT id, order_id, product_id, quantity, unit_price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`
	rows, err := m.DB.Query(context.Background(), qItems, orderID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Quantity, &it.UnitPrice); err != nil {
			return nil, nil, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}
	return &o, items, nil
}
