# Worker Service - Stock In Order

## 🎯 Propósito

El **Worker Service** es un microservicio independiente que se encarga de procesar tareas pesadas en segundo plano, como la generación de reportes Excel y el envío de emails. Funciona como un "cadete" que escucha mensajes en RabbitMQ y los procesa sin bloquear la API principal.

## 🏗️ Arquitectura

```
API Backend (puerto 8080)
    │
    ├─ Publica mensajes → RabbitMQ (puerto 5672)
    │                          │
    │                          ├─ Cola: reporting_queue
    │                          │
    │                          ↓
    │                    Worker Service
    │                          │
    │                          ├─ Conecta a PostgreSQL (puerto 5432)
    │                          ├─ Genera reportes Excel
    │                          └─ Envía emails (próximamente)
```

## 📦 Estructura del Proyecto

```
worker/
├── cmd/
│   └── api/
│       └── main.go              # Punto de entrada del worker
├── internal/
│   ├── config/
│   │   └── config.go            # Configuración del worker
│   ├── consumer/
│   │   └── consumer.go          # Lógica del consumidor RabbitMQ
│   ├── models/
│   │   ├── product.go           # Modelo de productos
│   │   ├── customer.go          # Modelo de clientes
│   │   └── supplier.go          # Modelo de proveedores
│   └── reports/
│       └── generator.go         # Generadores de reportes Excel
├── Dockerfile                   # Configuración Docker
├── go.mod                       # Dependencias Go
└── README.md                    # Este archivo
```

## 🔧 Tecnologías Utilizadas

- **Go 1.23**: Lenguaje de programación
- **RabbitMQ (amqp091-go)**: Cliente AMQP para consumir mensajes
- **PostgreSQL (pgx/v5)**: Base de datos
- **Excelize v2**: Generación de archivos Excel
- **Docker**: Contenedorización

## 📨 Formato de Mensajes

El worker escucha mensajes JSON en la cola `reporting_queue`:

```json
{
  "user_id": 1,
  "email_to": "usuario@ejemplo.com",
  "report_type": "products"
}
```

### Tipos de Reportes Soportados

- `products`: Reporte de productos
- `customers`: Reporte de clientes
- `suppliers`: Reporte de proveedores

## 🚀 Ejecución

### Con Docker Compose (Recomendado)

```bash
# Levantar todos los servicios
docker compose up -d

# Ver logs del worker
docker logs stock_in_order_worker -f

# Reiniciar solo el worker
docker compose restart worker
```

### Desarrollo Local

```bash
# Navegar a la carpeta del worker
cd worker

# Instalar dependencias
go mod download

# Configurar variables de entorno
export DB_DSN="postgres://user:pass@localhost:5433/stock_db?sslmode=disable"
export RABBITMQ_URL="amqp://user:pass@localhost:5672/"

# Ejecutar el worker
go run cmd/api/main.go
```

## 🔒 Variables de Entorno

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `DB_DSN` | String de conexión a PostgreSQL | `postgres://user:pass@postgres_db:5432/stock_db?sslmode=disable` |
| `RABBITMQ_URL` | URL de conexión a RabbitMQ | `amqp://user:pass@rabbitmq:5672/` |
| `SENDGRID_API_KEY` | API Key de SendGrid (opcional) | `SG.xxxxxx` |

## 📊 Monitoreo

### Ver Estado del Worker

```bash
# Estado del contenedor
docker compose ps worker

# Logs en tiempo real
docker logs stock_in_order_worker -f

# Últimas 50 líneas de logs
docker logs stock_in_order_worker --tail 50
```

### Monitorear RabbitMQ

Accede a la interfaz web de RabbitMQ:

```
URL: http://localhost:15672
Usuario: user
Contraseña: pass
```

Desde ahí puedes:
- Ver mensajes en la cola `reporting_queue`
- Monitorear el consumo de mensajes
- Ver conexiones activas del worker

## 🧪 Probar el Worker

Para probar el worker, puedes enviar un mensaje manualmente a RabbitMQ usando la interfaz web o un script:

### Opción 1: Interfaz Web de RabbitMQ

1. Accede a http://localhost:15672
2. Ve a la pestaña **Queues**
3. Haz clic en `reporting_queue`
4. En **Publish message**, pega:

```json
{
  "user_id": 1,
  "email_to": "test@ejemplo.com",
  "report_type": "products"
}
```

5. Haz clic en **Publish message**
6. Revisa los logs del worker: `docker logs stock_in_order_worker -f`

### Opción 2: Script de Go (próximamente)

El backend API tendrá endpoints que publiquen mensajes automáticamente.

## 🐛 Troubleshooting

### El worker no arranca

```bash
# Ver logs de error
docker logs stock_in_order_worker

# Verificar que RabbitMQ esté healthy
docker compose ps rabbitmq

# Verificar que PostgreSQL esté healthy
docker compose ps postgres_db
```

### El worker no procesa mensajes

```bash
# Verificar que la cola exista
# Ve a http://localhost:15672 → Queues

# Verificar conexiones activas
# Ve a http://localhost:15672 → Connections

# Reiniciar el worker
docker compose restart worker
```

### Errores de base de datos

```bash
# Verificar conectividad
docker exec stock_in_order_worker ping postgres_db

# Probar conexión manual
docker exec -it stock_in_order_worker sh
# Dentro del contenedor:
apk add --no-cache postgresql-client
psql -h postgres_db -U user -d stock_db
```

## 🔄 Flujo de Procesamiento

1. **API recibe petición**: Usuario solicita generar un reporte
2. **API publica mensaje**: Se envía un mensaje JSON a `reporting_queue`
3. **Worker consume mensaje**: El worker recibe el mensaje de la cola
4. **Genera reporte**: Se consulta la base de datos y se genera el Excel
5. **Envía por email**: (Próximamente) Se envía el archivo por email usando SendGrid
6. **ACK mensaje**: El worker confirma que procesó el mensaje exitosamente

## 📝 TODOs

- [ ] Implementar envío de emails con SendGrid
- [ ] Agregar soporte para reportes de ventas y compras
- [ ] Implementar reintentos con exponential backoff
- [ ] Agregar Dead Letter Queue (DLQ) para mensajes fallidos
- [ ] Implementar almacenamiento de reportes en S3 o similar
- [ ] Agregar métricas y observabilidad (Prometheus/Grafana)
- [ ] Implementar notificaciones cuando el reporte esté listo

## 🎓 Conceptos Clave

### ¿Por qué un Worker Service?

- **No bloquea la API**: La generación de reportes puede tomar varios segundos
- **Escalabilidad**: Puedes levantar múltiples workers para procesar más mensajes
- **Resiliencia**: Si el worker falla, el mensaje se reencola automáticamente
- **Separación de responsabilidades**: La API solo se encarga de recibir peticiones

### ¿Qué es AMQP?

**Advanced Message Queuing Protocol** es un protocolo para comunicación entre aplicaciones mediante colas de mensajes. RabbitMQ implementa AMQP.

### ¿Qué es un Consumer?

Un consumer es un proceso que escucha una cola y procesa los mensajes que llegan. En este caso, nuestro worker es el consumer.

## 📚 Referencias

- [RabbitMQ Go Tutorial](https://www.rabbitmq.com/tutorials/tutorial-one-go)
- [amqp091-go Documentation](https://pkg.go.dev/github.com/rabbitmq/amqp091-go)
- [Excelize Documentation](https://xuri.me/excelize/)
- [PostgreSQL pgx Driver](https://github.com/jackc/pgx)

---

**Autor**: Stock In Order Team  
**Última actualización**: 21 de Octubre 2025
