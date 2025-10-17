export interface SalesOrder {
  id: number
  customer_id: number | null
  customer_name?: string
  order_date: string
  status: string
  total_amount: number | null
  user_id: number
}

export interface OrderItem {
  id: number
  order_id: number
  product_id: number
  quantity: number
  unit_price: number
}

export interface OrderItemInput {
  productId: number
  quantity: number
}
