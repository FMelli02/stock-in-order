# Tarea 4: El Cartero Digital - Documentación

## 🎯 Objetivo

Integrar **SendGrid** para que el Worker Service envíe los reportes generados por email a los usuarios de forma automática.

## 📊 Flujo Completo End-to-End

```
Usuario → Frontend → API → RabbitMQ → Worker → SendGrid → 📧 Email con Excel
```

### Flujo Detallado

1. **Usuario** hace clic en "Recibir por Email" (botón morado)
2. **Frontend** hace POST a `/api/v1/reports/products/email`
3. **API Backend** publica mensaje JSON a `reporting_queue`
4. **API** responde HTTP 202 Accepted inmediatamente
5. **Worker** consume el mensaje de RabbitMQ
6. **Worker** conecta a PostgreSQL y consulta datos
7. **Worker** genera archivo Excel con excelize
8. **Worker** prepara email con plantilla HTML
9. **Worker** envía email via SendGrid con Excel adjunto
10. **Usuario** recibe email en su bandeja de entrada

## 🛠️ Cambios Implementados

### 1. Nuevo Paquete de Email

#### **Creado: `worker/internal/email/sendgrid.go`**

```go
package email

// Client encapsula la configuración de SendGrid
type Client struct {
    apiKey     string
    fromEmail  string
    fromName   string
    sgClient   *sendgrid.Client
    isDisabled bool // Para desarrollo sin SendGrid
}

// EmailAttachment representa un archivo adjunto
type EmailAttachment struct {
    Filename    string // "reporte_productos.xlsx"
    Content     []byte // Contenido del archivo
    ContentType string // MIME type
}
```

**Características:**
- ✅ **Modo Desarrollo**: Funciona sin API Key configurada
- ✅ **Plantillas HTML**: Emails profesionales con gradientes
- ✅ **Base64 Encoding**: Archivos adjuntos correctamente codificados
- ✅ **Logging**: Registra cada envío exitoso o error
- ✅ **Error Handling**: Manejo robusto de errores de SendGrid

**Plantillas HTML por Tipo de Reporte:**

1. **Productos**: Gradiente morado (📦)
   ```html
   background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
   ```
   
2. **Clientes**: Gradiente rosa-rojo (👥)
   ```html
   background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
   ```

3. **Proveedores**: Gradiente azul claro (🏭)
   ```html
   background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
   ```

**Cada plantilla incluye:**
- Header con icono y título
- Mensaje personalizado
- Lista de contenido del reporte
- Footer con copyright
- Diseño responsive

### 2. Consumer Actualizado

#### **Modificado: `worker/internal/consumer/consumer.go`**

**Cambios en la firma:**
```go
// ANTES
func StartConsumer(rabbitURL string, db *pgxpool.Pool) error

// DESPUÉS
func StartConsumer(rabbitURL string, db *pgxpool.Pool, emailClient *email.Client) error
```

**Nueva lógica en processReport:**

```go
func processReport(db *pgxpool.Pool, emailClient *email.Client, req ReportRequest) error {
    // 1. Generar Excel (igual que antes)
    reportBytes, err := reports.GenerateProductsReport(db, req.UserID)
    
    // 2. Preparar adjunto
    attachment := email.EmailAttachment{
        Filename:    "reporte_productos.xlsx",
        Content:     reportBytes,
        ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    }
    
    // 3. Enviar email con SendGrid
    if err := emailClient.SendReportEmail(req.Email, "", req.ReportType, attachment); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }
    
    log.Printf("📧 Email enviado exitosamente a %s", req.Email)
    return nil
}
```

**Nombres de archivos generados:**
- `reporte_productos.xlsx`
- `reporte_clientes.xlsx`
- `reporte_proveedores.xlsx`

### 3. Main.go Actualizado

#### **Modificado: `worker/cmd/api/main.go`**

```go
// Configurar cliente de SendGrid
emailClient := email.NewClient(
    cfg.SendGrid_APIKey,
    "noreply@stockinorder.com", // Email remitente
    "Stock in Order",             // Nombre remitente
)
log.Println("📧 Cliente de email configurado")

// Pasar emailClient al consumer
go func() {
    if err := consumer.StartConsumer(cfg.RabbitMQ_URL, dbpool, emailClient); err != nil {
        errChan <- err
    }
}()
```

