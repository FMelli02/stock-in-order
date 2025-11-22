# 📚 Implementaciones y Tecnologías - Stock In Order

**Proyecto:** Sistema de Gestión de Inventario con Trazabilidad de Lotes y Suscripciones  
**Última Actualización:** 22 de Noviembre, 2025  
**Estado:** En Producción ✅

---

## 🎯 Resumen Ejecutivo

**Stock In Order** es un sistema completo de gestión de inventario empresarial que incluye:
- Gestión de productos, clientes, proveedores
- Órdenes de compra y venta con sistema de lotes
- Trazabilidad completa con lógica FEFO
- **Sistema multi-tenant con organizaciones** ⭐ **[NUEVO]**
- Autenticación JWT con RBAC
- Auditoría de operaciones
- Integración con servicios externos
- Sistema de reportes y exportación
- Notificaciones por email
- **Sistema de suscripciones con MercadoPago** ⭐ **[NUEVO]**
- Monitoreo y logging estructurado

---

## 🏗️ Arquitectura General

```
┌─────────────────────────────────────────────────────────────┐
│                     FRONTEND (SPA)                          │
│                    React + TypeScript                       │
│                    Tailwind CSS + Vite                      │
└────────────────────────┬────────────────────────────────────┘
                         │ REST API (HTTP/HTTPS)
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    BACKEND (Microservicios)                 │
├─────────────────────────────────────────────────────────────┤
│  API Server (Go)      │  Worker (Go)     │  Scheduler (Go)  │
│  - Chi Router         │  - RabbitMQ      │  - Cron Jobs     │
│  - JWT Auth           │  - Email         │  - Reportes      │
│  - Business Logic     │  - Processing    │  - Alertas       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  CAPA DE DATOS                              │
├─────────────────────────────────────────────────────────────┤
│  PostgreSQL    │   RabbitMQ      │   Sistema de Archivos    │
│  (Base Datos)  │   (Mensajería)  │   (Reportes CSV)         │
└─────────────────────────────────────────────────────────────┘
```

---

## 💻 Stack Tecnológico

### **Frontend**

| Tecnología | Versión | Propósito |
|------------|---------|-----------|
| **React** | 18.x | Framework UI principal |
| **TypeScript** | 5.x | Tipado estático |
| **Vite** | 5.x | Build tool y dev server |
| **React Router** | 6.x | Navegación SPA |
| **Tailwind CSS** | 3.x | Framework de estilos |
| **React Hot Toast** | - | Notificaciones UI |
| **Axios** | - | Cliente HTTP |
| **Chart.js / Recharts** | - | Gráficos y visualizaciones |
| **React Hook Form** | - | Manejo de formularios |
| **Zod** | - | Validación de schemas |

**Estructura del Proyecto:**
```
frontend/
├── src/
│   ├── components/     # Componentes reutilizables
│   ├── pages/          # Vistas principales
│   ├── services/       # API clients
│   ├── types/          # Definiciones TypeScript
│   ├── hooks/          # Custom hooks
│   ├── utils/          # Utilidades
│   └── App.tsx         # Punto de entrada
├── public/             # Assets estáticos
└── nginx.conf          # Configuración para producción
```

---

### **Backend - API Server (Go)**

| Tecnología | Versión | Propósito |
|------------|---------|-----------|
| **Go** | 1.25+ | Lenguaje principal |
| **Chi Router** | 5.x | HTTP router |
| **pgx/v5** | 5.x | Driver PostgreSQL |
| **golang-jwt/jwt** | 5.x | Autenticación JWT |
| **bcrypt** | - | Hash de contraseñas |
| **slog** | stdlib | Logging estructurado |
| **godotenv** | - | Configuración de entorno |
| **crypto/aes** | stdlib | Encriptación de datos sensibles |

**Estructura del Proyecto:**
```
backend/
├── cmd/
│   ├── api/                 # Servidor HTTP principal
│   ├── hasher/              # Herramienta para hashear contraseñas
│   └── demo-encryption/     # Demo de encriptación
├── internal/
│   ├── config/              # Configuración de la app
│   ├── crypto/              # Encriptación/Desencriptación
│   ├── database/            # Conexión a PostgreSQL
│   ├── handlers/            # HTTP handlers (controladores)
│   ├── middleware/          # Middlewares (Auth, CORS, Logging)
│   ├── models/              # Modelos de datos + repositorios
│   ├── rabbitmq/            # Cliente RabbitMQ
│   ├── repository/          # Capa de persistencia
│   ├── router/              # Configuración de rutas
│   └── services/            # Lógica de negocio
└── migrations/              # Migraciones SQL
```

---

### **Backend - Worker (Go)**

| Tecnología | Propósito |
|------------|-----------|
| **Go** | Procesamiento asíncrono |
| **RabbitMQ Client** | Consumo de mensajes |
| **SendGrid SDK** | Envío de emails |
| **Template Engine** | Generación de HTML para emails |

**Funcionalidades:**
- Procesa emails de forma asíncrona
- Envía notificaciones de bajo stock
- Envía confirmaciones de órdenes
- Maneja reintentos automáticos

---

### **Backend - Scheduler (Go)**

| Tecnología | Propósito |
|------------|-----------|
| **Go** | Jobs programados |
| **Cron** | Ejecución periódica |
| **pgx/v5** | Acceso a base de datos |

**Funcionalidades:**
- Generación de reportes diarios
- Alertas de productos con bajo stock
- Limpieza de lotes agotados
- Exportación de datos a CSV

---

### **Base de Datos**

| Tecnología | Versión | Propósito |
|------------|---------|-----------|
| **PostgreSQL** | 15+ | Base de datos principal |
| **pgx** | 5.x | Driver nativo Go |
| **SQL Migrations** | - | Control de versiones de esquema |

**Características:**
- ✅ Transacciones ACID
- ✅ Índices optimizados
- ✅ Foreign Keys y Constraints
- ✅ Triggers y Funciones
- ✅ Row-level locking (FOR UPDATE)

---

### **Mensajería**

| Tecnología | Versión | Propósito |
|------------|---------|-----------|
| **RabbitMQ** | 3.x | Message broker |
| **AMQP Protocol** | 0.9.1 | Protocolo de mensajería |

**Colas Implementadas:**
- `reporting_queue` - Generación de reportes
- `email_queue` - Envío de emails
- `audit_queue` - Logs de auditoría

---

### **DevOps & Infraestructura**

| Tecnología | Propósito |
|------------|-----------|
| **Docker** | Contenedores |
| **Docker Compose** | Orquestación local |
| **Nginx** | Servidor web para frontend |
| **Git** | Control de versiones |
| **GitHub** | Repositorio remoto |

---

## 🔐 Sistema de Autenticación y Autorización

### **JWT (JSON Web Tokens)**

**Implementación:**
```go
// Generación de token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": user.ID,
    "email":   user.Email,
    "role":    user.Role,  // admin, repositor, vendedor
    "exp":     time.Now().Add(24 * time.Hour).Unix(),
})
```

**Endpoints de Autenticación:**
- `POST /api/v1/auth/register` - Registro de usuarios
- `POST /api/v1/auth/login` - Login y generación de token
- `GET /api/v1/auth/me` - Información del usuario autenticado

**Middleware:**
```go
// Protección de rutas
router.Use(middleware.Authenticate())     // Requiere token válido
router.Use(middleware.RequireRole("admin", "repositor"))  // RBAC
```

---

### **RBAC (Role-Based Access Control)**

**Roles Definidos:**

| Rol | Permisos |
|-----|----------|
| **admin** | Acceso total al sistema |
| **repositor** | Gestión de inventario y compras |
| **vendedor** | Creación de órdenes de venta |

**Configuración por Endpoint:**
```go
// Productos - todos autenticados
r.Get("/products", RequireAuth(handlers.GetProducts))

// Compras - admin o repositor
r.Post("/purchase-orders", RequireRole("admin", "repositor")(handlers.CreatePurchaseOrder))

// Ventas - todos autenticados
r.Post("/sales-orders", RequireAuth(handlers.CreateSalesOrder))

// Usuarios - solo admin
r.Get("/users", RequireRole("admin")(handlers.GetUsers))
```

---

### **Encriptación de Datos Sensibles**

**Tecnología:** AES-256-GCM

**Implementación:**
```go
// Encriptar claves de integración
encryptedKey, err := crypto.Encrypt(apiKey, encryptionKey)

// Desencriptar al usar
decryptedKey, err := crypto.Decrypt(encryptedKey, encryptionKey)
```

**Datos Encriptados:**
- ✅ API Keys de integraciones
- ✅ Tokens de acceso externos
- ✅ Credenciales de servicios

---

## 📊 Módulos Funcionales

### **1. Gestión de Productos**

**Tecnologías:**
- PostgreSQL (persistencia)
- Go (lógica de negocio)
- React (UI)

**Funcionalidades:**
- ✅ CRUD completo de productos
- ✅ Categorización (no implementada visualmente, pero preparada)
- ✅ Stock calculado dinámicamente desde lotes
- ✅ Alertas de stock mínimo
- ✅ Multitenancy (por usuario)

**Endpoints:**
```
GET    /api/v1/products           - Listar productos
GET    /api/v1/products/:id       - Ver detalle
POST   /api/v1/products           - Crear producto
PUT    /api/v1/products/:id       - Actualizar producto
DELETE /api/v1/products/:id       - Eliminar producto
```

