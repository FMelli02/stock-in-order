package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ReportRequest representa el mensaje que se enviará a la cola
type ReportRequest struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email_to"`
	ReportType string `json:"report_type"`
}

// User representa un usuario en la base de datos
type User struct {
	ID    int64
	Email string
}

// WeeklyReportsJob es el job que envía reportes semanales programados
type WeeklyReportsJob struct {
	channel *amqp.Channel
	db      *pgxpool.Pool
}

// NewWeeklyReportsJob crea una nueva instancia del job
func NewWeeklyReportsJob(ch *amqp.Channel, db *pgxpool.Pool) *WeeklyReportsJob {
	return &WeeklyReportsJob{
		channel: ch,
		db:      db,
	}
}

// Execute se ejecuta cuando el cron dispara la tarea
func (j *WeeklyReportsJob) Execute() {
	log.Println("⏰ [SCHEDULER] Ejecutando job de reportes semanales...")

	// Obtener todos los usuarios activos de la base de datos
	users, err := j.fetchActiveUsers()
	if err != nil {
		log.Printf("❌ Error al obtener usuarios: %v", err)
		return
	}

	if len(users) == 0 {
		log.Println("⚠️ No hay usuarios activos para enviar reportes")
		return
	}

	log.Printf("📊 Generando reportes semanales para %d usuarios", len(users))

	// Para cada usuario, enviar los 3 tipos de reportes
	reportTypes := []string{"products_weekly", "customers_weekly", "suppliers_weekly"}

	for _, user := range users {
		for _, reportType := range reportTypes {
			req := ReportRequest{
				UserID:     user.ID,
				Email:      user.Email,
				ReportType: reportType,
			}

			if err := j.publishReport(req); err != nil {
				log.Printf("❌ Error al publicar reporte %s para usuario %d: %v", reportType, user.ID, err)
				continue
			}
			log.Printf("✅ Reporte semanal enviado a la cola: %s para %s (UserID: %d)", reportType, user.Email, user.ID)
		}
	}

	log.Println("🎉 [SCHEDULER] Job de reportes semanales completado")
}

// fetchActiveUsers obtiene todos los usuarios activos de la base de datos
func (j *WeeklyReportsJob) fetchActiveUsers() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, email FROM users ORDER BY id`

	rows, err := j.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar usuarios: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email); err != nil {
			log.Printf("⚠️ Error al escanear usuario: %v", err)
			continue
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar usuarios: %w", err)
	}

	return users, nil
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