**Nota:** Si usas **Single Sender Verification** de SendGrid, debes cambiar el email remitente por el que verificaste.

### 4. Dependencias Actualizadas

#### **Modificado: `worker/go.mod`**

```go
require (
    github.com/jackc/pgx/v5 v5.7.1
    github.com/rabbitmq/amqp091-go v1.10.0
    github.com/sendgrid/sendgrid-go v3.16.1+incompatible  // ← NUEVO
    github.com/xuri/excelize/v2 v2.9.0
)

require (
    github.com/sendgrid/rest v2.6.9+incompatible  // ← Dependencia de SendGrid
    // ... otras dependencias
)
```

### 5. Variables de Entorno

#### **Modificado: `.env.example`**

```bash
# ========================================
# SendGrid API Key (para envío de emails)
# ========================================
# Para obtener tu API Key:
# 1. Crea una cuenta en https://sendgrid.com/
# 2. Ve a Settings > API Keys
# 3. Crea una nueva API Key con permisos de "Mail Send"
# 4. Copia la key y pégala aquí
#
# Nota: Si dejas este campo vacío, los emails NO se enviarán
# pero el sistema funcionará en "modo desarrollo" (solo logs)
SENDGRID_API_KEY=
```

#### **Ya configurado: `docker-compose.yml`**

```yaml
worker:
  environment:
    SENDGRID_API_KEY: "${SENDGRID_API_KEY:-}"  # Ya estaba configurado
```

## 🎨 Ejemplo de Email Enviado

### Asunto
```
📦 Tu Reporte de Productos está Listo
```

### Contenido HTML
```
┌─────────────────────────────────────┐
│   📦 Reporte de Productos           │  (Fondo gradiente morado)
└─────────────────────────────────────┘

¡Hola!

Tu reporte de Productos ha sido generado exitosamente 
y está adjunto en este email.

El archivo está en formato Excel (.xlsx) y contiene 
toda la información actualizada de tu inventario.

¿Qué incluye el reporte?
  ✅ Código y nombre de productos
  ✅ Descripción y categoría
  ✅ Precios actualizados
  ✅ Stock disponible
  ✅ Fechas de registro

Gracias por usar Stock in Order.

─────────────────────────────────────
Este es un email automático, por favor no responder.
Stock in Order © 2025
```

### Archivo Adjunto
```
📎 reporte_productos.xlsx (6.2 KB)
```

## 🔄 Modos de Operación

### Modo Producción (con SendGrid configurado)

**Requisitos:**
- Variable `SENDGRID_API_KEY` configurada
- Single Sender Verification completada en SendGrid
- Email remitente verificado

**Comportamiento:**
```
📧 Cliente de email configurado
📨 Mensaje recibido
🔨 Generando reporte: UserID=1, Email=usuario@ejemplo.com
📊 Reporte generado: 6403 bytes
✅ Email enviado exitosamente a usuario@ejemplo.com (código: 202)
✅ Reporte procesado exitosamente
```

**El usuario recibe:**
- Email profesional con HTML
- Archivo Excel adjunto
- En menos de 30 segundos

### Modo Desarrollo (sin SendGrid)

**Requisitos:**
- Variable `SENDGRID_API_KEY` vacía o no configurada

**Comportamiento:**
```
⚠️  SENDGRID_API_KEY no configurado. Los emails NO se enviarán.
📧 Cliente de email configurado
📨 Mensaje recibido
🔨 Generando reporte: UserID=1, Email=test@example.com
📊 Reporte generado: 6403 bytes
📧 [MODO DEV] Email simulado a test@example.com - Adjunto: reporte_productos.xlsx (6403 bytes)
📧 Email enviado exitosamente a test@example.com
✅ Reporte procesado exitosamente
```

**Ventajas:**
- ✅ No requiere cuenta SendGrid
- ✅ No consume cuota de emails
- ✅ Útil para testing y CI/CD
- ✅ El resto del sistema funciona normalmente

## 🧪 Testing

### Prueba 1: Modo Desarrollo (Sin SendGrid)