**Modelo de Datos:**
```go
type Product struct {
    ID                int64   `json:"id"`
    Name              string  `json:"name"`
    Price             float64 `json:"price"`
    StockMinimo       int     `json:"stock_minimo"`
    Notificado        bool    `json:"notificado"`
    UserID            int64   `json:"user_id"`
    CalculatedQuantity int    `json:"quantity"` // Calculado desde lotes
}
```

---

### **2. Sistema de Lotes (Batch Tracking)**

**Tecnologías:**
- PostgreSQL (tabla `product_batches`)
- FEFO Algorithm (First Expired, First Out)
- Transacciones ACID

**Características Principales:**
- ✅ Trazabilidad completa de inventario
- ✅ Fechas de vencimiento
- ✅ Números de lote
- ✅ Consumo inteligente FEFO
- ✅ Múltiples lotes por producto

**Estructura:**
```sql
CREATE TABLE product_batches (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    lote_number TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 0,
    expiry_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Índices:**
- `idx_product_batches_product_id` - Consultas rápidas
- `idx_product_batches_expiry_date` - Ordenamiento FEFO
- `idx_product_batches_user_id` - Multitenancy

**Algoritmo FEFO:**
```go
func ConsumeStockFEFO(tx, productID, userID, quantity) error {
    // 1. Obtener lotes ordenados por vencimiento
    SELECT * FROM product_batches
    WHERE product_id = ? AND user_id = ? AND quantity > 0
    ORDER BY expiry_date ASC NULLS LAST, created_at ASC
    FOR UPDATE  // Lock transaccional
    
    // 2. Iterar y consumir del más próximo a vencer
    // 3. Si falta stock → error → ROLLBACK
}
```

---

### **3. Órdenes de Compra (Purchase Orders)**

**Tecnologías:**
- PostgreSQL (transacciones)
- Go (lógica)
- React (UI)

**Funcionalidades:**
- ✅ Crear órdenes de compra
- ✅ Asociar a proveedores
- ✅ Múltiples items por orden
- ✅ Estados: pending, completed, cancelled
- ✅ Al completar → crea lotes automáticamente
- ✅ Registro de movimientos de stock

**Endpoints:**
```
GET    /api/v1/purchase-orders              - Listar órdenes
GET    /api/v1/purchase-orders/:id          - Ver detalle
POST   /api/v1/purchase-orders              - Crear orden
PUT    /api/v1/purchase-orders/:id/status   - Cambiar estado
GET    /api/v1/purchase-orders/export       - Exportar a CSV
```

**Flujo de Negocio:**
```
1. Usuario crea orden (estado: pending)
   ➜ INSERT INTO purchase_orders
   ➜ INSERT INTO purchase_order_items (con lote_number, expiry_date)

2. Mercadería llega, usuario marca como "completed"
   ➜ Por cada item:
      - INSERT INTO product_batches (crea lote)
      - INSERT INTO stock_movements (+cantidad)
      - UPDATE products SET notificado = false

3. Stock disponible aumenta automáticamente
```

**Modelo:**
```go
type PurchaseOrderItem struct {
    ID              int64      `json:"id"`
    PurchaseOrderID int64      `json:"purchase_order_id"`
    ProductID       int64      `json:"product_id"`
    Quantity        int        `json:"quantity"`
    UnitCost        float64    `json:"unit_cost"`
    LoteNumber      string     `json:"lote_number,omitempty"`     // ✨
    ExpiryDate      *time.Time `json:"expiry_date,omitempty"`     // ✨
}
```

---

### **4. Órdenes de Venta (Sales Orders)**

**Tecnologías:**
- FEFO Algorithm
- Transacciones con locks
- Validación de stock

**Funcionalidades:**
- ✅ Crear órdenes de venta
- ✅ Asociar a clientes
- ✅ Consumo automático de stock FEFO
- ✅ Validación de stock suficiente
- ✅ ROLLBACK si falta stock
- ✅ Registro de movimientos

**Endpoints:**
```
GET    /api/v1/sales-orders              - Listar órdenes
GET    /api/v1/sales-orders/:id          - Ver detalle
POST   /api/v1/sales-orders              - Crear orden
GET    /api/v1/sales-orders/export       - Exportar a CSV
```

**Flujo FEFO:**
```
1. Usuario crea orden de venta (30 unidades)
   ➜ ConsumeStockFEFO(productID, 30)

2. Sistema obtiene lotes ordenados:
   Lote A: 20 unidades, vence 2025-12-01 (primero)
   Lote B: 50 unidades, vence 2026-06-01 (después)

3. Consume del más próximo a vencer:
   ➜ Lote A: 20 → 0 (agotado)
   ➜ Lote B: 50 → 40 (restante: 30 - 20 = 10)

4. Si hubiera faltado stock → error → ROLLBACK completo
```

**Seguridad Transaccional:**
```sql
BEGIN;
    SELECT ... FOR UPDATE;  -- Lock de lotes
    UPDATE product_batches SET quantity = ...;
    INSERT INTO order_items ...;
    INSERT INTO stock_movements ...;
