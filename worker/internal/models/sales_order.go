package models

import "time"

// SalesOrderPDFRequest representa el mensaje para generar un PDF de orden de venta
type SalesOrderPDFRequest struct {
	OrderID int64  `json:"order_id"`
	UserID  int64  `json:"user_id"`
	Email   string `json:"email"`
}

// SalesOrderWithItems contiene todos los datos necesarios para el PDF
type SalesOrderWithItems struct {
	ID           int64          `json:"id"`
	CustomerName string         `json:"customer_name"`
	OrderDate    time.Time      `json:"order_date"`
	Status       string         `json:"status"`
	TotalAmount  float64        `json:"total_amount"`
	Items        []OrderItemPDF `json:"items"`
}

// OrderItemPDF representa un item de la orden para el PDF
type OrderItemPDF struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}