```bash
# 1. Asegúrate de que SENDGRID_API_KEY no esté configurada
# (debe estar comentada o vacía en .env)

# 2. Reconstruir servicios
docker compose down
docker compose up -d --build

# 3. Verificar modo desarrollo
docker logs stock_in_order_worker | grep "SENDGRID"
# Debe mostrar: "⚠️  SENDGRID_API_KEY no configurado"

# 4. Enviar mensaje de prueba
cd worker/test-publisher
go run main.go

# 5. Verificar logs
docker logs stock_in_order_worker --tail 10
# Debe mostrar: "[MODO DEV] Email simulado a..."
```

### Prueba 2: Modo Producción (Con SendGrid)

```bash
# 1. Crear archivo .env
cp .env.example .env

# 2. Configurar SendGrid API Key
nano .env
# Agregar: SENDGRID_API_KEY=SG.xxxxxx...

# 3. Actualizar email remitente en worker/cmd/api/main.go
# Cambiar "noreply@stockinorder.com" por tu email verificado

# 4. Reconstruir servicios
docker compose down
docker compose up -d --build

# 5. Verificar configuración
docker logs stock_in_order_worker | grep "SENDGRID"
# NO debe mostrar advertencia

# 6. Probar desde el frontend
# - Ir a http://localhost:5173
# - Login
# - Productos → "Recibir por Email"

# 7. Verificar logs
docker logs stock_in_order_worker -f
# Debe mostrar: "✅ Email enviado exitosamente (código: 202)"

# 8. Revisar bandeja de entrada
# El email debe llegar en < 30 segundos
```

### Prueba 3: Verificar Plantilla HTML

Para ver cómo se ve el email sin enviarlo:

```bash
# Crear archivo test-email.html con el contenido de una plantilla
# Abrir en el navegador para visualizar
```

## 📊 Comparación: Antes vs Después

| Aspecto | Tarea 3 | Tarea 4 |
|---------|---------|---------|
| **Generación** | ✅ Worker genera Excel | ✅ Worker genera Excel |
| **Envío Email** | ❌ TODO | ✅ SendGrid |
| **Adjunto** | ❌ No implementado | ✅ Excel adjunto |
| **Plantilla** | ❌ N/A | ✅ HTML profesional |
| **Modo Dev** | ❌ No considerado | ✅ Funciona sin API Key |
| **Usuario notificado** | ❌ No | ✅ Sí, por email |

## 🔐 Seguridad

### ✅ Implementado

1. **API Key en variable de entorno**: No hardcodeada
2. **Email validado**: SendGrid requiere verificación
3. **Adjuntos seguros**: Base64 encoding
4. **Modo desarrollo**: Sistema funciona sin credenciales

### ⚠️ Recomendaciones

1. **Nunca** subir `.env` a Git (ya en `.gitignore`)
2. **Rotar** API Key cada 90 días
3. **Usar** Restricted Access en SendGrid
4. **Monitorear** uso en SendGrid Dashboard
5. **Implementar** rate limiting si es público

## 🐛 Troubleshooting

### Error: "Error 401 Unauthorized"

**Síntoma:**
```
❌ Error al procesar reporte: failed to send email: error al enviar email: SendGrid respondió con código 401
```

**Causa:** API Key inválida

**Solución:**
1. Verifica que copiaste la API Key completa
2. Genera una nueva API Key en SendGrid
3. Actualiza el archivo `.env`
4. Reconstruye el worker: `docker compose up -d --build worker`

### Error: "Error 403 Forbidden"

**Síntoma:**
```
SendGrid respondió con código 403: Forbidden
```

**Causa:** Email remitente no verificado

**Solución:**
1. Completa Single Sender Verification en SendGrid
2. Verifica tu email haciendo clic en el link
3. Actualiza `worker/cmd/api/main.go` con el email verificado
4. Reconstruye: `docker compose up -d --build worker`

### El email no llega

**Síntoma:** Logs muestran envío exitoso pero no llega

**Diagnóstico:**
```bash
docker logs stock_in_order_worker | grep "Email enviado"
# Si muestra "✅ Email enviado exitosamente (código: 202)"
# entonces SendGrid lo recibió correctamente
```

**Solución:**
1. Revisa carpeta de **spam/correo no deseado**
2. Verifica en SendGrid Activity Feed
3. Si SendGrid lo envió pero no llega, el problema es tu proveedor de email
4. Considera usar Domain Authentication