COMMIT;
```

---

### **5. Gestión de Clientes**

**Funcionalidades:**
- ✅ CRUD de clientes
- ✅ Datos de contacto
- ✅ Asociación a órdenes de venta

**Endpoints:**
```
GET    /api/v1/customers      - Listar
POST   /api/v1/customers      - Crear
PUT    /api/v1/customers/:id  - Actualizar
DELETE /api/v1/customers/:id  - Eliminar
```

---

### **6. Gestión de Proveedores (Suppliers)**

**Funcionalidades:**
- ✅ CRUD de proveedores
- ✅ Datos de contacto
- ✅ Asociación a órdenes de compra

**Endpoints:**
```
GET    /api/v1/suppliers      - Listar
POST   /api/v1/suppliers      - Crear
PUT    /api/v1/suppliers/:id  - Actualizar
DELETE /api/v1/suppliers/:id  - Eliminar
```

---

### **7. Movimientos de Stock (Stock Movements)**

**Tecnología:** Auditoría automática

**Funcionalidades:**
- ✅ Registro de todos los movimientos
- ✅ Trazabilidad completa
- ✅ Motivos: PURCHASE_ORDER, SALES_ORDER, ADJUSTMENT

**Estructura:**
```sql
CREATE TABLE stock_movements (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity_change INTEGER NOT NULL,  -- Positivo o negativo
    reason TEXT NOT NULL,               -- PURCHASE_ORDER, SALES_ORDER
    reference_id TEXT,                  -- ID de la orden
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Ejemplos:**
```
Compra recibida:
  quantity_change: +100
  reason: PURCHASE_ORDER
  reference_id: "15"

Venta realizada:
  quantity_change: -30
  reason: SALES_ORDER
  reference_id: "24"
```

---

### **8. Dashboard y KPIs**

**Tecnologías:**
- PostgreSQL (queries agregadas)
- Chart.js (gráficos)

**Métricas Implementadas:**
- ✅ Total de productos
- ✅ Productos con bajo stock
- ✅ Total de clientes
- ✅ Total de proveedores
- ✅ Órdenes de venta (total y por estado)
- ✅ Órdenes de compra (total y por estado)

**Endpoints:**
```
GET /api/v1/dashboard/kpis  - Todas las métricas
```

**Consulta de Productos con Bajo Stock (con lotes):**
```sql
SELECT p.id, p.name, COALESCE(SUM(pb.quantity), 0) as stock
FROM products p
LEFT JOIN product_batches pb ON p.id = pb.product_id
WHERE p.user_id = $1
GROUP BY p.id, p.name, p.stock_minimo
HAVING COALESCE(SUM(pb.quantity), 0) <= p.stock_minimo
```

---

### **9. Sistema de Reportes**

**Tecnologías:**
- Go (generación)
- CSV (formato)
- Scheduler (automatización)

**Reportes Disponibles:**
- ✅ Exportación de productos
- ✅ Exportación de órdenes de compra
- ✅ Exportación de órdenes de venta
- ✅ Filtros por fecha y estado

**Endpoints:**
```
GET /api/v1/products/export           - Productos a CSV
GET /api/v1/purchase-orders/export    - Compras a CSV
GET /api/v1/sales-orders/export       - Ventas a CSV
```

**Características:**
- ✅ Generación on-demand
- ✅ Descarga directa
- ✅ Headers descriptivos
- ✅ Encoding UTF-8

---

### **10. Sistema de Notificaciones**

**Tecnologías:**
- RabbitMQ (cola de mensajes)
- SendGrid (SMTP)
- Worker (procesamiento)

**Tipos de Notificaciones:**
- ✅ Bajo stock de productos
- ✅ Confirmación de órdenes
- ✅ Alertas de vencimiento (preparado)

**Flujo:**
```
1. API detecta bajo stock
   ➜ Publica mensaje en RabbitMQ

2. Worker consume mensaje
   ➜ Genera HTML del email
   ➜ Envía vía SendGrid

3. Usuario recibe email
```

**Configuración:**
```env
SENDGRID_API_KEY=xxx
SENDGRID_FROM_EMAIL=noreply@stockinorder.com
SENDGRID_FROM_NAME=Stock In Order
```

---

### **11. Sistema de Integrations**

**Tecnologías:**
- PostgreSQL (tabla `integrations`)
- Encriptación AES-256

**Funcionalidades:**
- ✅ Almacenar credenciales de APIs externas
- ✅ Encriptación de API Keys
- ✅ Gestión por usuario

**Estructura:**
```sql
CREATE TABLE integrations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    service_name TEXT NOT NULL,     -- SendGrid, Stripe, etc.
    api_key_encrypted BYTEA NOT NULL,
    config JSONB,
    is_active BOOLEAN DEFAULT true
);
```

**Uso:**
```go
// Guardar integración
encryptedKey := crypto.Encrypt(apiKey)
INSERT INTO integrations (user_id, service_name, api_key_encrypted)

// Usar integración
decryptedKey := crypto.Decrypt(apiKeyEncrypted)
sendgridClient := sendgrid.NewClient(decryptedKey)
```

---

### **12. Sistema de Auditoría**

**Tecnologías:**
- PostgreSQL (tabla `audit_logs`)
- Goroutines (async)

**Funcionalidades:**
- ✅ Log de todas las operaciones críticas
- ✅ Registro asíncrono (no bloquea requests)
- ✅ Información de usuario, acción, timestamp

**Estructura:**
```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    action TEXT NOT NULL,           -- CREATE_PRODUCT, UPDATE_ORDER, etc.
    entity_type TEXT NOT NULL,      -- product, order, customer
    entity_id BIGINT,
    details JSONB,
    ip_address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Implementación:**
```go
// Async logging
go auditRepo.Log(ctx, AuditLog{
    UserID:     userID,
    Action:     "CREATE_SALES_ORDER",
    EntityType: "sales_order",
    EntityID:   orderID,
    Details:    map[string]interface{}{"items": len(items)},
    IPAddress:  r.RemoteAddr,
})
```

**Endpoints:**
```
GET /api/v1/audit-logs  - Consultar logs (admin only)
```

---

## 🔧 Middleware y Utilidades

### **Middlewares Implementados**

| Middleware | Propósito |
|------------|-----------|
| **Authenticate** | Valida JWT en todas las rutas protegidas |
| **RequireRole** | Verifica roles específicos (RBAC) |
| **CORS** | Configura políticas de origen cruzado |
| **Logger** | Registra todas las requests HTTP |
| **Recover** | Captura panics y retorna 500 |
| **RequestID** | Genera ID único por request |

**Ejemplo de Configuración:**
```go
router.Use(middleware.Logger)
router.Use(middleware.Recover)
router.Use(middleware.CORS)
router.Use(middleware.RequestID)

// Rutas protegidas
router.Group(func(r chi.Router) {
    r.Use(middleware.Authenticate)
    r.Get("/products", handlers.GetProducts)
})

// Rutas con role específico
router.Group(func(r chi.Router) {
    r.Use(middleware.RequireRole("admin"))
    r.Get("/users", handlers.GetUsers)
})
```

---

### **Logging Estructurado**

**Tecnología:** `log/slog` (Go stdlib)

**Características:**
- ✅ JSON output
- ✅ Niveles: INFO, WARN, ERROR
- ✅ Contexto enriquecido
- ✅ Búsqueda fácil en logs

**Ejemplo:**
```go
slog.Info("ConsumeStockFEFO: processing batch",
    "batchID", b.id,
    "lote", b.loteNumber,
    "available", b.quantity,
    "needed", remaining,
    "expiry", expiryStr)
```

**Output:**
```json
{
  "time": "2025-11-05T21:45:00Z",
  "level": "INFO",
  "msg": "ConsumeStockFEFO: processing batch",
  "batchID": 28,
  "lote": "TEST-VENCE-PRONTO",
  "available": 30,
  "needed": 25,
  "expiry": "2025-12-01"
}
```

---

## 🗄️ Base de Datos - Esquema Completo

### **Tablas Principales**

| Tabla | Propósito | Filas Aprox. |
|-------|-----------|--------------|
| **users** | Usuarios del sistema | 1-100 |
| **products** | Catálogo de productos | 100-10,000 |
| **product_batches** | Lotes de inventario | 500-50,000 |
| **customers** | Clientes | 100-5,000 |
| **suppliers** | Proveedores | 10-500 |
| **sales_orders** | Órdenes de venta | 1,000-100,000 |
| **order_items** | Items de ventas | 5,000-500,000 |
| **purchase_orders** | Órdenes de compra | 500-50,000 |
| **purchase_order_items** | Items de compras | 2,000-200,000 |
| **stock_movements** | Movimientos de stock | 10,000-1,000,000 |
| **integrations** | Integraciones externas | 1-50 |
| **audit_logs** | Logs de auditoría | 50,000-5,000,000 |
| **subscriptions** | Suscripciones de pago | 1-1,000 | **[NUEVO]**

### **Relaciones Clave**

```
users (1) ─── (N) products
users (1) ─── (N) subscriptions       # [NUEVO]
products (1) ─── (N) product_batches
products (1) ─── (N) stock_movements

suppliers (1) ─── (N) purchase_orders
purchase_orders (1) ─── (N) purchase_order_items

customers (1) ─── (N) sales_orders
sales_orders (1) ─── (N) order_items

users (1) ─── (N) integrations
users (1) ─── (N) audit_logs
```

### **Migraciones Aplicadas**

| # | Nombre | Descripción |
|---|--------|-------------|
| 000001 | `create_users_table` | Tabla de usuarios |
| 000002 | `create_products_table` | Tabla de productos |
| 000003 | `create_suppliers_table` | Tabla de proveedores |
| 000004 | `create_customers_table` | Tabla de clientes |
| 000005 | `create_sales_orders_tables` | Órdenes de venta |
| 000006 | `create_purchase_orders_tables` | Órdenes de compra |
| 000007 | `create_stock_movements_table` | Movimientos |
| 000008 | `seed_initial_data` | Datos iniciales |
| 000009 | `add_user_roles` | Sistema de roles |
| 000010 | `add_stock_alerts_to_products` | Alertas de stock |
| 000011 | `create_integrations_table` | Integraciones |
| 000012 | `create_audit_logs_table` | Auditoría |
| 000013 | `add_batch_tracking` | **Sistema de lotes** ⭐ |
| 000014 | `add_batch_fields_to_purchase_items` | Campos de lote en compras ⭐ |
| 000015 | `create_subscriptions_table` | **Tabla de suscripciones** ⭐ |
| 000016 | `add_plan_id_to_subscriptions` | Plan ID y features adicionales |
| 000017 | `add_organization_id_to_users` | **Sistema multi-tenant** ⭐ **[NUEVO]** |

---

## 🚀 Características Avanzadas

### **1. Multitenancy**

**Implementación:**
- Cada usuario tiene sus propios datos aislados
- Filtrado por `user_id` en todas las consultas
- Indices optimizados por usuario

**Garantías:**
- ✅ Usuario A no puede ver datos de Usuario B
- ✅ Escalable horizontalmente
- ✅ Sin contaminación de datos

---

### **2. Concurrencia y Locks**

**Tecnología:** PostgreSQL Row-Level Locking

**Implementación:**
```sql
SELECT * FROM product_batches
WHERE product_id = ? AND quantity > 0
ORDER BY expiry_date ASC
FOR UPDATE;  -- Bloquea las filas
```

**Previene:**
- ❌ Stock negativo
- ❌ Condiciones de carrera
- ❌ Doble venta del mismo lote

---

### **3. Transacciones ACID**

**Garantías:**
- **Atomicidad:** Todo o nada
- **Consistencia:** Stock siempre correcto
- **Aislamiento:** Transacciones no se interfieren
- **Durabilidad:** Cambios persistentes

**Ejemplo:**
```go
tx.Begin()
    ConsumeStockFEFO()     // Si falla → ROLLBACK
    InsertOrderItem()      // Si falla → ROLLBACK
    InsertMovement()       // Si falla → ROLLBACK
tx.Commit()                // Todo OK → COMMIT
```

---

### **4. Validaciones Multicapa**

**Frontend (TypeScript + Zod):**
```typescript
const schema = z.object({
  quantity: z.number().min(1).max(10000),
  unit_price: z.number().min(0),
})
```

**Backend (Go):**
```go
if quantity <= 0 {
    return errors.New("quantity must be positive")
}
```

**Base de Datos (Constraints):**
```sql
quantity INTEGER NOT NULL CHECK (quantity >= 0)
```

---

### **5. Rate Limiting (Preparado)**

**Tecnología:** Middleware Chi

**Configuración Sugerida:**
```go
router.Use(middleware.Throttle(100)) // 100 req/min
```

---

### **6. Caching (Preparado para Redis)**

**Estrategia:**
- Cache de productos (TTL: 5 min)
- Cache de dashboard (TTL: 1 min)
- Invalidación en escritura

---

## 📈 Métricas y Monitoreo

### **Logging**

**Ubicación:** Docker logs

**Comandos:**
```bash
# Ver logs de API
docker logs -f stock_in_order_api

# Ver logs de Worker
docker logs -f stock_in_order_worker

# Buscar errores
docker logs stock_in_order_api | grep ERROR
```

---

### **Monitoreo de Salud**

**Endpoints:**
```
GET /health           - Estado del servicio
GET /health/db        - Estado de PostgreSQL
GET /health/rabbitmq  - Estado de RabbitMQ
```

---

## 🔒 Seguridad Implementada

### **Autenticación y Autorización**
- ✅ JWT con expiración (24h)
- ✅ Contraseñas hasheadas con bcrypt (cost 10)
- ✅ RBAC (3 roles)
- ✅ Validación de ownership en todas las queries

### **Protección de Datos**
- ✅ Encriptación AES-256 para API keys
- ✅ HTTPS ready (configurar reverse proxy)
- ✅ CORS configurado
- ✅ SQL Injection prevention (prepared statements)

### **Auditoría**
- ✅ Log de todas las operaciones críticas
- ✅ IP tracking
- ✅ Timestamp de acciones

---

## 🧪 Testing y Calidad

### **Validaciones Realizadas**

- ✅ Migración de 27 productos a lotes
- ✅ Stock calculado correctamente
- ✅ FEFO consume en orden correcto
- ✅ ROLLBACK en stock insuficiente
- ✅ Sin stock negativo
- ✅ Logs estructurados funcionando

### **Comandos de Verificación**

```bash
# Verificar stock total
docker exec stock_in_order_postgres psql -U user -d stock_db -c \
"SELECT COUNT(*) as lotes, SUM(quantity) as stock FROM product_batches;"

# Verificar orden FEFO
docker exec stock_in_order_postgres psql -U user -d stock_db -c \
"SELECT * FROM product_batches WHERE product_id = 21 
ORDER BY expiry_date ASC NULLS LAST, created_at ASC;"

# Verificar movimientos
docker exec stock_in_order_postgres psql -U user -d stock_db -c \
"SELECT * FROM stock_movements ORDER BY created_at DESC LIMIT 10;"
```

---

## 📦 Despliegue

### **Entornos**

| Entorno | URL | Propósito |
|---------|-----|-----------|
| **Desarrollo** | localhost:8080 | Desarrollo local |
| **Producción** | TBD | Producción |

### **Variables de Entorno Requeridas**

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=stock_db

# JWT
JWT_SECRET=your-secret-key-here

# Encryption
ENCRYPTION_KEY=32-byte-hex-key

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

# SendGrid
SENDGRID_API_KEY=SG.xxx
SENDGRID_FROM_EMAIL=noreply@example.com
SENDGRID_FROM_NAME=Stock In Order

# Sentry (opcional)
SENTRY_DSN=https://xxx@sentry.io/xxx

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

### **Docker Compose Services**

```yaml
services:
  postgres:     # Base de datos
  rabbitmq:     # Message broker
  api:          # API backend
  worker:       # Procesamiento async
  scheduler:    # Jobs programados
  frontend:     # React app
```

---

### **13. Sistema Multi-Tenant con Organizaciones**

**Tecnologías:**
- PostgreSQL (columna organization_id)
- JWT (claims de organización)
- Go (middleware de contexto)
- React (filtrado automático)

**Funcionalidades:**
- ✅ Cada admin es una organización independiente
- ✅ Vendedores/Repositores comparten inventario del admin
- ✅ Aislamiento completo entre organizaciones
- ✅ JWT incluye organization_id
- ✅ Filtrado automático en todas las queries
- ✅ Migration sin pérdida de datos

**Migración 000017:**
```sql
-- Agregar columna organization_id a users
ALTER TABLE users ADD COLUMN organization_id BIGINT;

-- Foreign key auto-referencial
ALTER TABLE users 
ADD CONSTRAINT fk_users_organization 
FOREIGN KEY (organization_id) 
REFERENCES users(id) 
ON DELETE CASCADE;

-- Para admins existentes: organization_id = su propio ID
UPDATE users 
SET organization_id = id 
WHERE role = 'admin';

-- Índice para performance
CREATE INDEX idx_users_organization_id 
ON users(organization_id);
```

**Modelo de Organización:**
```
Admin (ID=5, organization_id=5)
  └─ Productos (user_id=5) → 15 productos
  └─ Clientes (user_id=5) → 10 clientes
  └─ Proveedores (user_id=5) → 5 proveedores
  └─ Vendedor1 (ID=11, organization_id=5) → Ve los 15 productos
  └─ Vendedor2 (ID=12, organization_id=5) → Ve los 15 productos
  └─ Repositor1 (ID=13, organization_id=5) → Ve los 15 productos

Admin Nuevo (ID=14, organization_id=14)
  └─ Productos (user_id=14) → 0 productos (organización nueva)
  └─ Vendedor3 (ID=15, organization_id=14) → Ve 0 productos
```

**JWT con Organization ID:**
```go
// Generación de token (LoginUser)
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id":         user.ID,
    "email":           user.Email,
    "role":            user.Role,
    "organization_id": user.OrganizationID,  // ⭐ NUEVO
    "exp":             time.Now().Add(24 * time.Hour).Unix(),
})
```

**Middleware de Contexto:**
```go
// Extraer organization_id del JWT
func JWTMiddleware(next http.Handler, jwtSecret string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ... validar JWT ...
        
        claims, _ := token.Claims.(jwt.MapClaims)
        
        // Extraer organization_id (con fallback a user_id para admins)
        orgIDVal, _ := claims["organization_id"]
        var orgID int64
        switch v := orgIDVal.(type) {
        case float64:
            orgID = int64(v)
        case int64:
            orgID = v
        default:
            // Fallback: usar user_id para tokens viejos
            orgID = userID
        }
        
        // Inyectar en contexto
        ctx := context.WithValue(r.Context(), organizationIDKey, orgID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Helper para obtener organization_id
func OrganizationIDFromContext(ctx context.Context) (int64, bool) {
    v := ctx.Value(organizationIDKey)
    if v == nil {
        return 0, false
    }
    orgID, ok := v.(int64)
    return orgID, ok
}
```

**Handlers Actualizados (todos):**
```go
// Antes (solo user_id)
func ListProducts(db *pgxpool.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, _ := middleware.UserIDFromContext(r.Context())
        products, _ := productModel.GetAllForUser(userID)
        // ...
    }
}

// Después (organization_id)
func ListProducts(db *pgxpool.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        organizationID, _ := middleware.OrganizationIDFromContext(r.Context())  // ⭐
        products, _ := productModel.GetAllForUser(organizationID)
        // ...
    }
}
```

**Modelos con Organization ID:**
```go
// Product Model
func (m *ProductModel) GetAllForUser(organizationID int64) ([]Product, error) {
    const q = `
        SELECT p.id, p.name, p.sku, ...
        FROM products p
        WHERE p.user_id = $1  -- user_id sigue siendo la FK, pero recibe organization_id
        ORDER BY p.name
    `
    rows, _ := m.DB.Query(ctx, q, organizationID)
    // ...
}

