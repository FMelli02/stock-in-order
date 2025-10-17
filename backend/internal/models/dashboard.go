package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardMetrics contiene valores agregados para el dashboard
type DashboardMetrics struct {
	TotalProducts      int `json:"total_products"`
	TotalCustomers     int `json:"total_customers"`
	TotalSuppliers     int `json:"total_suppliers"`
	PendingSalesOrders int `json:"pending_sales_orders"`
	ProductsLowStock   int `json:"products_low_stock"`
}

// DashboardModel accede a datos agregados
type DashboardModel struct {
	DB *pgxpool.Pool
}

// GetMetrics ejecuta consultas de agregación para un usuario
func (m *DashboardModel) GetMetrics(userID int64) (DashboardMetrics, error) {
	ctx := context.Background()

	var metrics DashboardMetrics

	// TotalProducts
	if err := m.DB.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE user_id = $1`, userID).Scan(&metrics.TotalProducts); err != nil {
		return DashboardMetrics{}, err
	}
	// TotalCustomers
	if err := m.DB.QueryRow(ctx, `SELECT COUNT(*) FROM customers WHERE user_id = $1`, userID).Scan(&metrics.TotalCustomers); err != nil {
		return DashboardMetrics{}, err
	}
	// TotalSuppliers
	if err := m.DB.QueryRow(ctx, `SELECT COUNT(*) FROM suppliers WHERE user_id = $1`, userID).Scan(&metrics.TotalSuppliers); err != nil {
		return DashboardMetrics{}, err
	}
	// PendingSalesOrders
	if err := m.DB.QueryRow(ctx, `SELECT COUNT(*) FROM sales_orders WHERE user_id = $1 AND status = 'pending'`, userID).Scan(&metrics.PendingSalesOrders); err != nil {
		return DashboardMetrics{}, err
	}
	// ProductsLowStock (threshold fijo = 5)
	if err := m.DB.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE user_id = $1 AND quantity <= 5`, userID).Scan(&metrics.ProductsLowStock); err != nil {
		return DashboardMetrics{}, err
	}

	return metrics, nil
}
