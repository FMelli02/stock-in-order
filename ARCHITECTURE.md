# 🏗️ Arquitectura del Sistema - Stock In Order

## 📋 Visión General

Stock In Order es una aplicación de gestión de inventario con arquitectura de microservicios que utiliza procesamiento asíncrono para tareas pesadas.

## 🎭 Componentes del Sistema

```
┌─────────────────────────────────────────────────────────────────┐
│                         USUARIO                                 │
│                     (Navegador Web)                             │
└────────────┬────────────────────────────────────────────────────┘
             │
             │ HTTP
             │
     ┌───────▼────────┐
     │   Frontend     │ Puerto 5173
     │  (React + TS)  │ Nginx en Docker
     └───────┬────────┘
             │
             │ HTTP REST API
             │
     ┌───────▼────────┐
     │  Backend API   │ Puerto 8080
     │   (Go + Chi)   │
     └───┬────────┬───┘
         │        │
         │        │ AMQP
         │        │
         │    ┌───▼──────────┐
         │    │   RabbitMQ   │ Puertos 5672 (AMQP)
         │    │ (Message     │         15672 (Admin)
         │    │  Broker)     │
         │    └───┬──────────┘
         │        │
         │        │ Consume mensajes
         │        │
         │    ┌───▼──────────┐
         │    │Worker Service│ Sin puerto HTTP
         │    │  (Go + AMQP) │ Solo consumer
         │    └───┬──────────┘
         │        │
         │        │ Query data
         │        │
    ┌────▼────────▼────┐
    │   PostgreSQL     │ Puerto 5433 (externo)
    │  (Base de Datos) │        5432 (interno)
    └──────────────────┘
```

## 🧩 Servicios

### 1. Frontend (React + TypeScript)
- **Puerto**: 5173 (mapeado a 80 interno)
- **Tecnologías**: React 18, TypeScript, Vite, Recharts
- **Responsabilidades**:
  - Interfaz de usuario
  - Gestión de estado con Context API
  - Autenticación JWT
  - Dashboards y visualizaciones
  - Exportación de reportes (descarga Excel)

### 2. Backend API (Go + Chi)
- **Puerto**: 8080
- **Tecnologías**: Go 1.25, Chi Router, pgx/v5, JWT
- **Responsabilidades**:
  - REST API para el frontend
  - Autenticación y autorización
  - CRUD de entidades (productos, clientes, proveedores, órdenes)
  - **Publisher de mensajes** a RabbitMQ
  - Generación sincrónica de reportes simples

### 3. Worker Service (Go + AMQP)
- **Puerto**: Ninguno (no expone HTTP)
- **Tecnologías**: Go 1.23, amqp091-go, excelize/v2, pgx/v5
- **Responsabilidades**:
  - **Consumer de mensajes** de RabbitMQ
  - Generación de reportes Excel pesados
  - Envío de emails (próximamente con SendGrid)
  - Procesamiento asíncrono de tareas

### 4. RabbitMQ (Message Broker)
- **Puertos**: 
  - 5672: AMQP (comunicación entre apps)
  - 15672: Management UI (interfaz web)
- **Credenciales**: user / pass
- **Responsabilidades**:
  - Cola de mensajes para tareas asíncronas
  - Desacoplamiento entre API y Worker
  - Garantía de entrega de mensajes
  - Reintentos automáticos en caso de fallo

### 5. PostgreSQL (Base de Datos)
- **Puerto**: 5433 (mapeado a 5432 interno)
- **Credenciales**: user / pass
- **Base de datos**: stock_db
- **Responsabilidades**:
  - Almacenamiento persistente de datos
  - Transacciones ACID
  - Relaciones entre entidades

## 🔄 Flujos de Procesamiento

### Flujo 1: Operación Sincrónica Normal

```
Usuario → Frontend → API → PostgreSQL → API → Frontend → Usuario
                      ↓
                 Respuesta inmediata
```

**Ejemplo**: Crear un producto, obtener lista de clientes

### Flujo 2: Generación de Reporte Asíncrono

```
1. Usuario solicita reporte → Frontend → API
2. API publica mensaje → RabbitMQ (cola: reporting_queue)
3. API responde "Solicitud aceptada" → Frontend → Usuario
4. Worker consume mensaje ← RabbitMQ
5. Worker genera Excel ← PostgreSQL
6. Worker envía email → Usuario (próximamente)
```

**Ventajas**:
- No bloquea la API
- Escalable (múltiples workers)
- Resiliente (reintentos automáticos)
- Usuario puede seguir trabajando

### Flujo 3: Manejo de Errores

```
1. Worker procesa mensaje
2. Si falla:
   ├─ NACK al mensaje
   ├─ RabbitMQ reencola el mensaje
   └─ Otro worker (o el mismo) reintenta
3. Si éxito:
   └─ ACK al mensaje (confirmación)
```

## 📨 Estructura de Mensajes

### Cola: `reporting_queue`

**Mensaje JSON**:
```json
{
  "user_id": 1,
  "email_to": "usuario@ejemplo.com",
  "report_type": "products"
}
```

