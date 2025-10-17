export interface PurchaseOrder {
  id: number
  supplier_id: number | null
  supplier_name?: string
  order_date: string
  status: string
  user_id: number
}

export interface PurchaseOrderItem {
  id: number
  purchase_order_id: number
  product_id: number
  quantity: number
  unit_cost: number
}

export interface PurchaseOrderItemInput {
  productId: number
  quantity: number
  unitCost: number
}
