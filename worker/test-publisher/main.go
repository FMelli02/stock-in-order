package main

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ReportRequest representa un mensaje para generar un reporte
type ReportRequest struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email_to"`
	ReportType string `json:"report_type"`
}

func main() {
	// Conectar a RabbitMQ
	rabbitURL := "amqp://user:pass@localhost:5672/"
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Crear canal
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declarar la cola
	queueName := "reporting_queue"
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Crear mensaje de prueba
	msg := ReportRequest{
		UserID:     1,
		Email:      "test@example.com",
		ReportType: "products",
	}

	body, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	// Publicar mensaje
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}

	fmt.Println("✅ Mensaje enviado a la cola:", queueName)
	fmt.Println("📨 Contenido:", string(body))
	fmt.Println("\n💡 Revisa los logs del worker con: docker logs stock_in_order_worker -f")
}