// Customer Model
func (m *CustomerModel) GetAllForUser(organizationID int64) ([]Customer, error) {
    const q = `SELECT * FROM customers WHERE user_id = $1`
    rows, _ := m.DB.Query(ctx, q, organizationID)
    // ...
}

// Supplier Model
func (m *SupplierModel) GetAllForUser(organizationID int64) ([]Supplier, error) {
    const q = `SELECT * FROM suppliers WHERE user_id = $1`
    rows, _ := m.DB.Query(ctx, q, organizationID)
    // ...
}
```

**Creación de Usuarios por Admin:**
```go
// CreateUserByAdmin - el vendedor hereda organization_id del admin
func CreateUserByAdmin(db *pgxpool.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Obtener organization_id del admin que está creando el usuario
        adminOrgID, ok := middleware.OrganizationIDFromContext(r.Context())
        if !ok {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        
        var input CreateUserInput
        json.NewDecoder(r.Body).Decode(&input)
        
        newUser := &models.User{
            Name:           input.Name,
            Email:          input.Email,
            PasswordHash:   hashPassword(input.Password),
            Role:           input.Role,  // "vendedor" o "repositor"
            OrganizationID: adminOrgID,  // ⭐ Hereda del admin
        }
        
        userModel.Insert(newUser)
        // ...
    }
}
```

**Registro Público (cada uno su organización):**
```go
// RegisterUser - cada registro público crea un admin independiente
func RegisterUser(store userInserter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var input RegisterUserInput
        json.NewDecoder(r.Body).Decode(&input)
        
        // Registro público SIEMPRE crea admins
        user := &models.User{
            Name:           input.Name,
            Email:          input.Email,
            PasswordHash:   hashPassword(input.Password),
            Role:           "admin",        // ⭐ Forzado a admin
            OrganizationID: 0,              // Se auto-asigna en Insert()
        }
        
        store.Insert(user)
        // Insert() detecta role=admin y hace: organization_id = user.ID
        // ...
    }
}
```

**Insert() con Auto-Asignación:**
```go
func (m *UserModel) Insert(user *User) error {
    tx, _ := m.DB.Begin(ctx)
    defer tx.Rollback(ctx)
    
    // Convertir organization_id=0 a NULL para la base de datos
    var orgID interface{}
    if user.OrganizationID == 0 {
        orgID = nil
    } else {
        orgID = user.OrganizationID
    }
    
    // Insertar usuario
    const q = `INSERT INTO users (name, email, password_hash, role, organization_id)
               VALUES ($1, $2, $3, $4, $5)
               RETURNING id, created_at`
    var id int64
    err := tx.QueryRow(ctx, q, user.Name, user.Email, user.PasswordHash, 
                       user.Role, orgID).Scan(&id, &createdAt)
    
    user.ID = id
    
    // Si es admin y no tiene organization_id, auto-asignarse
    if user.Role == "admin" && user.OrganizationID == 0 {
        const qUpdate = `UPDATE users SET organization_id = $1 WHERE id = $1`
        tx.Exec(ctx, qUpdate, id)
        user.OrganizationID = id  // ⭐ El admin es su propia organización
    }
    
    tx.Commit(ctx)
    return nil
}
```

**Handlers Actualizados (18+):**
- ✅ `product_handlers.go` - 7 funciones
- ✅ `customer_handlers.go` - 6 funciones
- ✅ `supplier_handlers.go` - 6 funciones
- ✅ `sales_order_handlers.go` - 5 funciones
- ✅ `purchase_order_handlers.go` - 5 funciones
- ✅ `dashboard_handlers.go` - 3 funciones ⭐ (fix crítico)
- ✅ `report_handlers.go` - 3 funciones
- ✅ `audit_handlers.go` - 1 función
- ✅ `integration_handlers.go` - 3 funciones
- ✅ `subscription_handlers.go` - ya usa user_id (correcto)
- ✅ `user_handlers.go` - CreateUserByAdmin, RegisterUser

**Fix del Dashboard (Bug Crítico):**
```go
// ANTES (incorrecto - usaba UserIDFromContext)
func GetDashboardKPIs(db *pgxpool.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        organizationID, _ := middleware.UserIDFromContext(r.Context())  // ❌
        // ...
    }
}