**Tipos de reportes**:
- `products`: Listado de productos
- `customers`: Listado de clientes
- `suppliers`: Listado de proveedores

## 🐳 Docker Compose

### Servicios Configurados

| Servicio | Contenedor | Puerto | Depende de |
|----------|-----------|--------|------------|
| postgres_db | stock_in_order_postgres | 5433:5432 | - |
| rabbitmq | stock_in_order_rabbitmq | 5672, 15672 | - |
| migrate | stock_in_order_migrate | - | postgres_db |
| api | stock_in_order_api | 8080:8080 | postgres_db, rabbitmq, migrate |
| worker | stock_in_order_worker | - | postgres_db, rabbitmq, migrate |
| frontend | stock_in_order_frontend | 5173:80 | api, postgres_db |

### Healthchecks

- **PostgreSQL**: `pg_isready -U user -d stock_db` cada 5s
- **RabbitMQ**: `rabbitmq-diagnostics ping` cada 10s

Todos los servicios dependientes esperan a que estén **healthy** antes de iniciar.

## 🔐 Variables de Entorno

### API Backend

```env
PORT=":8080"
DB_DSN="postgres://user:pass@postgres_db:5432/stock_db?sslmode=disable"
JWT_SECRET="dev-jwt-secret-change-me"
ENVIRONMENT="development"
RABBITMQ_URL="amqp://user:pass@rabbitmq:5672/"
SENTRY_DSN=""  # Opcional
```

### Worker Service

```env
DB_DSN="postgres://user:pass@postgres_db:5432/stock_db?sslmode=disable"
RABBITMQ_URL="amqp://user:pass@rabbitmq:5672/"
SENDGRID_API_KEY=""  # Próximamente
```

### Frontend

```env
VITE_API_URL="http://localhost:8080/api/v1"
VITE_SENTRY_DSN=""  # Opcional
```

## 📊 Modelo de Datos

### Entidades Principales

- **users**: Usuarios del sistema (autenticación)
- **products**: Productos del inventario
- **customers**: Clientes
- **suppliers**: Proveedores
- **sales_orders**: Órdenes de venta
- **sales_order_items**: Ítems de órdenes de venta
- **purchase_orders**: Órdenes de compra
- **purchase_order_items**: Ítems de órdenes de compra
- **stock_movements**: Movimientos de stock (historial)

## 🚀 Comandos Útiles

### Iniciar Todo

```bash
docker compose up -d
```

### Ver Estado

```bash
docker compose ps
```

### Ver Logs

```bash
# Todos los servicios
docker compose logs -f

# Solo el worker
docker logs stock_in_order_worker -f

# Solo la API
docker logs stock_in_order_api -f
```

### Reiniciar un Servicio

```bash
docker compose restart worker
docker compose restart api
```

### Detener Todo

```bash
docker compose down
```

### Detener y Eliminar Volúmenes

```bash
docker compose down -v
```

## 🧪 Testing

### Probar el Worker

1. **Usar script de prueba**:
```bash
cd worker/test-publisher
go run main.go
```

2. **Usar interfaz de RabbitMQ**:
- Acceder a http://localhost:15672
- Ir a Queues → reporting_queue
- Publish message con JSON

3. **Ver logs del worker**:
```bash
docker logs stock_in_order_worker -f
```

## 🔍 Monitoreo

### RabbitMQ Management UI

- **URL**: http://localhost:15672
- **Credenciales**: user / pass
- **Funcionalidades**:
  - Ver colas y mensajes
  - Monitorear throughput
  - Ver conexiones activas
  - Publicar mensajes manualmente
  - Ver estadísticas de consumo

### PostgreSQL

```bash
# Conectar a la base de datos
docker exec -it stock_in_order_postgres psql -U user -d stock_db

# Ver tablas
\dt

# Consultar datos
SELECT * FROM products;
```

## 🎯 Próximas Mejoras

### Corto Plazo

- [ ] Integrar SendGrid para envío de emails
- [ ] Crear endpoints en API para publicar mensajes
- [ ] Implementar notificaciones al usuario cuando el reporte esté listo
- [ ] Agregar más tipos de reportes (ventas, compras)

### Mediano Plazo

- [ ] Implementar Dead Letter Queue (DLQ) para mensajes fallidos
- [ ] Agregar múltiples workers para mejor escalabilidad
- [ ] Almacenamiento de reportes en S3 o similar
- [ ] WebSockets para notificaciones en tiempo real

### Largo Plazo

- [ ] Métricas y observabilidad (Prometheus + Grafana)
- [ ] Trazabilidad distribuida (OpenTelemetry)
- [ ] CI/CD con GitHub Actions
- [ ] Kubernetes para orquestación
- [ ] Autoscaling de workers según carga

## 📚 Referencias

- [Documentación del Worker](./worker/README.md)
- [Documentación del Backend](./backend/README.md)
- [Documentación del Frontend](./frontend/README.md)

---

**Última actualización**: 21 de Octubre 2025  
**Versión**: 2.0.0 (con Worker Service)
