package alerts

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/worker/internal/email"
)

// ProductAlert representa un producto con stock bajo
type ProductAlert struct {
	ID          int64
	Name        string
	Quantity    int
	StockMinimo int
	UserEmail   string
}

// CheckStockLevels chequea todos los productos con stock bajo y envía alertas
func CheckStockLevels(db *pgxpool.Pool, emailClient *email.Client) error {
	log.Println("🔍 Chequeando niveles de stock...")

	// Consulta SQL para obtener productos con stock bajo que no han sido notificados
	query := `
		SELECT p.id, p.name, p.quantity, p.stock_minimo, u.email
		FROM products p
		JOIN users u ON p.user_id = u.id
		WHERE p.quantity <= p.stock_minimo 
		  AND p.notificado = false
		ORDER BY p.quantity ASC
	`

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("error al ejecutar query de stock alerts: %w", err)
	}
	defer rows.Close()

	var alerts []ProductAlert
	for rows.Next() {
		var alert ProductAlert
		err := rows.Scan(&alert.ID, &alert.Name, &alert.Quantity, &alert.StockMinimo, &alert.UserEmail)
		if err != nil {
			log.Printf("❌ Error al escanear fila: %v", err)
			continue
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error al iterar sobre resultados: %w", err)
	}

	if len(alerts) == 0 {
		log.Println("✅ No hay productos con stock bajo. Todo está bajo control.")
		return nil
	}

	log.Printf("⚠️  Encontrados %d productos con stock bajo", len(alerts))

	// Procesar cada alerta
	for _, alert := range alerts {
		log.Printf("📧 Enviando alerta para producto: %s (Stock: %d/%d) a %s",
			alert.Name, alert.Quantity, alert.StockMinimo, alert.UserEmail)

		// Enviar el email de alerta
		if err := emailClient.SendStockAlertEmail(alert.UserEmail, alert.Name, alert.Quantity, alert.StockMinimo); err != nil {
			log.Printf("❌ Error al enviar email para producto ID %d: %v", alert.ID, err)
			continue
		}

		// Marcar el producto como notificado
		if err := markAsNotified(db, alert.ID); err != nil {
			log.Printf("❌ Error al marcar producto ID %d como notificado: %v", alert.ID, err)
			continue
		}

		log.Printf("✅ Alerta enviada y marcada como notificada para: %s", alert.Name)
	}

	log.Printf("🎉 Proceso de alertas completado. %d alertas enviadas.", len(alerts))
	return nil
}

// markAsNotified marca un producto como notificado en la base de datos
func markAsNotified(db *pgxpool.Pool, productID int64) error {
	query := `UPDATE products SET notificado = true WHERE id = $1`

	_, err := db.Exec(context.Background(), query, productID)
	if err != nil {
		return fmt.Errorf("error al actualizar producto: %w", err)
	}

	return nil
}