// DESPUÉS (correcto - usa OrganizationIDFromContext)
func GetDashboardKPIs(db *pgxpool.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        organizationID, _ := middleware.OrganizationIDFromContext(r.Context())  // ✅
        dm := &models.DashboardModel{DB: db}
        kpis, _ := dm.GetDashboardKPIs(organizationID)
        // ...
    }
}
```

**Verificación en Producción:**
```sql
-- Estado de usuarios
SELECT id, email, role, organization_id FROM users ORDER BY id;

-- Resultado:
-- 5  | francoleproso1@gmail.com    | admin    | 5  ✅
-- 11 | vendedor@test.com           | vendedor | 5  ✅
-- 12 | mellimacifranco@gmail.com   | vendedor | 5  ✅
-- 13 | prueba@example.com          | vendedor | 5  ✅
-- 14 | pruebaregistro@gmail.com    | admin    | 14 ✅

-- Productos por organización
SELECT user_id, COUNT(*) as total FROM products GROUP BY user_id;

-- Resultado:
-- 5  | 15  ✅ (admin principal)
-- 14 | 0   ✅ (admin nuevo, organización vacía)

-- Verificar que vendedores ven productos del admin
SELECT 
    u.id as user_id,
    u.email,
    u.organization_id,
    COUNT(p.id) as products_visible
FROM users u
LEFT JOIN products p ON p.user_id = u.organization_id
WHERE u.organization_id IS NOT NULL
GROUP BY u.id, u.email, u.organization_id;

-- Resultado:
-- 5  | francoleproso1@gmail.com  | 5  | 15  ✅
-- 11 | vendedor@test.com         | 5  | 15  ✅
-- 12 | mellimacifranco@gmail.com | 5  | 15  ✅
-- 13 | prueba@example.com        | 5  | 15  ✅
-- 14 | pruebaregistro@gmail.com  | 14 | 0   ✅
```

**Logs de Debug:**
```json
// Login exitoso con organization_id
{
  "time": "2025-11-20T21:25:52Z",
  "level": "INFO",
  "msg": "JWT Middleware",
  "user_id": 5,
  "email": "francoleproso1@gmail.com",
  "role": "admin",
  "organization_id": 5,  // ⭐
  "path": "/api/v1/products"
}

// ListProducts con organization_id correcto
{
  "time": "2025-11-20T21:25:52Z",
  "level": "INFO",
  "msg": "ListProducts called",
  "organizationID": 5  // ⭐
}

// Resultado: 15 productos
{
  "time": "2025-11-20T21:25:52Z",
  "level": "INFO",
  "msg": "ListProducts result",
  "organizationID": 5,
  "count": 15  // ✅
}
```

**Garantías del Sistema:**
- ✅ Admins tienen su propia organización (organization_id = su ID)
- ✅ Vendedores/Repositores heredan organization_id del admin que los creó
- ✅ Cada organización ve solo sus datos (productos, clientes, proveedores)
- ✅ JWT incluye organization_id para autenticación stateless
- ✅ Middleware inyecta organization_id en contexto de cada request
- ✅ Todos los handlers usan OrganizationIDFromContext()
- ✅ Dashboard muestra métricas correctas por organización
- ✅ Migraciones aplicadas sin pérdida de datos
- ✅ Sistema multi-tenant 100% funcional

---

### **14. Gestión de Usuarios con Organizaciones**

**Endpoints:**
```
POST /api/v1/auth/register          - Registro público (crea admin)
POST /api/v1/auth/login             - Login (genera JWT con organization_id)
POST /api/v1/admin/users            - Admin crea vendedor/repositor
GET  /api/v1/admin/users            - Listar usuarios de la organización
PUT  /api/v1/admin/users/:id        - Actualizar usuario
DELETE /api/v1/admin/users/:id      - Eliminar usuario
```

**Flujos de Usuario:**

1. **Registro Público:**
```
Usuario → POST /auth/register
  ↓
Backend crea user con role="admin"
  ↓
Insert() auto-asigna organization_id = user.ID
  ↓
Usuario tiene su propia organización vacía
```

2. **Admin Crea Vendedor:**
```
Admin (org_id=5) → POST /admin/users {role: "vendedor"}
  ↓
Backend extrae organization_id=5 del JWT del admin
  ↓
Crea vendedor con organization_id=5
  ↓
Vendedor puede ver los 15 productos del admin
```

3. **Login de Vendedor:**
```
Vendedor → POST /auth/login
  ↓
Backend busca user en BD (organization_id=5)
  ↓
Genera JWT con organization_id=5
  ↓
Frontend recibe token
  ↓
Todas las peticiones filtran por organization_id=5
  ↓
Ve productos, clientes, proveedores del admin
```

---

### **15. Sistema de Suscripciones y Pagos**

**Tecnologías:**
- MercadoPago API (Payment Gateway)
- PostgreSQL (persistencia)
- Webhooks (notificaciones asíncronas)
- Go (backend)
- React (frontend)

**Funcionalidades:**
- ✅ Múltiples planes de suscripción (Básico, Pro, Enterprise)
- ✅ Integración con MercadoPago (checkout y webhooks)
- ✅ Gestión de estados (pending, active, cancelled, expired)
- ✅ Renovación automática de suscripciones
- ✅ Cancelación de suscripciones
- ✅ Verificación de firma de webhooks (seguridad)
- ✅ Límites por plan (productos, órdenes, features)

**Endpoints de Suscripciones:**
```
POST   /api/v1/subscriptions/create-checkout  - Crear checkout de pago
GET    /api/v1/subscriptions/status           - Ver suscripción actual
POST   /api/v1/subscriptions/cancel           - Cancelar suscripción
POST   /api/v1/webhooks/mercadopago           - Recibir notificaciones
```

**Modelo de Datos:**
```go
type Subscription struct {
    ID              int64      `json:"id"`
    UserID          int64      `json:"user_id"`
    PlanType        string     `json:"plan_type"`        // basico, pro, enterprise
    Status          string     `json:"status"`           // pending, active, cancelled, expired
    MercadoPagoID   string     `json:"mercadopago_id"`   // ID del pago
    StartDate       time.Time  `json:"start_date"`
    EndDate         time.Time  `json:"end_date"`
    AutoRenew       bool       `json:"auto_renew"`
    CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
}

type Plan struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Price           float64                `json:"price"`
    Currency        string                 `json:"currency"`
    BillingCycle    string                 `json:"billing_cycle"`  // monthly, yearly
    Features        PlanFeatures           `json:"features"`
}

type PlanFeatures struct {
    MaxProducts     int    `json:"max_products"`
    MaxOrders       int    `json:"max_orders"`
    Reports         bool   `json:"reports"`
    APIAccess       bool   `json:"api_access"`
    Integrations    bool   `json:"integrations"`
    PrioritySupport bool   `json:"priority_support"`
    CustomReports   bool   `json:"custom_reports"`
    MultiUser       bool   `json:"multi_user"`
}
```

**Planes Configurados:**

| Plan | Precio | Productos | Órdenes/mes | API | Integraciones | Soporte |
|------|--------|-----------|-------------|-----|---------------|---------|
| **Básico** | $5,000 ARS | 200 | 100 | ❌ | ❌ | Email |
| **Pro** | $15,000 ARS | 1,000 | 500 | ✅ | ✅ | Prioritario |
| **Enterprise** | $40,000 ARS | ∞ | ∞ | ✅ | ✅ | Dedicado |

**Flujo de Suscripción:**
```
1. Usuario selecciona plan en /pricing
   ↓
2. Frontend → POST /subscriptions/create-checkout
   ↓
3. Backend crea preferencia en MercadoPago
   ↓
4. Usuario redirigido a checkout de MercadoPago
   ↓
5. Usuario completa pago
   ↓
6. MercadoPago → POST /webhooks/mercadopago
   ↓
7. Backend verifica firma y actualiza subscription.status = 'active'
   ↓
8. Usuario tiene acceso completo según plan
```

**Webhooks de MercadoPago:**
```go
// Tipos de notificación soportados
switch topic {
    case "payment":
        // Pago aprobado → activar suscripción
        // Pago rechazado → mantener pending
        // Pago cancelado → marcar cancelled
    
    case "merchant_order":
        // Orden completada → verificar pago
}

