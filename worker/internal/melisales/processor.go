package melisales

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/worker/internal/models"
	"stock-in-order/worker/internal/services"
)

// MercadoLibreNotification representa la notificación recibida del webhook
type MercadoLibreNotification struct {
	ID            int64  `json:"_id"`
	Resource      string `json:"resource"` // "/orders/123456789"
	UserID        int64  `json:"user_id"`  // ML user ID
	Topic         string `json:"topic"`    // "orders_v2"
	ApplicationID int64  `json:"application_id"`
	Attempts      int    `json:"attempts"`
	Sent          string `json:"sent"`
	Received      string `json:"received"`
}

// SalesOrder representa una orden de venta en nuestro sistema
type SalesOrder struct {
	ID           int64           `json:"id"`
	CustomerID   sql.NullInt64   `json:"customer_id"`
	CustomerName string          `json:"customer_name,omitempty"`
	OrderDate    time.Time       `json:"order_date"`
	Status       string          `json:"status"`
	TotalAmount  sql.NullFloat64 `json:"total_amount"`
	UserID       int64           `json:"user_id"`
}

// OrderItem representa un item de la orden
type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// ProcessSale procesa una venta de Mercado Libre
func ProcessSale(db *pgxpool.Pool, mlService *services.MercadoLibreService, integrationModel *models.IntegrationModel, notificationBody []byte, encryptionKey string) error {
	// 1. Parsear la notificación
	var notification MercadoLibreNotification
	if err := json.Unmarshal(notificationBody, &notification); err != nil {
		return fmt.Errorf("error parsing notification: %w", err)
	}

	log.Printf("🔔 Procesando venta de Mercado Libre - Order ID: %s, User ID: %d", notification.Resource, notification.UserID)

	// 2. Extraer el order ID del resource (ej: "/orders/123456789")
	orderID, err := extractOrderID(notification.Resource)
	if err != nil {
		return fmt.Errorf("error extracting order ID: %w", err)
	}

	// 3. Buscar la integración usando el external_user_id
	externalUserIDStr := strconv.FormatInt(notification.UserID, 10)
	integration, err := integrationModel.GetByExternalUserID(externalUserIDStr, "mercadolibre")
	if err != nil {
		return fmt.Errorf("error getting integration for user %d: %w", notification.UserID, err)
	}

	log.Printf("✅ Integración encontrada - UserID interno: %d", integration.UserID)

	// 4. Obtener la integración completa con tokens (desencriptados)
	integrationModel.EncryptionKey = encryptionKey
	integrationWithTokens, err := integrationModel.GetByUserAndPlatform(integration.UserID, "mercadolibre")
	if err != nil {
		return fmt.Errorf("error getting integration with tokens: %w", err)
	}

	// 5. Obtener los detalles de la orden desde Mercado Libre
	mlOrder, err := mlService.GetOrder(orderID, integrationWithTokens.AccessToken)
	if err != nil {
		return fmt.Errorf("error getting order from Mercado Libre: %w", err)
	}

	log.Printf("📦 Orden obtenida de Mercado Libre - ID: %d, Status: %s, Total: %.2f %s",
		mlOrder.ID, mlOrder.Status, mlOrder.TotalAmount, mlOrder.CurrencyID)

	// 6. Verificar que la orden esté confirmada/pagada
	if !isOrderValid(mlOrder.Status) {
		log.Printf("⚠️  Orden con status %s - no se procesará", mlOrder.Status)
		return nil // No es un error, solo no procesamos órdenes no válidas
	}

	// 7. Mapear los items de la orden a productos en nuestro sistema
	orderItems, err := mapOrderItems(db, mlOrder, integration.UserID)
	if err != nil {
		return fmt.Errorf("error mapping order items: %w", err)
	}

	if len(orderItems) == 0 {
		log.Printf("⚠️  No se encontraron productos coincidentes para la orden %d", orderID)
		return fmt.Errorf("no matching products found for order %d", orderID)
	}

	// 8. Crear la orden de venta en nuestro sistema
	salesOrder := &SalesOrder{
		CustomerID:   sql.NullInt64{Valid: false}, // No tenemos customer ID de Mercado Libre
		CustomerName: fmt.Sprintf("%s %s (%s)", mlOrder.Buyer.FirstName, mlOrder.Buyer.LastName, mlOrder.Buyer.Nickname),
		OrderDate:    time.Now(),
		Status:       "completed",
		TotalAmount:  sql.NullFloat64{Float64: mlOrder.TotalAmount, Valid: true},
		UserID:       integration.UserID,
	}

	if err := createSalesOrder(db, salesOrder, orderItems); err != nil {
		return fmt.Errorf("error creating sales order: %w", err)
	}

	log.Printf("✅ Orden de venta creada exitosamente - ID: %d, Total: %.2f", salesOrder.ID, mlOrder.TotalAmount)

	return nil
}

