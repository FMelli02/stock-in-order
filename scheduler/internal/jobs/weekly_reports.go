package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ReportRequest representa el mensaje que se enviará a la cola
type ReportRequest struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email_to"`
	ReportType string `json:"report_type"`
}

// WeeklyReportsJob es el job que envía reportes semanales programados
type WeeklyReportsJob struct {
	channel *amqp.Channel
}

// NewWeeklyReportsJob crea una nueva instancia del job
func NewWeeklyReportsJob(ch *amqp.Channel) *WeeklyReportsJob {
	return &WeeklyReportsJob{
		channel: ch,
	}
}

// Execute se ejecuta cuando el cron dispara la tarea
func (j *WeeklyReportsJob) Execute() {
	log.Println("⏰ [SCHEDULER] Ejecutando job de reportes semanales...")

	// Lista de reportes a generar semanalmente
	reports := []ReportRequest{
		{
			UserID:     1,
			Email:      "admin@stockinorder.com",
			ReportType: "products_weekly",
		},
		{
			UserID:     1,
			Email:      "admin@stockinorder.com",
			ReportType: "customers_weekly",
		},
		{
			UserID:     1,
			Email:      "admin@stockinorder.com",
			ReportType: "suppliers_weekly",
		},
	}

	// Enviar cada reporte a la cola
	for _, req := range reports {
		if err := j.publishReport(req); err != nil {
			log.Printf("❌ Error al publicar reporte %s: %v", req.ReportType, err)
			continue
		}
		log.Printf("✅ Reporte semanal enviado a la cola: %s para %s", req.ReportType, req.Email)
	}

	log.Println("🎉 [SCHEDULER] Job de reportes semanales completado")
}

// publishReport publica un mensaje en la cola de RabbitMQ
func (j *WeeklyReportsJob) publishReport(req ReportRequest) error {
	queueName := "reporting_queue"

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