// Verificación de seguridad
func VerifyWebhookSignature(xSignature, xRequestID string, dataID string) bool {
    expectedSignature := GenerateHMAC(dataID + xRequestID, secret)
    return CompareSignatures(xSignature, expectedSignature)
}
```

**Estados de Suscripción:**

| Estado | Descripción | Acceso |
|--------|-------------|--------|
| **pending** | Pago iniciado, no completado | ❌ Limitado |
| **active** | Pago confirmado, suscripción activa | ✅ Completo |
| **cancelled** | Usuario canceló, válida hasta end_date | ✅ Hasta fin periodo |
| **expired** | Periodo terminado sin renovación | ❌ Bloqueado |

**Renovación Automática:**
```go
// Scheduler ejecuta diariamente
func RenewSubscriptions() {
    // 1. Buscar suscripciones próximas a vencer (auto_renew = true)
    expiringSubscriptions := FindExpiringSubscriptions(3 días)
    
    // 2. Por cada suscripción:
    for _, sub := range expiringSubscriptions {
        // Crear nuevo pago en MercadoPago
        payment := CreateRecurringPayment(sub.UserID, sub.PlanType)
        
        // Enviar email de renovación
        SendRenewalEmail(sub.UserID, payment.CheckoutURL)
    }
}
```

---

### **14. Middleware de Paywall (Patovica)**

**Tecnología:** Middleware Go

**Funcionalidades:**
- ✅ Verificación de suscripción activa
- ✅ Validación de límites por plan
- ✅ Bloqueo automático en endpoints protegidos
- ✅ Respuesta HTTP 402 Payment Required
- ✅ Mensajes personalizados por límite

**Implementación:**
```go
// Middleware principal
func RequireActiveSubscription(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := GetUserIDFromContext(r)
        
        // Obtener suscripción
        subscription := GetActiveSubscription(userID)
        
        // Verificar si está activa
        if subscription == nil || !IsActive(subscription) {
            w.WriteHeader(http.StatusPaymentRequired) // 402
            json.NewEncoder(w).Encode(map[string]interface{}{
                "error": "Suscripción requerida",
                "upgrade_url": "/pricing",
            })
            return
        }
        
        // Agregar suscripción al contexto
        ctx := context.WithValue(r.Context(), "subscription", subscription)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Verificación de límites
func CheckPlanLimits(resource string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            subscription := GetSubscriptionFromContext(r)
            plan := GetPlanFeatures(subscription.PlanType)
            
            switch resource {
            case "products":
                count := CountUserProducts(subscription.UserID)
                if count >= plan.MaxProducts {
                    w.WriteHeader(http.StatusPaymentRequired)
                    json.NewEncoder(w).Encode(map[string]interface{}{
                        "error": fmt.Sprintf("Límite de %d productos alcanzado", plan.MaxProducts),
                        "upgrade_url": "/pricing",
                    })
                    return
                }
            
            case "orders":
                count := CountMonthlyOrders(subscription.UserID)
                if count >= plan.MaxOrders {
                    w.WriteHeader(http.StatusPaymentRequired)
                    json.NewEncoder(w).Encode(map[string]interface{}{
                        "error": fmt.Sprintf("Límite de %d órdenes/mes alcanzado", plan.MaxOrders),
                        "upgrade_url": "/pricing",
                    })
                    return
                }
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Rutas Protegidas (42 endpoints):**
```go
// Productos
r.Post("/products", RequireActiveSubscription(CheckPlanLimits("products")(CreateProduct)))
r.Put("/products/{id}", RequireActiveSubscription(UpdateProduct))

// Órdenes
r.Post("/sales-orders", RequireActiveSubscription(CheckPlanLimits("orders")(CreateSalesOrder)))
r.Post("/purchase-orders", RequireActiveSubscription(CheckPlanLimits("orders")(CreatePurchaseOrder)))

// Reportes (solo planes Pro y Enterprise)
r.Get("/products/export", RequireActiveSubscription(RequirePlanFeature("reports")(ExportProducts)))

// API Access (solo planes con api_access = true)
r.Get("/api/v1/external/*", RequireActiveSubscription(RequirePlanFeature("api_access")(APIHandler)))

// Integraciones
r.Post("/integrations", RequireActiveSubscription(RequirePlanFeature("integrations")(CreateIntegration)))
```

**Respuestas del Paywall:**
```json
// Sin suscripción
{
  "error": "Necesitas una suscripción activa para acceder a esta función",
  "upgrade_url": "/pricing"
}

// Límite alcanzado
{
  "error": "Límite de 200 productos alcanzado. Actualiza a plan Pro para 1,000 productos.",
  "current_count": 200,
  "plan_limit": 200,
  "upgrade_url": "/pricing"
}

// Feature no disponible
{
  "error": "Esta función requiere plan Pro o superior",
  "current_plan": "basico",
  "required_plan": "pro",
  "upgrade_url": "/pricing"
}
```

---

### **16. Paywall Middleware (Patovica)**

**Tecnologías:**
- React 18
- TypeScript
- Tailwind CSS
- React Router
- Axios

**Páginas Implementadas:**

#### **PricingPage.tsx** (Página Pública)
- ✅ 3 tarjetas de precios con diseño moderno
- ✅ Plan "Pro" destacado como "Más Popular"
- ✅ Lista de características por plan
- ✅ Botones "Suscribirme" con integración a MercadoPago
- ✅ Sección de FAQ
- ✅ CTA para empresas
- ✅ Responsive design (grid 3 columnas)
- ✅ Iconos SVG inline personalizados

**Componentes Clave:**
```typescript
const handleSubscribe = async (planId: string) => {
    try {
        setLoading(true);
        
        // Crear checkout en backend
        const response = await api.post('/subscriptions/create-checkout', {
            plan_type: planId,
            billing_cycle: 'monthly'
        });
        
        // Redirigir a MercadoPago
        window.location.href = response.data.checkout_url;
        
    } catch (error) {
        const apiError = error as { 
            response?: { 
                status?: number; 
                data?: { error?: string } 
            } 
        };
        
        if (apiError.response?.status === 401) {
            // No autenticado → redirigir a login
            navigate('/login', { 
                state: { from: '/pricing', planId } 
            });
        } else {
            toast.error('Error al iniciar suscripción');
        }
    } finally {
        setLoading(false);
    }
};
```

**Características Visuales:**
```tsx
// Plan destacado
<div className={`
    relative p-8 rounded-2xl shadow-xl
    ${isPro ? 'border-4 border-indigo-600' : 'border border-gray-200'}
    ${isPro ? 'scale-105 z-10' : ''}
    hover:scale-105 transition-transform duration-300
`}>
    {isPro && (
        <div className="absolute top-0 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
            <span className="bg-gradient-to-r from-indigo-600 to-purple-600 text-white px-4 py-1 rounded-full text-sm font-semibold">
                Más Popular
            </span>
        </div>
    )}
    
    {/* Contenido del plan */}
</div>
```

#### **BillingPage.tsx** (Página Protegida)
- ✅ Dashboard de suscripción actual
- ✅ Badge de estado (Activa/Cancelada/Expirada)
- ✅ Información del plan (fechas, auto-renovación)
- ✅ Lista de características incluidas
- ✅ Botón "Actualizar Plan" → redirige a /pricing
- ✅ Botón "Cancelar Suscripción" con confirmación
- ✅ Vista placeholder si no hay suscripción
- ✅ Toast de error 402 al ser redirigido
- ✅ Sidebar con información de pago

**Estado de Suscripción:**
```typescript
interface Subscription {
    id: number;
    user_id: number;
    plan_type: string;
    status: 'pending' | 'active' | 'cancelled' | 'expired';
    mercadopago_id: string;
    start_date: string;
    end_date: string;
    auto_renew: boolean;
    cancelled_at?: string;
    plan?: Plan;
}

const fetchSubscription = async () => {
    try {
        const response = await api.get('/subscriptions/status');
        setSubscription(response.data);
    } catch (error) {
        // Manejar errores
    }
};

const handleCancelSubscription = async () => {
    if (!window.confirm('¿Estás seguro de cancelar tu suscripción?')) return;
    
    try {
        await api.post('/subscriptions/cancel', {
            subscription_id: subscription.id
        });
        
        toast.success('Suscripción cancelada');
        fetchSubscription(); // Recargar datos
        
    } catch (error) {
        toast.error('Error al cancelar');
    }
};
```

**Badges de Estado:**
```typescript
const getStatusBadge = (status: string) => {
    const badges = {
        active: 'bg-green-100 text-green-800',
        pending: 'bg-yellow-100 text-yellow-800',
        cancelled: 'bg-orange-100 text-orange-800',
        expired: 'bg-red-100 text-red-800'
    };
    
    const labels = {
        active: 'Activa',
        pending: 'Pendiente',
        cancelled: 'Cancelada',
        expired: 'Expirada'
    };
    
    return (
        <span className={`px-3 py-1 rounded-full text-sm font-semibold ${badges[status]}`}>
            {labels[status]}
        </span>
    );
};
```

**Rutas Agregadas:**
```typescript
// App.tsx
const router = createBrowserRouter([
    {
        path: '/',
        element: <MainLayout />,
        children: [
            // Rutas públicas
            { path: 'login', element: <LoginPage /> },
            { path: 'register', element: <RegisterPage /> },
            { path: 'pricing', element: <PricingPage /> }, // ✨ NUEVA
            
            // Rutas protegidas
            {
                element: <ProtectedRoute />,
                children: [
                    { path: 'dashboard', element: <DashboardPage /> },
                    { path: 'products', element: <ProductsPage /> },
                    { path: 'billing', element: <BillingPage /> }, // ✨ NUEVA
                    // ... otras rutas
                ]
            }
        ]
    }
]);
```

**Navegación en Sidebar:**
```tsx
// Sidebar.tsx
<nav>
    {/* Sección de Ventas */}
    <NavLink to="/products">📦 Productos</NavLink>
    <NavLink to="/sales-orders">🛒 Órdenes de Venta</NavLink>
    
    <div className="border-t border-gray-600 my-2"></div>
    
    {/* Sección de Suscripción */}
    <NavLink to="/billing">💳 Mi Suscripción</NavLink>
    <NavLink to="/pricing">💎 Ver Planes</NavLink>
    
    <div className="border-t border-gray-600 my-2"></div>
    
    {/* Otras secciones */}
</nav>
```

**Interceptor 402 Payment Required:**
```typescript
// services/api.ts
api.interceptors.response.use(
    (response) => response,
    (error) => {
        // Error 401: No autenticado
        if (error.response?.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login';
        }
        
        // Error 402: Payment Required
        if (error.response?.status === 402) {
            const errorData = error.response.data;
            
            // Guardar mensaje en sessionStorage
            sessionStorage.setItem('paymentRequired', JSON.stringify({
                message: errorData.error || 'Suscripción requerida',
                upgrade_url: errorData.upgrade_url || '/pricing'
            }));
            
            // Redirigir a billing (evitar loops)
            if (!window.location.pathname.includes('/billing') && 
                !window.location.pathname.includes('/pricing')) {
                window.location.href = '/billing';
            }
        }
        
        return Promise.reject(error);
    }
);
```

**Features Implementadas:**

| Feature | PricingPage | BillingPage |
|---------|-------------|-------------|
| Ver planes disponibles | ✅ | ❌ |
| Comparar características | ✅ | ❌ |
| Iniciar suscripción | ✅ | ❌ |
| Ver suscripción actual | ❌ | ✅ |
| Cancelar suscripción | ❌ | ✅ |
| Actualizar plan | ❌ | ✅ |
| Badge de estado | ❌ | ✅ |
| Auto-redirect en 402 | ✅ | ✅ |
| Toast de errores | ✅ | ✅ |

---

### **17. Frontend de Suscripciones (La Vidriera)**

**Tecnología:** HMAC-SHA256

**Implementación:**
```go
// Verificar firma de MercadoPago
func VerifyMercadoPagoSignature(xSignature, xRequestID, dataID string) bool {
    secret := os.Getenv("MERCADOPAGO_WEBHOOK_SECRET")
    
    // Generar firma esperada
    manifest := fmt.Sprintf("id:%s;request-id:%s", dataID, xRequestID)
    expectedSignature := GenerateHMAC(manifest, secret)
    
    // Extraer firma del header (formato: "ts=123,v1=abc")
    parts := strings.Split(xSignature, ",")
    var receivedSignature string
    for _, part := range parts {
        if strings.HasPrefix(part, "v1=") {
            receivedSignature = strings.TrimPrefix(part, "v1=")
            break
        }
    }
    
    // Comparación segura (constant-time)
    return hmac.Equal(
        []byte(receivedSignature),
        []byte(expectedSignature)
    )
}

func GenerateHMAC(message, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(message))
    return hex.EncodeToString(h.Sum(nil))
}
```

**Validaciones de Seguridad:**
```go
// Handler de webhook
func HandleMercadoPagoWebhook(w http.ResponseWriter, r *http.Request) {
    // 1. Verificar método HTTP
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // 2. Obtener headers de firma
    xSignature := r.Header.Get("x-signature")
    xRequestID := r.Header.Get("x-request-id")
    
    if xSignature == "" || xRequestID == "" {
        http.Error(w, "Missing signature headers", http.StatusBadRequest)
        return
    }
    
    // 3. Leer body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Cannot read body", http.StatusBadRequest)
        return
    }
    
    // 4. Parsear JSON
    var payload WebhookPayload
    if err := json.Unmarshal(body, &payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // 5. Verificar firma
    if !VerifyMercadoPagoSignature(xSignature, xRequestID, payload.Data.ID) {
        slog.Error("Invalid webhook signature", "request_id", xRequestID)
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // 6. Procesar webhook (idempotente)
    ProcessWebhook(payload)
    
    // 7. Responder rápido (MercadoPago espera 200 en <2s)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}
```

**Prevención de Ataques:**
- ✅ Verificación de firma HMAC
- ✅ Validación de headers requeridos
- ✅ Procesamiento idempotente (evita duplicados)
- ✅ Timeout corto en respuesta
- ✅ Logging de intentos fallidos
- ✅ Rate limiting (preparado)

---

### **18. Seguridad en Webhooks**

**Tablas Nuevas:**

```sql
-- Tabla de suscripciones
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_type TEXT NOT NULL CHECK (plan_type IN ('basico', 'pro', 'enterprise')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'active', 'cancelled', 'expired')),
    mercadopago_id TEXT,
    start_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date TIMESTAMPTZ NOT NULL,
    auto_renew BOOLEAN NOT NULL DEFAULT true,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_end_date ON subscriptions(end_date);
CREATE INDEX idx_subscriptions_mercadopago_id ON subscriptions(mercadopago_id);

-- Trigger para updated_at
CREATE OR REPLACE FUNCTION update_subscriptions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_subscriptions_updated_at();

-- Constraint: Solo una suscripción activa por usuario
CREATE UNIQUE INDEX idx_one_active_subscription_per_user 
ON subscriptions(user_id) 
WHERE status = 'active';
```

**Consultas Optimizadas:**
```sql
-- Obtener suscripción activa de un usuario
SELECT * FROM subscriptions
WHERE user_id = $1 
  AND status = 'active'
  AND end_date > NOW()
LIMIT 1;

-- Buscar suscripciones próximas a vencer (renovación)
SELECT * FROM subscriptions
WHERE status = 'active'
  AND auto_renew = true
  AND end_date BETWEEN NOW() AND NOW() + INTERVAL '3 days';

-- Expirar suscripciones vencidas (job diario)
UPDATE subscriptions
SET status = 'expired'
WHERE status = 'active'
  AND end_date < NOW();

-- Estadísticas de suscripciones
SELECT 
    plan_type,
    status,
    COUNT(*) as total,
    SUM(CASE WHEN auto_renew THEN 1 ELSE 0 END) as with_auto_renew
FROM subscriptions
GROUP BY plan_type, status;
```

---

## 🎯 Funcionalidades Destacadas (Actualizado)

### **Top 16 Features**

1. **Sistema Multi-Tenant con Organizaciones** ⭐⭐⭐ **[NUEVO - Nov 2025]**
   - Admins con su propia organización
   - Vendedores/Repositores comparten inventario del admin
   - Aislamiento completo de datos entre organizaciones
   - Migración completa sin pérdida de datos

2. **Sistema de Lotes con FEFO** ⭐⭐⭐
   - Minimiza pérdidas por vencimiento
   - Trazabilidad completa
   - Cumplimiento normativo

3. **Sistema de Suscripciones Multi-Plan** ⭐⭐⭐
   - 3 planes configurables
   - Integración MercadoPago
   - Renovación automática
   - Webhooks seguros

4. **Paywall Middleware Inteligente** ⭐⭐⭐
   - 42 endpoints protegidos
   - Validación de límites por plan
   - Respuestas 402 personalizadas
   - Verificación de features

5. **Frontend de Suscripciones** ⭐⭐
   - Página de precios moderna
   - Dashboard de facturación
   - Auto-redirect en 402
   - Gestión completa

6. **Autenticación JWT + RBAC**
   - 3 roles configurables
   - Seguridad multicapa
   - Tokens con expiración
   - Organization ID en JWT

7. **Transacciones ACID con Locks**
   - Previene stock negativo
   - Seguridad en concurrencia
   - FOR UPDATE locks

8. **Webhooks Seguros** ⭐⭐
   - Verificación HMAC-SHA256
   - Procesamiento idempotente
   - Respuesta rápida (<2s)

9. **Auditoría Completa**
   - Log de todas las operaciones
   - Trazabilidad de cambios
   - IP tracking

10. **Sistema de Notificaciones**
    - Emails asíncronos
    - RabbitMQ queue
    - SendGrid integration

11. **Reportes y Exportación**
    - CSV on-demand
    - Filtros avanzados
    - Datos completos

12. **Dashboard con KPIs**
    - Métricas en tiempo real
    - Stock bajo automático
    - Visualización clara
    - Filtrado por organización

13. **Logging Estructurado**
    - JSON output
    - Búsqueda fácil
    - Debugging eficiente

14. **Encriptación de Datos Sensibles**
    - AES-256-GCM
    - API keys seguras
    - Credenciales protegidas

15. **Límites por Plan Dinámicos** ⭐
    - Productos máximos
    - Órdenes mensuales
    - Features condicionales

16. **Gestión de Usuarios Multi-Organización** ⭐ **[NUEVO]**
    - Admins pueden crear vendedores/repositores
    - Compartición automática de inventario
    - Permisos heredados por organización

---

## 📚 Documentación Generada

| Documento | Páginas | Contenido |
|-----------|---------|-----------|
| **TAREA_1_COMPLETADA.md** | 10 | Migración de base de datos |
| **TAREA_2_COMPLETADA.md** | 8 | Refactorización de lectura |
| **TAREA_3_COMPLETADA.md** | 12 | Entrada de stock (compras) |
| **TAREA_4_COMPLETADA.md** | 15 | Salida FEFO (ventas) |
| **RESUMEN_LOTES_COMPLETO.md** | 10 | Diagramas y flujos |
| **PROYECTO_LOTES_FINAL.md** | 20 | Resumen completo del proyecto |
| **GUIA_PRUEBAS_FEFO.md** | 15 | Testing paso a paso |
| **TAREA_3.1_COMPLETADA.md** | 12 | Base de datos de suscripciones |
| **TAREA_3.2_COMPLETADA.md** | 18 | Webhook de MercadoPago |
| **TAREA_4_PATOVICA_COMPLETADA.md** | 20 | Middleware de Paywall |
| **TAREA_5_VIDRIERA_COMPLETADA.md** | 25 | Frontend de suscripciones |
| **RESUMEN_TAREA_5.md** | 8 | Resumen ejecutivo Tarea 5 |
| **BUG_FIXES_ORDENES_COMPRA.md** | 10 | Fixes en órdenes de compra |
| **TAREA-3-DELEGACION.md** | 15 | Sistema multi-tenant con organizaciones |
| **IMPLEMENTACIONES.md** | 35 | Este documento |

**Total:** ~218 páginas de documentación técnica

---

## 🎓 Tecnologías Aprendidas e Implementadas

### **Backend**
- ✅ Go (Golang) - Programación concurrente
- ✅ Chi Router - HTTP routing
- ✅ PostgreSQL - Base de datos relacional
- ✅ pgx/v5 - Driver nativo Go
- ✅ JWT - Autenticación stateless
- ✅ bcrypt - Hashing seguro
- ✅ AES-256 - Encriptación simétrica
- ✅ RabbitMQ - Message broker
- ✅ Goroutines - Concurrencia
- ✅ Channels - Comunicación entre goroutines

### **Frontend**
- ✅ React 18 - Framework UI
- ✅ TypeScript - Tipado estático
- ✅ Vite - Build tool moderno
- ✅ Tailwind CSS - Utility-first CSS
- ✅ React Router - SPA routing
- ✅ React Hook Form - Formularios
- ✅ Axios - HTTP client

### **DevOps**
- ✅ Docker - Contenedores
- ✅ Docker Compose - Orquestación
- ✅ Nginx - Web server
- ✅ Git - Control de versiones

### **Arquitectura**
- ✅ Microservicios (API, Worker, Scheduler)
- ✅ REST API
- ✅ ACID Transactions
- ✅ FEFO Algorithm
- ✅ RBAC Pattern
- ✅ Repository Pattern
- ✅ Middleware Pattern

### **Base de Datos**
- ✅ Migraciones SQL
- ✅ Indices optimizados
- ✅ Foreign Keys
- ✅ Row-level locking
- ✅ Triggers (preparado)
- ✅ JSONB (para config)

---

## 🚀 Roadmap Futuro (Sugerencias)

### **Mejoras de Producto**

1. **Frontend Avanzado**
   - [ ] Vista de lotes por producto
   - [ ] Gráficos de vencimientos
   - [ ] Dashboard mejorado con Chart.js
   - [ ] Modo oscuro

2. **Reportes Avanzados**
   - [ ] Lotes próximos a vencer
   - [ ] Análisis de rotación de inventario
   - [ ] Predicción de demanda
   - [ ] Reportes visuales (PDF)

3. **Integraciones**
   - [ ] API de facturación electrónica
   - [ ] Pagos con Stripe/Mercado Pago
   - [ ] Envío de facturas automáticas
   - [ ] Importación desde Excel

4. **Mobile App**
   - [ ] React Native
   - [ ] Escaneo de códigos de barras
   - [ ] Notificaciones push

### **Mejoras Técnicas**

1. **Performance**
   - [ ] Redis para caching
   - [ ] CDN para frontend
   - [ ] Optimización de queries
   - [ ] Paginación en todas las listas

2. **Seguridad**
   - [ ] Rate limiting
   - [ ] 2FA (Two-Factor Auth)
   - [ ] Refresh tokens
   - [ ] WAF (Web Application Firewall)

3. **Monitoreo**
   - [ ] Sentry para error tracking
   - [ ] Prometheus + Grafana
   - [ ] Alertas automáticas
   - [ ] Health checks avanzados

4. **Testing**
   - [ ] Unit tests (Go)
   - [ ] Integration tests
   - [ ] E2E tests (Cypress/Playwright)
   - [ ] CI/CD pipeline

---

## 📞 Resumen Técnico

### **Lenguajes Utilizados**

| Lenguaje | % Uso | Propósito |
|----------|-------|-----------|
| **Go** | 55% | Backend completo (API + Multi-Tenant + Suscripciones + Worker) |
| **TypeScript** | 30% | Frontend (Dashboard + Pricing + Billing + Organizations) |
| **SQL** | 10% | Base de datos (17 migraciones) |
| **Bash** | 3% | Scripts |
| **Markdown** | 2% | Documentación |

### **Líneas de Código (Estimado)**

```
Backend (Go):        ~7,500 líneas  (+2,500 suscripciones +500 multi-tenant)
Frontend (TS/TSX):   ~4,500 líneas  (+1,500 pricing/billing)
SQL Migrations:      ~1,200 líneas  (+200 subscriptions +100 organizations)
Documentación:       ~12,000 líneas (+4,000 nuevas tareas +500 multi-tenant)
──────────────────────────────────────────────────────────────────────────────
Total:               ~26,300 líneas (+9,300 nuevas desde inicio)
```

### **Complejidad del Proyecto**

- **Complejidad Técnica:** Alta ⭐⭐⭐⭐⭐
- **Complejidad de Negocio:** Alta ⭐⭐⭐⭐
- **Escalabilidad:** Alta ⭐⭐⭐⭐⭐
- **Mantenibilidad:** Alta ⭐⭐⭐⭐⭐
- **Documentación:** Excelente ⭐⭐⭐⭐⭐

---

## 🏆 Logros del Proyecto

### **Técnicos**
- ✅ Arquitectura de microservicios funcional
- ✅ Sistema de lotes con FEFO implementado desde cero
- ✅ **Sistema Multi-Tenant con Organizaciones** ⭐ **[NUEVO - Nov 2025]**
- ✅ **Sistema de Suscripciones Multi-Plan**
- ✅ **Integración con MercadoPago** (checkout + webhooks)
- ✅ **Paywall Middleware** (42 rutas protegidas)
- ✅ **Frontend de Pagos** (2 páginas nuevas)
- ✅ **Webhooks Seguros** con verificación HMAC
- ✅ Transacciones ACID con locks avanzados
- ✅ Encriptación de datos sensibles
- ✅ Logging estructurado completo
- ✅ **17 migraciones de base de datos** (incluyendo organizaciones)
- ✅ **JWT con organization_id** en claims
- ✅ **18+ handlers actualizados** para multi-tenant
- ✅ RBAC con 3 roles
- ✅ **Compartición automática de inventario** entre usuarios de una organización

### **De Negocio**
- ✅ Trazabilidad completa de inventario
- ✅ **Sistema de equipos de trabajo** (admins + vendedores/repositores) ⭐ **[NUEVO]**
- ✅ **Monetización con planes de suscripción**
- ✅ **Límites configurables por plan**
- ✅ **Pasarela de pagos integrada**
- ✅ **Compartición de inventario** entre usuarios de una organización ⭐
- ✅ **Aislamiento completo** entre organizaciones diferentes ⭐
- ✅ Minimización de pérdidas por vencimiento
- ✅ Cumplimiento de normativas sanitarias
- ✅ Reportes exportables
- ✅ Sistema de alertas automático
- ✅ Dashboard con KPIs por organización

### **Operacionales**
- ✅ Zero downtime en migraciones
- ✅ Sin pérdida de datos
- ✅ Rollback seguro disponible
- ✅ Docker Compose para fácil deployment
- ✅ Documentación exhaustiva

---

## 🎯 Conclusión

**Stock In Order** es un sistema de gestión de inventario de **nivel empresarial** que implementa:

- ✅ **15+ tecnologías** principales
- ✅ **19 módulos funcionales** completos (incluyendo multi-tenant)
- ✅ **14 tablas** de base de datos optimizadas
- ✅ **50+ endpoints** REST API
- ✅ **3 servicios** backend (API, Worker, Scheduler)
- ✅ **Sistema Multi-Tenant** con organizaciones ⭐ **[NUEVO]**
- ✅ **Sistema de Suscripciones** con MercadoPago
- ✅ **Paywall Middleware** (42 rutas protegidas)
- ✅ **Frontend de Pagos** (PricingPage + BillingPage)
- ✅ **Webhooks Seguros** con HMAC-SHA256
- ✅ **FEFO Algorithm** para rotación óptima
- ✅ **RBAC** con autenticación JWT (incluye organization_id)
- ✅ **Auditoría** completa de operaciones
- ✅ **Transacciones ACID** con locks
- ✅ **Documentación** de 218+ páginas

El proyecto está **listo para producción** y preparado para escalar.

---

**Autor:** Stock In Order Team  
**Versión del Documento:** 3.0 *(Actualizado con Sistema Multi-Tenant)*  
**Fecha:** 22 de Noviembre, 2025  
**Estado del Proyecto:** ✅ EN PRODUCCIÓN

---

### **19. Base de Datos de Suscripciones**

---

## 📖 Referencias y Recursos

### **Documentación Oficial**
- [Go Documentation](https://go.dev/doc/)
- [React Documentation](https://react.dev/)
- [PostgreSQL Manual](https://www.postgresql.org/docs/)
- [Chi Router](https://github.com/go-chi/chi)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/tutorials)

### **Best Practices**
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [REST API Design](https://restfulapi.net/)
- [Database Indexing](https://use-the-index-luke.com/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

---

**¡Gracias por usar Stock In Order! 🚀**
