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

// PurchaseOrderFilters contiene los filtros opcionales para exportar órdenes de compra
type PurchaseOrderFilters struct {
	DateFrom string
	DateTo   string
	Status   string
}

// PurchaseOrder represents the header of a purchase order.
// Mirrors the sales order but linked to a supplier.
type PurchaseOrder struct {
	ID           int64         `json:"id"`
	SupplierID   sql.NullInt64 `json:"supplier_id"`
	SupplierName string        `json:"supplier_name,omitempty"`
	OrderDate    *time.Time    `json:"order_date,omitempty"` // Pointer to allow NULL/omitted value
	Status       string        `json:"status"`
	UserID       int64         `json:"user_id"`
}

// PurchaseOrderItem represents a product item belonging to a purchase order.
// Stock will be increased when items are received (handled elsewhere).
type PurchaseOrderItem struct {
	ID              int64      `json:"id"`
	PurchaseOrderID int64      `json:"purchase_order_id"`
	ProductID       int64      `json:"product_id"`
	Quantity        int        `json:"quantity"`
	UnitCost        float64    `json:"unit_cost"`
	LoteNumber      string     `json:"lote_number,omitempty"` // Número de lote (opcional)
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"` // Fecha de vencimiento (opcional)
}

// PurchaseOrderModel wraps DB access for purchase orders.
type PurchaseOrderModel struct {
	DB *pgxpool.Pool
}

