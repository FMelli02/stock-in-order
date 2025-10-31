package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client encapsula la conexión y canal de RabbitMQ
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *slog.Logger
}

// ReportRequest representa un mensaje para generar un reporte
type ReportRequest struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email_to"`
	ReportType string `json:"report_type"`
}

// Connect establece conexión a RabbitMQ y retorna un cliente
func Connect(rabbitURL string, logger *slog.Logger) (*Client, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declarar la cola (idempotente - si ya existe, no hace nada)
	queueName := "reporting_queue"
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Info("Conectado a RabbitMQ", "queue", queueName)

	return &Client{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

// Close cierra la conexión a RabbitMQ
func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error("Error cerrando canal de RabbitMQ", "error", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("Error cerrando conexión a RabbitMQ", "error", err)
			return err
		}
	}
	return nil
}

// PublishReportRequest publica un mensaje de solicitud de reporte a la cola
func (c *Client) PublishReportRequest(ctx context.Context, req ReportRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",                // exchange (default)
		"reporting_queue", // routing key (queue name)
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	c.logger.Info("Mensaje publicado a RabbitMQ",
		"user_id", req.UserID,
		"email", req.Email,
		"report_type", req.ReportType,
	)

	return nil
}

// PublishMessage publica un mensaje genérico a una cola específica
func (c *Client) PublishMessage(ctx context.Context, queueName string, body []byte) error {
	// Declarar la cola si no existe
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	// Publicar el mensaje
	err = c.channel.PublishWithContext(
		ctx,
		"",        // exchange (default)
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to queue %s: %w", queueName, err)
	}

	c.logger.Info("Mensaje publicado a RabbitMQ",
		"queue", queueName,
		"size", len(body),
	)

	return nil
}
