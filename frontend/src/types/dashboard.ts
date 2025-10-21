export interface DashboardMetrics {
  total_products: number
  total_customers: number
  total_suppliers: number
  pending_sales_orders: number
  products_low_stock: number
}

export interface DashboardKPIs {
  total_products: number
  low_stock_products: number
  current_month_sales: number
  pending_sales_orders: number
}

export interface TopSellingProduct {
  product_name: string
  total_sold: number
}

export interface SalesEvolutionPoint {
  date: string
  total: number
}

export interface ChartData {
  top_selling_products: TopSellingProduct[]
  sales_evolution: SalesEvolutionPoint[]
}