// Create inserts a purchase order and its items atomically. Does NOT update stock.
func (m *PurchaseOrderModel) Create(order *PurchaseOrder, items []PurchaseOrderItem) error {
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

	const insertOrder = `
		INSERT INTO purchase_orders (supplier_id, order_date, status, user_id)
		VALUES ($1, NOW(), COALESCE($2, 'pending'), $3)
		RETURNING id, order_date`

	var orderDate time.Time
	if err := tx.QueryRow(ctx, insertOrder,
		order.SupplierID, order.Status, order.UserID,
	).Scan(&order.ID, &orderDate); err != nil {
		return err
	}
	order.OrderDate = &orderDate

	const insertItem = `
		INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity, unit_cost, lote_number, expiry_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	for i := range items {
		items[i].PurchaseOrderID = order.ID
		if err := tx.QueryRow(ctx, insertItem,
			items[i].PurchaseOrderID,
			items[i].ProductID,
			items[i].Quantity,
			items[i].UnitCost,
			items[i].LoteNumber,
			items[i].ExpiryDate,
		).Scan(&items[i].ID); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil
	return nil
}

// GetAllForUser returns all purchase orders for the given user.
// GetAllForUser lists all purchase orders for an organization (sin paginación - DEPRECATED).
func (m *PurchaseOrderModel) GetAllForUser(userID int64) ([]PurchaseOrder, error) {
	const q = `
		SELECT 
			po.id, po.supplier_id, po.order_date, po.status, po.user_id,
			s.name AS supplier_name
		FROM purchase_orders po
		LEFT JOIN suppliers s ON po.supplier_id = s.id
		WHERE po.user_id = $1
		ORDER BY po.id DESC`

	rows, err := m.DB.Query(context.Background(), q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []PurchaseOrder{} // Initialize as empty slice instead of nil
	for rows.Next() {
		var o PurchaseOrder
		var supplierName sql.NullString
		var orderDate time.Time
		if err := rows.Scan(&o.ID, &o.SupplierID, &orderDate, &o.Status, &o.UserID, &supplierName); err != nil {
			return nil, err
		}
		o.OrderDate = &orderDate
		if supplierName.Valid {
			o.SupplierName = supplierName.String
		}
		out = append(out, o)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// GetAllForUserPaginated lists paginated purchase orders for an organization.
func (m *PurchaseOrderModel) GetAllForUserPaginated(userID int64, filters Filters) ([]PurchaseOrder, Metadata, error) {
	// Contar total de registros
	const qCount = `
		SELECT COUNT(*) 
		FROM purchase_orders po 
		WHERE po.user_id = $1`

	var totalRecords int
	err := m.DB.QueryRow(context.Background(), qCount, userID).Scan(&totalRecords)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// Query con paginación y JOIN con suppliers
	const q = `
		SELECT 
			po.id, po.supplier_id, po.order_date, po.status, po.user_id,
			s.name AS supplier_name
		FROM purchase_orders po
		LEFT JOIN suppliers s ON po.supplier_id = s.id
		WHERE po.user_id = $1
		ORDER BY po.order_date DESC
		LIMIT $2 OFFSET $3`

	rows, err := m.DB.Query(context.Background(), q, userID, filters.PageSize, filters.Offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	out := []PurchaseOrder{}
	for rows.Next() {
		var o PurchaseOrder
		var supplierName sql.NullString
		var orderDate time.Time
		if err := rows.Scan(&o.ID, &o.SupplierID, &orderDate, &o.Status, &o.UserID, &supplierName); err != nil {
			return nil, Metadata{}, err
		}
		o.OrderDate = &orderDate
		if supplierName.Valid {
			o.SupplierName = supplierName.String
		}
		out = append(out, o)
	}
	if rows.Err() != nil {
		return nil, Metadata{}, rows.Err()
	}

	return out, metadata, nil
}

// GetAllForUserWithFilters returns purchase orders for the user with optional filters
func (m *PurchaseOrderModel) GetAllForUserWithFilters(userID int64, filters PurchaseOrderFilters) ([]PurchaseOrder, error) {
	query := `
		SELECT 
			po.id, po.supplier_id, po.order_date, po.status, po.user_id,
			s.name AS supplier_name
		FROM purchase_orders po
		LEFT JOIN suppliers s ON po.supplier_id = s.id
		WHERE po.user_id = $1`

	args := []interface{}{userID}
	argIndex := 2

	// Agregar filtros dinámicamente
	if filters.DateFrom != "" {
		query += fmt.Sprintf(" AND po.order_date >= $%d", argIndex)
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		query += fmt.Sprintf(" AND po.order_date <= $%d", argIndex)
		args = append(args, filters.DateTo)
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND po.status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	query += " ORDER BY po.id DESC"

	rows, err := m.DB.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []PurchaseOrder{}
	for rows.Next() {
		var o PurchaseOrder
		var supplierName sql.NullString
		var orderDate time.Time
		if err := rows.Scan(&o.ID, &o.SupplierID, &orderDate, &o.Status, &o.UserID, &supplierName); err != nil {
			return nil, err
		}
		o.OrderDate = &orderDate
		if supplierName.Valid {
			o.SupplierName = supplierName.String
		}
		out = append(out, o)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// GetByID returns a specific purchase order for the user along with its items.
func (m *PurchaseOrderModel) GetByID(orderID int64, userID int64) (*PurchaseOrder, []PurchaseOrderItem, error) {
	const qOrder = `
		SELECT 
			po.id, po.supplier_id, po.order_date, po.status, po.user_id,
			s.name AS supplier_name
		FROM purchase_orders po
		LEFT JOIN suppliers s ON po.supplier_id = s.id
		WHERE po.id = $1 AND po.user_id = $2`

	var o PurchaseOrder
	var supplierName sql.NullString
	var orderDate time.Time
	err := m.DB.QueryRow(context.Background(), qOrder, orderID, userID).
		Scan(&o.ID, &o.SupplierID, &orderDate, &o.Status, &o.UserID, &supplierName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	o.OrderDate = &orderDate
	if supplierName.Valid {
		o.SupplierName = supplierName.String
	}

	const qItems = `
		SELECT id, purchase_order_id, product_id, quantity, unit_cost, lote_number, expiry_date
		FROM purchase_order_items
		WHERE purchase_order_id = $1
		ORDER BY id`
	rows, err := m.DB.Query(context.Background(), qItems, orderID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []PurchaseOrderItem
	for rows.Next() {
		var it PurchaseOrderItem
		var loteNumber sql.NullString
		var expiryDate sql.NullTime
		if err := rows.Scan(&it.ID, &it.PurchaseOrderID, &it.ProductID, &it.Quantity, &it.UnitCost, &loteNumber, &expiryDate); err != nil {
			return nil, nil, err
		}
		if loteNumber.Valid {
			it.LoteNumber = loteNumber.String
		}
		if expiryDate.Valid {
			it.ExpiryDate = &expiryDate.Time
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}
	return &o, items, nil
}

// UpdateStatus updates the status of a purchase order. If setting to 'completed', increases product stock for all items.
func (m *PurchaseOrderModel) UpdateStatus(orderID int64, userID int64, newStatus string) error {
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

	// Lock the order row and verify ownership
	const qOrder = `SELECT status FROM purchase_orders WHERE id = $1 AND user_id = $2 FOR UPDATE`
	var current string
	if err := tx.QueryRow(ctx, qOrder, orderID, userID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	// If transitioning to completed and not already completed, increase stock
	if newStatus == "completed" && current != "completed" {
		slog.Info("UpdateStatus: transitioning to completed", "orderID", orderID, "userID", userID)
		const qItems = `
			SELECT product_id, quantity, lote_number, expiry_date
			FROM purchase_order_items
			WHERE purchase_order_id = $1`
		rows, err := tx.Query(ctx, qItems, orderID)
		if err != nil {
			slog.Error("UpdateStatus: failed to query items", "error", err)
			return err
		}

		// Read all items into a slice first (can't use tx while iterating rows)
		type item struct {
			productID  int64
			qty        int
			loteNumber sql.NullString
			expiryDate sql.NullTime
		}
		var items []item
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.productID, &it.qty, &it.loteNumber, &it.expiryDate); err != nil {
				rows.Close()
				slog.Error("UpdateStatus: failed to scan item", "error", err)
				return err
			}
			items = append(items, it)
		}
		rows.Close()

		if rows.Err() != nil {
			slog.Error("UpdateStatus: rows error", "error", rows.Err())
			return rows.Err()
		}

		// Now create batch entries for each item (tx is free now)
		const insertBatch = `
			INSERT INTO product_batches (product_id, user_id, quantity, lote_number, expiry_date)
			VALUES ($1, $2, $3, $4, $5)`
		const resetNotified = `UPDATE products SET notificado = false WHERE id = $1`

		for _, it := range items {
			slog.Info("UpdateStatus: creating product batch",
				"productID", it.productID,
				"qty", it.qty,
				"lote", it.loteNumber.String,
				"expiry", it.expiryDate.Time,
				"userID", userID)

			// Prepare lote_number (use empty string if NULL)
			loteNum := ""
			if it.loteNumber.Valid {
				loteNum = it.loteNumber.String
			}

			// Prepare expiry_date (use NULL if not provided)
			var expiryDate interface{}
			if it.expiryDate.Valid {
				expiryDate = it.expiryDate.Time
			} else {
				expiryDate = nil
			}

			// Insert new batch
			_, err := tx.Exec(ctx, insertBatch, it.productID, userID, it.qty, loteNum, expiryDate)
			if err != nil {
				slog.Error("UpdateStatus: failed to create product batch", "productID", it.productID, "error", err)
				return err
			}
			slog.Info("UpdateStatus: product batch created successfully", "productID", it.productID)

			// Reset notificado flag (product now has stock from new batch)
			_, err = tx.Exec(ctx, resetNotified, it.productID)
			if err != nil {
				slog.Error("UpdateStatus: failed to reset notificado flag", "productID", it.productID, "error", err)
				return err
			}

			// Insert stock movement (positive for purchase)
			const insertMovement = `
				INSERT INTO stock_movements (product_id, quantity_change, reason, reference_id, user_id)
				VALUES ($1, $2, $3, $4, $5)`
			if _, err := tx.Exec(ctx, insertMovement,
				it.productID,
				it.qty,
				"PURCHASE_ORDER",
				fmt.Sprintf("%d", orderID),
				userID,
			); err != nil {
				return err
			}
		}
	}

	// Update the status
	const upd = `UPDATE purchase_orders SET status = $1 WHERE id = $2`
	if _, err := tx.Exec(ctx, upd, newStatus, orderID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil
	return nil
}
