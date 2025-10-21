package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

// Create inserts a sales order with items and updates stock atomically.
func (m *SalesOrderModel) Create(order *SalesOrder, items []OrderItem) error {
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

	// Insert items and update stock
	const insertItem = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	const updateStock = `
		UPDATE products SET quantity = quantity - $1
		WHERE id = $2 AND quantity - $1 >= 0`

	for i := range items {
		items[i].OrderID = order.ID
		// Default price 0 for now
		// Insert item
		if err := tx.QueryRow(ctx, insertItem, items[i].OrderID, items[i].ProductID, items[i].Quantity, items[i].UnitPrice).
			Scan(&items[i].ID); err != nil {
			return err
		}
		// Update stock, ensure non-negative
		tag, err := tx.Exec(ctx, updateStock, items[i].Quantity, items[i].ProductID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrInsufficientStock
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
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil
	return nil
}

// GetAllForUser returns all sales orders for the given user.
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
