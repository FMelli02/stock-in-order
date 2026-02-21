package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

// DashboardKPIs representa los indicadores clave de rendimiento del dashboard
type DashboardKPIs struct {
	TotalProducts      int     `json:"total_products"`
	LowStockProducts   int     `json:"low_stock_products"`
	CurrentMonthSales  float64 `json:"current_month_sales"`
	PendingSalesOrders int     `json:"pending_sales_orders"`
}

// TopSellingProduct representa un producto más vendido
type TopSellingProduct struct {
	ProductName string `json:"product_name"`
	TotalSold   int    `json:"total_sold"`
}

// SalesEvolutionPoint representa un punto en el gráfico de evolución de ventas
type SalesEvolutionPoint struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

// MonthlyRevenue representa los ingresos de un mes
type MonthlyRevenue struct {
	Month     string  `json:"month"`
	Sales     float64 `json:"sales"`
	Purchases float64 `json:"purchases"`
}

// LowStockProduct representa un producto con stock bajo
type LowStockProduct struct {
	ProductID    int64  `json:"product_id"`
	ProductName  string `json:"product_name"`
	CurrentStock int    `json:"current_stock"`
	MinStock     int    `json:"min_stock"`
}

// ChartData contiene todos los datos para los gráficos del dashboard
type ChartData struct {
	TopSellingProducts []TopSellingProduct   `json:"top_selling_products"`
	SalesEvolution     []SalesEvolutionPoint `json:"sales_evolution"`
	MonthlyRevenue     []MonthlyRevenue      `json:"monthly_revenue"`
	LowStockProducts   []LowStockProduct     `json:"low_stock_products"`
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
	// Ahora calcula desde product_batches
	if err := m.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT p.id
			FROM products p
			LEFT JOIN product_batches pb ON p.id = pb.product_id
			WHERE p.user_id = $1
			GROUP BY p.id
			HAVING COALESCE(SUM(pb.quantity), 0) <= 5
		) AS low_stock_products
	`, userID).Scan(&metrics.ProductsLowStock); err != nil {
		// Si no hay productos con bajo stock, QueryRow retorna ErrNoRows
		if !errors.Is(err, pgx.ErrNoRows) {
			return DashboardMetrics{}, err
		}
		metrics.ProductsLowStock = 0
	}

	return metrics, nil
}

// GetDashboardKPIs obtiene los KPIs principales del dashboard
func (m *DashboardModel) GetDashboardKPIs(userID int64) (*DashboardKPIs, error) {
	kpis := &DashboardKPIs{}
	ctx := context.Background()

	// Total de productos
	err := m.DB.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM products 
		WHERE user_id = $1
	`, userID).Scan(&kpis.TotalProducts)
	if err != nil {
		return nil, err
	}

	// Productos con stock bajo (menor o igual a 5 o menor a stock_minimo si está definido)
	err = m.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT p.id,
				   COALESCE(SUM(pb.quantity), 0) as current_stock,
				   COALESCE(p.stock_minimo, 5) as min_stock
			FROM products p
			LEFT JOIN product_batches pb ON p.id = pb.product_id
			WHERE p.user_id = $1
			GROUP BY p.id, p.stock_minimo
			HAVING COALESCE(SUM(pb.quantity), 0) <= COALESCE(p.stock_minimo, 5)
		) AS low_stock_count
	`, userID).Scan(&kpis.LowStockProducts)
	if err != nil {
		return nil, err
	}

	// Ventas del mes actual
	err = m.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(oi.quantity * oi.unit_price), 0) 
		FROM sales_orders so 
		JOIN order_items oi ON so.id = oi.order_id 
		WHERE so.user_id = $1 
		AND so.order_date >= date_trunc('month', CURRENT_DATE)
	`, userID).Scan(&kpis.CurrentMonthSales)
	if err != nil {
		return nil, err
	}

	// Órdenes de venta pendientes
	err = m.DB.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM sales_orders 
		WHERE user_id = $1 AND status = 'pending'
	`, userID).Scan(&kpis.PendingSalesOrders)
	if err != nil {
		return nil, err
	}

	return kpis, nil
}

// GetChartData obtiene los datos para los gráficos del dashboard
func (m *DashboardModel) GetChartData(userID int64) (*ChartData, error) {
	data := &ChartData{
		TopSellingProducts: []TopSellingProduct{},
		SalesEvolution:     []SalesEvolutionPoint{},
		MonthlyRevenue:     []MonthlyRevenue{},
		LowStockProducts:   []LowStockProduct{},
	}
	ctx := context.Background()

	// Top 5 productos más vendidos
	rows, err := m.DB.Query(ctx, `
		SELECT 
			p.name,
			COALESCE(SUM(oi.quantity), 0) as total_sold
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN sales_orders so ON oi.order_id = so.id AND so.user_id = $1
		WHERE p.user_id = $1
		GROUP BY p.id, p.name
		ORDER BY total_sold DESC
		LIMIT 5
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product TopSellingProduct
		if err := rows.Scan(&product.ProductName, &product.TotalSold); err != nil {
			return nil, err
		}
		data.TopSellingProducts = append(data.TopSellingProducts, product)
	}

	// Evolución de ventas (últimos 30 días)
	rows, err = m.DB.Query(ctx, `
		WITH date_series AS (
			SELECT generate_series(
				CURRENT_DATE - INTERVAL '29 days',
				CURRENT_DATE,
				'1 day'::interval
			)::date AS date
		)
		SELECT 
			ds.date::text,
			COALESCE(SUM(oi.quantity * oi.unit_price), 0) as total
		FROM date_series ds
		LEFT JOIN sales_orders so ON DATE(so.order_date) = ds.date AND so.user_id = $1
		LEFT JOIN order_items oi ON so.id = oi.order_id
		GROUP BY ds.date
		ORDER BY ds.date
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var point SalesEvolutionPoint
		if err := rows.Scan(&point.Date, &point.Total); err != nil {
			return nil, err
		}
		data.SalesEvolution = append(data.SalesEvolution, point)
	}

	// Ingresos mensuales (últimos 6 meses) - Ventas vs Compras
	rows, err = m.DB.Query(ctx, `
		WITH months AS (
			SELECT generate_series(
				date_trunc('month', CURRENT_DATE - INTERVAL '5 months'),
				date_trunc('month', CURRENT_DATE),
				'1 month'::interval
			) AS month
		),
		sales_by_month AS (
			SELECT 
				date_trunc('month', so.order_date) as month,
				SUM(oi.quantity * oi.unit_price) as total_sales
			FROM sales_orders so
			JOIN order_items oi ON so.id = oi.order_id
			WHERE so.user_id = $1
			GROUP BY date_trunc('month', so.order_date)
		),
		purchases_by_month AS (
			SELECT 
				date_trunc('month', po.order_date) as month,
				SUM(poi.quantity * poi.unit_cost) as total_purchases
			FROM purchase_orders po
			JOIN purchase_order_items poi ON po.id = poi.purchase_order_id
			WHERE po.user_id = $1
			GROUP BY date_trunc('month', po.order_date)
		)
		SELECT 
			TO_CHAR(m.month, 'YYYY-MM') as month,
			COALESCE(s.total_sales, 0) as sales,
			COALESCE(p.total_purchases, 0) as purchases
		FROM months m
		LEFT JOIN sales_by_month s ON s.month = m.month
		LEFT JOIN purchases_by_month p ON p.month = m.month
		ORDER BY m.month
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var revenue MonthlyRevenue
		if err := rows.Scan(&revenue.Month, &revenue.Sales, &revenue.Purchases); err != nil {
			return nil, err
		}
		data.MonthlyRevenue = append(data.MonthlyRevenue, revenue)
	}

	// Productos con stock bajo (lista detallada con nombres)
	rows, err = m.DB.Query(ctx, `
		SELECT 
			p.id,
			p.name,
			COALESCE(SUM(pb.quantity), 0) as current_stock,
			COALESCE(p.stock_minimo, 5) as min_stock
		FROM products p
		LEFT JOIN product_batches pb ON p.id = pb.product_id
		WHERE p.user_id = $1
		GROUP BY p.id, p.name, p.stock_minimo
		HAVING COALESCE(SUM(pb.quantity), 0) <= COALESCE(p.stock_minimo, 5)
		ORDER BY current_stock ASC
		LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product LowStockProduct
		if err := rows.Scan(&product.ProductID, &product.ProductName, &product.CurrentStock, &product.MinStock); err != nil {
			return nil, err
		}
		data.LowStockProducts = append(data.LowStockProducts, product)
	}

	return data, nil
}