### Worker en modo desarrollo cuando no debería

**Síntoma:**
```
⚠️  SENDGRID_API_KEY no configurado
```

**Causa:** Variable de entorno no cargada

**Solución:**
```bash
# Verificar que .env existe
ls -la .env

# Verificar contenido
cat .env | grep SENDGRID

# Reconstruir con --env-file
docker compose --env-file .env up -d --build worker
```

## 📚 Documentación Adicional

Ver **GUIA-SENDGRID.md** para:
- ✅ Paso a paso para crear cuenta SendGrid
- ✅ Cómo obtener API Key
- ✅ Configuración de Single Sender Verification
- ✅ Domain Authentication (avanzado)
- ✅ Troubleshooting detallado
- ✅ Límites del plan gratuito

## 🚀 Próximos Pasos

### Mejoras Futuras

1. **Sistema de Notificaciones** (Tarea 5)
   - WebSocket para notificar cuando el reporte esté listo
   - Barra de notificaciones en el frontend
   - Historial de reportes solicitados

2. **Templates Personalizables**
   - Permitir al admin personalizar plantillas
   - Logo de la empresa en emails
   - Colores corporativos

3. **Almacenamiento de Reportes**
   - Guardar reportes generados en S3/MinIO
   - Link de descarga en el email
   - Expiración después de 7 días

4. **Métricas y Monitoreo**
   - Dashboard de reportes enviados
   - Tasa de apertura de emails
   - Errores de envío
   - Tiempo promedio de generación

5. **Reportes con Filtros**
   - Extender a sales-orders con filtros de fecha
   - Extender a purchase-orders con filtros
   - Reportes personalizados por usuario

## ✅ Checklist de Implementación

- [x] Instalado SendGrid SDK en worker
- [x] Creado paquete `internal/email/sendgrid.go`
- [x] Implementado `Client` con modo desarrollo
- [x] Creadas 3 plantillas HTML (productos, clientes, proveedores)
- [x] Actualizado `consumer.go` para enviar emails
- [x] Modificado `main.go` para inicializar email client
- [x] Actualizado `go.mod` con dependencias
- [x] Agregado `SENDGRID_API_KEY` a `.env.example`
- [x] Probado en modo desarrollo (sin API Key)
- [x] Creada documentación: `GUIA-SENDGRID.md`
- [x] Servicios reconstruidos exitosamente
- [x] Test con `test-publisher` funcionando
- [ ] Configurado SendGrid API Key real (opcional)
- [ ] Probado envío real de email (requiere API Key)

## 📊 Logs de Ejemplo

### Modo Desarrollo
```
🚀 Iniciando Worker Service...
✅ Conectado a PostgreSQL
⚠️  SENDGRID_API_KEY no configurado. Los emails NO se enviarán.
📧 Cliente de email configurado
📬 Worker conectado a RabbitMQ
📨 Mensaje recibido: {"user_id":1,"email_to":"test@example.com","report_type":"products"}
🔨 Generando reporte: UserID=1, Email=test@example.com, Type=products
📊 Reporte generado: 6403 bytes
📧 [MODO DEV] Email simulado a test@example.com - Adjunto: reporte_productos.xlsx (6403 bytes)
📧 Email enviado exitosamente a test@example.com
✅ Reporte procesado exitosamente para UserID=1, ReportType=products
```

### Modo Producción (con SendGrid)
```
🚀 Iniciando Worker Service...
✅ Conectado a PostgreSQL
📧 Cliente de email configurado
📬 Worker conectado a RabbitMQ
📨 Mensaje recibido: {"user_id":1,"email_to":"usuario@ejemplo.com","report_type":"products"}
🔨 Generando reporte: UserID=1, Email=usuario@ejemplo.com, Type=products
📊 Reporte generado: 6403 bytes
✅ Email enviado exitosamente a usuario@ejemplo.com (código: 202)
📧 Email enviado exitosamente a usuario@ejemplo.com
✅ Reporte procesado exitosamente para UserID=1, ReportType=products
```

---

**Fecha de implementación:** 21 de Octubre 2025  
**Versión:** 4.0.0  
**Estado:** ✅ Completado y Operativo (Modo Desarrollo)  
**Próximo:** Configurar SendGrid API Key para producción