// extractOrderID extrae el ID de orden del resource path
func extractOrderID(resource string) (int64, error) {
	// Resource viene como "/orders/123456789"
	parts := strings.Split(resource, "/")
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid resource format: %s", resource)
	}

	orderIDStr := parts[2]
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid order ID: %s", orderIDStr)
	}

	return orderID, nil
}

// isOrderValid verifica si una orden debe ser procesada
func isOrderValid(status string) bool {
	// Solo procesamos órdenes confirmadas o pagadas
	validStatuses := []string{"confirmed", "paid", "partially_paid"}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// mapOrderItems mapea los items de Mercado Libre a productos en nuestro sistema
func mapOrderItems(db *pgxpool.Pool, mlOrder *services.MLOrder, userID int64) ([]OrderItem, error) {
	ctx := context.Background()
	var items []OrderItem

	for _, mlItem := range mlOrder.OrderItems {
		// Buscar producto por SKU
		sku := mlItem.Item.SellerSKU
		if sku == "" {
			log.Printf("⚠️  Item sin SKU - ID: %s, Title: %s", mlItem.Item.ID, mlItem.Item.Title)
			continue
		}

		var productID int64
		query := `SELECT id FROM products WHERE sku = $1 AND user_id = $2`
		err := db.QueryRow(ctx, query, sku, userID).Scan(&productID)
		if err != nil {
			log.Printf("⚠️  Producto no encontrado - SKU: %s, Error: %v", sku, err)
			continue
		}

		items = append(items, OrderItem{
			ProductID: productID,
			Quantity:  mlItem.Quantity,
			UnitPrice: mlItem.UnitPrice,
		})

		log.Printf("✅ Item mapeado - SKU: %s → ProductID: %d, Quantity: %d", sku, productID, mlItem.Quantity)
	}

	return items, nil
}

// createSalesOrder crea una orden de venta en la base de datos
func createSalesOrder(db *pgxpool.Pool, order *SalesOrder, items []OrderItem) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Insertar el header de la orden
	const insertOrder = `
		INSERT INTO sales_orders (customer_id, order_date, status, total_amount, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, order_date`

	if err := tx.QueryRow(ctx, insertOrder,
		order.CustomerID, order.OrderDate, order.Status, order.TotalAmount, order.UserID,
	).Scan(&order.ID, &order.OrderDate); err != nil {
		return err
	}

	// Insertar items y actualizar stock
	const insertItem = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	const updateStock = `
		UPDATE products SET quantity = quantity - $1
		WHERE id = $2 AND user_id = $3 AND quantity - $1 >= 0`

	for i := range items {
		items[i].OrderID = order.ID

		// Insertar item
		if err := tx.QueryRow(ctx, insertItem,
			items[i].OrderID, items[i].ProductID, items[i].Quantity, items[i].UnitPrice,
		).Scan(&items[i].ID); err != nil {
			return err
		}

		// Actualizar stock (asegurar que no sea negativo)
		tag, err := tx.Exec(ctx, updateStock, items[i].Quantity, items[i].ProductID, order.UserID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("insufficient stock for product %d", items[i].ProductID)
		}

		// Insertar movimiento de stock
		const insertMovement = `
			INSERT INTO stock_movements (product_id, movement_type, quantity, reference_id, user_id)
			VALUES ($1, 'sale', $2, $3, $4)`

		_, err = tx.Exec(ctx, insertMovement, items[i].ProductID, -items[i].Quantity, order.ID, order.UserID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
