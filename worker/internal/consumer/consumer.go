package consumer

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"

	"stock-in-order/worker/internal/email"
	"stock-in-order/worker/internal/reports"
)

// ReportRequest representa la estructura del mensaje JSON que llega desde la cola
type ReportRequest struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email_to"`
	ReportType string `json:"report_type"` // "products", "customers", "suppliers"
}

// StartConsumer inicia el consumidor que escucha la cola de RabbitMQ
func StartConsumer(rabbitURL string, db *pgxpool.Pool, emailClient *email.Client) error {
	// Conectar a RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	// Crear un canal
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Declarar la cola (si no existe, se crea)
	queueName := "reporting_queue"
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable (la cola sobrevive reinicios de RabbitMQ)
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	log.Printf("📬 Worker conectado a RabbitMQ. Escuchando cola: %s", q.Name)

	// Configurar QoS (prefetch): procesar 1 mensaje a la vez
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Registrar el consumidor
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer (empty = auto-generated)
		false,  // auto-ack (false = manual ack)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Canal para mantener el proceso vivo
	forever := make(chan bool)

	// Goroutine que procesa mensajes
	go func() {
		for d := range msgs {
			log.Printf("📨 Mensaje recibido: %s", d.Body)

			// Parsear el mensaje JSON
			var req ReportRequest
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Printf("❌ Error al parsear mensaje: %v", err)
				d.Nack(false, false) // Rechazar mensaje sin reencolar
				continue
			}

			// Procesar el reporte
			if err := processReport(db, emailClient, req); err != nil {
				log.Printf("❌ Error al procesar reporte: %v", err)
				d.Nack(false, true) // Rechazar y reencolar para reintentar
				continue
			}

			// Confirmar que el mensaje fue procesado exitosamente
			log.Printf("✅ Reporte procesado exitosamente para UserID=%d, ReportType=%s", req.UserID, req.ReportType)
			d.Ack(false)
		}
	}()

	log.Printf("🚀 Worker listo. Presiona CTRL+C para salir.")
	<-forever // Bloquear indefinidamente

	return nil
}

// processReport procesa una solicitud de reporte
func processReport(db *pgxpool.Pool, emailClient *email.Client, req ReportRequest) error {
	log.Printf("🔨 Generando reporte: UserID=%d, Email=%s, Type=%s", req.UserID, req.Email, req.ReportType)

	var reportBytes []byte
	var err error
	var filename string

	// Generar el reporte según el tipo
	switch req.ReportType {
	case "products":
		reportBytes, err = reports.GenerateProductsReport(db, req.UserID)
		filename = "reporte_productos.xlsx"
	case "products_weekly":
		reportBytes, err = reports.GenerateProductsReport(db, req.UserID)
		filename = "reporte_productos_semanal.xlsx"
	case "customers":
		reportBytes, err = reports.GenerateCustomersReport(db, req.UserID)
		filename = "reporte_clientes.xlsx"
	case "customers_weekly":
		reportBytes, err = reports.GenerateCustomersReport(db, req.UserID)
		filename = "reporte_clientes_semanal.xlsx"
	case "suppliers":
		reportBytes, err = reports.GenerateSuppliersReport(db, req.UserID)
		filename = "reporte_proveedores.xlsx"
	case "suppliers_weekly":
		reportBytes, err = reports.GenerateSuppliersReport(db, req.UserID)
		filename = "reporte_proveedores_semanal.xlsx"
	default:
		return fmt.Errorf("unknown report type: %s", req.ReportType)
	}

	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	log.Printf("📊 Reporte generado: %d bytes", len(reportBytes))

	// Enviar el reporte por email usando SendGrid
	attachment := email.EmailAttachment{
		Filename:    filename,
		Content:     reportBytes,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if err := emailClient.SendReportEmail(req.Email, "", req.ReportType, attachment); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("📧 Email enviado exitosamente a %s", req.Email)

	return nil
}
