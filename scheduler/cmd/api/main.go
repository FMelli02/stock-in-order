package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron/v3"

	"stock-in-order/scheduler/internal/config"
	"stock-in-order/scheduler/internal/jobs"
)

func main() {
	log.Println("⏰ Iniciando Scheduler Service...")

	// Cargar configuración
	cfg := config.LoadConfig()
	log.Printf("📝 Configuración cargada: RabbitMQ_URL=%s", maskConnectionString(cfg.RabbitMQ_URL))

	// Conectar a RabbitMQ
	conn, err := amqp.Dial(cfg.RabbitMQ_URL)
	if err != nil {
		log.Fatalf("❌ No se pudo conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Crear canal
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ No se pudo crear canal de RabbitMQ: %v", err)
	}
	defer ch.Close()

	// Declarar la cola (asegurarse de que existe)
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
		log.Fatalf("❌ No se pudo declarar la cola: %v", err)
	}

	log.Printf("✅ Conectado a RabbitMQ, cola: %s", queueName)

	// Crear el scheduler de cron
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))

	// Crear el job de reportes semanales
	weeklyJob := jobs.NewWeeklyReportsJob(ch)

	// Programar el job
	// Cron expression: "*/5 * * * *" = cada 5 minutos (para testing)
	// Para producción: "0 9 * * MON" = cada lunes a las 9:00 AM
	cronExpression := "*/5 * * * *" // TESTING: cada 5 minutos
	// cronExpression := "0 9 * * MON" // PRODUCCION: cada lunes a las 9 AM

	_, err = c.AddFunc(cronExpression, weeklyJob.Execute)
	if err != nil {
		log.Fatalf("❌ Error al agregar job al scheduler: %v", err)
	}

	log.Printf("📅 Job programado con expresión cron: %s", cronExpression)
	log.Println("🚀 Scheduler iniciado. Esperando próxima ejecución...")

	// Iniciar el scheduler
	c.Start()

	// Mantener el programa corriendo
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Ejecutar el job inmediatamente al iniciar (opcional, para testing)
	log.Println("🔥 Ejecutando job inicial...")
	weeklyJob.Execute()

	// Bloquear hasta recibir señal de terminación
	sig := <-sigChan
	log.Printf("🛑 Señal recibida: %v. Deteniendo scheduler...", sig)

	// Detener el scheduler
	ctx := c.Stop()
	<-ctx.Done()

	log.Println("👋 Scheduler Service finalizado")
}

// maskConnectionString oculta credenciales en los logs
func maskConnectionString(s string) string {
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return "***"
}
