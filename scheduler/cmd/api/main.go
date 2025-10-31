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

	// Declarar la cola de reportes (asegurarse de que existe)
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
		log.Fatalf("❌ No se pudo declarar la cola de reportes: %v", err)
	}

	log.Printf("✅ Conectado a RabbitMQ, cola de reportes: %s", queueName)

	// Declarar la cola de alertas de stock
	stockAlertsQueue := "stock_alerts_queue"
	_, err = ch.QueueDeclare(
		stockAlertsQueue, // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("❌ No se pudo declarar la cola de stock alerts: %v", err)
	}

	log.Printf("✅ Cola de alertas de stock declarada: %s", stockAlertsQueue)

	// Crear el scheduler de cron
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))

	// Crear el job de reportes semanales
	weeklyJob := jobs.NewWeeklyReportsJob(ch)

	// Crear el job de alertas de stock
	stockAlertsJob := jobs.NewStockAlertsJob(ch)

	// Programar el job de reportes semanales
	// Cron expression: "*/5 * * * *" = cada 5 minutos (para testing)
	// Para producción: "0 9 * * MON" = cada lunes a las 9:00 AM
	cronExpression := "*/5 * * * *" // TESTING: cada 5 minutos
	// cronExpression := "0 9 * * MON" // PRODUCCION: cada lunes a las 9 AM

	_, err = c.AddFunc(cronExpression, weeklyJob.Execute)
	if err != nil {
		log.Fatalf("❌ Error al agregar job de reportes al scheduler: %v", err)
	}

	log.Printf("📅 Job de reportes semanales programado con expresión cron: %s", cronExpression)

	// Programar el job de alertas de stock (cada hora)
	stockAlertsCron := "0 * * * *" // Cada hora en punto
	// stockAlertsCron := "*/2 * * * *" // TESTING: cada 2 minutos

	_, err = c.AddFunc(stockAlertsCron, stockAlertsJob.Execute)
	if err != nil {
		log.Fatalf("❌ Error al agregar job de stock alerts al scheduler: %v", err)
	}

	log.Printf("👁️  Job de alertas de stock programado con expresión cron: %s", stockAlertsCron)
	log.Println("🚀 Scheduler iniciado. Esperando próxima ejecución...")

	// Iniciar el scheduler
	c.Start()

	// Mantener el programa corriendo
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Ejecutar los jobs inmediatamente al iniciar (opcional, para testing)
	log.Println("🔥 Ejecutando job inicial de reportes...")
	weeklyJob.Execute()

	log.Println("🔥 Ejecutando job inicial de stock alerts...")
	stockAlertsJob.Execute()

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
