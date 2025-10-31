package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StockAlertRequest representa el mensaje que se enviará a la cola de alertas
type StockAlertRequest struct {
	TaskType string `json:"task_type"`
}

// StockAlertsJob es el job que chequea alertas de stock bajo cada hora
type StockAlertsJob struct {
	channel *amqp.Channel
}

// NewStockAlertsJob crea una nueva instancia del job
func NewStockAlertsJob(ch *amqp.Channel) *StockAlertsJob {
	return &StockAlertsJob{
		channel: ch,
	}
}

// Execute se ejecuta cuando el cron dispara la tarea
func (j *StockAlertsJob) Execute() {
	log.Println("👁️  [SCHEDULER] Ejecutando job de alertas de stock...")

	// Crear el mensaje de chequeo de stock
	req := StockAlertRequest{
		TaskType: "check_stock_levels",
	}

	// Publicar el mensaje a la cola
	if err := j.publishStockAlert(req); err != nil {
		log.Printf("❌ Error al publicar tarea de stock alert: %v", err)
		return
	}

	log.Println("✅ Tarea de stock alerts enviada a la cola")
	log.Println("🎉 [SCHEDULER] Job de alertas de stock completado")
}

// publishStockAlert publica un mensaje en la cola de alertas de stock
func (j *StockAlertsJob) publishStockAlert(req StockAlertRequest) error {
	queueName := "stock_alerts_queue"

	// Serializar el mensaje a JSON
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publicar mensaje
	err = j.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
