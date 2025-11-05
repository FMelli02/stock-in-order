package models

import "time"

// ProductBatch representa un lote de productos con fecha de vencimiento
type ProductBatch struct {
	ID         int64      `json:"id"`
	ProductID  int        `json:"product_id"`
	UserID     int        `json:"user_id"`
	LoteNumber *string    `json:"lote_number"` // Número de lote (puede ser NULL)
	Quantity   int        `json:"quantity"`    // Cantidad en este lote
	ExpiryDate *time.Time `json:"expiry_date"` // Fecha de vencimiento (puede ser NULL)
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateProductBatchRequest representa la petición para crear un nuevo lote
type CreateProductBatchRequest struct {
	ProductID  int        `json:"product_id" validate:"required,gt=0"`
	LoteNumber *string    `json:"lote_number"`
	Quantity   int        `json:"quantity" validate:"required,gt=0"`
	ExpiryDate *time.Time `json:"expiry_date"`
}

// ProductWithBatches representa un producto con su información de lotes
type ProductWithBatches struct {
	Product
	Batches           []ProductBatch `json:"batches,omitempty"`
	TotalQuantity     int            `json:"total_quantity"`      // Suma de todos los lotes
	ExpiringSoon      int            `json:"expiring_soon"`       // Cantidad que vence en próximos 30 días
	HasExpiredBatches bool           `json:"has_expired_batches"` // Si tiene lotes vencidos
}
