package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StockMovement representa un movimiento de inventario para un producto.
// quantity_change > 0 para entradas; < 0 para salidas.
// reason: 'SALES_ORDER', 'PURCHASE_ORDER', 'MANUAL_ADJUSTMENT', etc.
// reference_id puede contener el ID de la orden asociada (opcional).
type StockMovement struct {
	ID             int64     `json:"id"`
	ProductID      int64     `json:"product_id"`
	QuantityChange int       `json:"quantity_change"`
	Reason         string    `json:"reason"`
	ReferenceID    string    `json:"reference_id"`
	UserID         int64     `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// StockMovementModel para acceso a datos de movimientos
type StockMovementModel struct {
	DB *pgxpool.Pool
}

// GetForProduct devuelve los movimientos de stock de un producto para el usuario
func (m *StockMovementModel) GetForProduct(productID int64, userID int64) ([]StockMovement, error) {
	const q = `
		SELECT id, product_id, quantity_change, reason, reference_id, user_id, created_at
		FROM stock_movements
		WHERE product_id = $1 AND user_id = $2
		ORDER BY created_at`

	rows, err := m.DB.Query(context.Background(), q, productID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []StockMovement{} // Initialize as empty slice instead of nil
	for rows.Next() {
		var sm StockMovement
		if err := rows.Scan(&sm.ID, &sm.ProductID, &sm.QuantityChange, &sm.Reason, &sm.ReferenceID, &sm.UserID, &sm.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sm)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}
