# 🔍 Logging y Monitoreo de Errores - Guía de Configuración

## 📋 Descripción General

Este sistema implementa **logging estructurado** con `slog` (Go 1.21+) y **monitoreo de errores** con Sentry para tracking proactivo de problemas en producción.

---

## 🏗️ Componentes Implementados

### Backend (Go)

#### 1. **Logging Estructurado con slog**
- ✅ Logger JSON configurado en `cmd/api/main.go`
- ✅ Todos los logs en formato JSON estructurado
- ✅ Niveles: INFO, WARN, ERROR
- ✅ Timestamps en formato RFC3339
- ✅ Contexto enriquecido (método, path, duración, status code)

**Ejemplo de log:**
```json
{
  "time": "2025-10-17T20:03:51Z",
  "level": "INFO",
  "msg": "HTTP Request",
  "method": "POST",
  "path": "/api/v1/users/login",
  "status": 200,
  "duration_ms": 45,
  "bytes": 256,
  "remote_addr": "172.18.0.1:58392"
}
```

#### 2. **Middleware de Logging** (`middleware/logging.go`)
- Captura automática de todas las requests HTTP
- Métricas: método, path, status, duración, bytes, user-agent
- Response writer wrapper para capturar status codes

#### 3. **Middleware de Sentry** (`middleware/sentry.go`)
- Captura de panics con recuperación automática
- Envío automático de excepciones a Sentry
- Contexto de request añadido (método, path, headers)
- Respuesta 500 al cliente en caso de panic

#### 4. **Integración Sentry**
- SDK: `github.com/getsentry/sentry-go`
- Inicialización en `main.go`
- Configuración de environment y release
- Flush automático al cerrar aplicación

---

### Frontend (React + TypeScript)

#### 1. **Inicialización de Sentry** (`main.tsx`)
- SDK: `@sentry/react`
- Browser Tracing para performance monitoring
- Session Replay para reproducir errores del usuario
- Configuración por environment (dev/prod)

#### 2. **ErrorBoundary** (`App.tsx`)
- Captura de errores de renderizado de React
- Fallback UI elegante con detalles del error
- Botón de "Reintentar" para recovery
- Reporte automático a Sentry

#### 3. **Interceptor de Axios** (`services/api.ts`)
- Captura de errores de API (4xx, 5xx)
- Captura de errores de red (timeouts, conexión)
- Tags enriquecidos: endpoint, método HTTP, status code
- Extra context: request/response data

---

## 🚀 Configuración de Sentry

### Paso 1: Crear Proyectos en Sentry

1. Registrarse en [sentry.io](https://sentry.io)
2. Crear un proyecto para **Backend (Go)**
3. Crear un proyecto para **Frontend (React)**
4. Copiar los DSN de cada proyecto

### Paso 2: Configurar Variables de Entorno

Crear archivo `.env` en la raíz del proyecto:

```bash
# Backend Sentry DSN
SENTRY_DSN_BACKEND=https://xxxxxxxxxxxxx@o1234567.ingest.sentry.io/1234567

# Frontend Sentry DSN
SENTRY_DSN_FRONTEND=https://yyyyyyyyyyy@o1234567.ingest.sentry.io/7654321
```

### Paso 3: Reconstruir Contenedores

```bash
docker compose down
docker compose up --build -d
```

### Paso 4: Verificar en Sentry Dashboard

1. Ir a Sentry → Issues
2. Generar un error de prueba
3. Ver el error capturado con contexto completo

---

## 🧪 Testing de Captura de Errores

### Backend - Test de Panic

Agregar temporalmente en un handler:
```go
panic("Test error for Sentry")
```

Resultado esperado:
- ✅ Sentry captura el panic con stack trace
- ✅ Usuario recibe HTTP 500
- ✅ Logs muestran el error con contexto
- ✅ Aplicación continúa funcionando

### Frontend - Test de Error

Agregar temporalmente en un componente:
```typescript
throw new Error("Test error for Sentry")
```

Resultado esperado:
- ✅ ErrorBoundary muestra UI de fallback
- ✅ Sentry captura el error con stack trace
- ✅ Session replay disponible
- ✅ Usuario puede hacer "Reintentar"

### API Error - Test de Network

Hacer request a endpoint inválido:
```typescript
api.get('/invalid-endpoint')
```

Resultado esperado:
- ✅ Sentry captura el 404
- ✅ Tags: endpoint, método, status code
- ✅ Context: request data

---

## 📊 Estructura de Logs

### Backend Logs (JSON)

**Request Log:**
```json
{
  "time": "2025-10-17T20:05:30Z",
  "level": "INFO",
  "msg": "HTTP Request",
  "method": "GET",
  "path": "/api/v1/products",
  "status": 200,
  "duration_ms": 12,
  "bytes": 1024,
  "remote_addr": "172.18.0.1:58392",
  "user_agent": "Mozilla/5.0..."
}
```

**Error Log:**
```json
{
  "time": "2025-10-17T20:05:45Z",
  "level": "ERROR",
  "msg": "Error conectando a la base de datos",
  "error": "connection refused"
}
```

---

## 🔧 Configuración Avanzada

### Ajustar Sample Rates (Frontend)

En `frontend/src/main.tsx`:

```typescript
Sentry.init({
  // ...
  tracesSampleRate: 0.1,        // 10% de transacciones en producción
  replaysSessionSampleRate: 0.1, // 10% de sesiones normales
  replaysOnErrorSampleRate: 1.0, // 100% de sesiones con errores
})
```

### Ajustar Niveles de Log (Backend)

En `backend/cmd/api/main.go`:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
  Level: slog.LevelInfo, // Cambiar a LevelDebug, LevelWarn, etc.
}))
```

### Añadir Context a Sentry

Backend:
```go
sentry.CaptureException(err)
sentry.ConfigureScope(func(scope *sentry.Scope) {
  scope.SetTag("user_id", userId)
  scope.SetExtra("order_id", orderId)
})
```

Frontend:
```typescript
Sentry.setUser({ id: userId, email: userEmail })
Sentry.setTag("page", "checkout")
```

---

## 📈 Beneficios

### Proactivos
- ✅ **Detección temprana** de errores antes de que usuarios reporten
- ✅ **Alertas automáticas** por Slack/Email cuando hay errores críticos
- ✅ **Tendencias** de errores en tiempo real

### Debugging
- ✅ **Stack traces completos** con código fuente
- ✅ **Breadcrumbs** de las acciones previas al error
- ✅ **Session Replay** para ver exactamente qué hizo el usuario
- ✅ **Request/Response data** para reproducir bugs

### Performance
- ✅ **Transaction monitoring** para endpoints lentos
- ✅ **Duration metrics** en cada request
- ✅ **Identificación de bottlenecks**

---

## 🔒 Consideraciones de Seguridad

### ⚠️ NO loguear información sensible:
- ❌ Passwords
- ❌ JWT tokens completos
- ❌ Números de tarjetas de crédito
- ❌ PII (Personally Identifiable Information)

### ✅ Buenas prácticas:
- Usar `maskAllText: true` en Session Replay para producción
- Configurar `beforeSend` para filtrar datos sensibles
- Limitar sample rates en producción
- Configurar data scrubbing en Sentry dashboard

---

## 📚 Recursos Adicionales

- [Documentación slog](https://pkg.go.dev/log/slog)
- [Sentry Go SDK](https://docs.sentry.io/platforms/go/)
- [Sentry React SDK](https://docs.sentry.io/platforms/javascript/guides/react/)
- [Best Practices](https://docs.sentry.io/product/sentry-basics/integrate-backend/)

---

## 🎯 Estado Actual

✅ **Logging estructurado** funcionando con JSON  
✅ **Middleware de logging** capturando todas las requests  
✅ **Middleware de Sentry** recuperando panics  
✅ **ErrorBoundary React** capturando errores de UI  
✅ **Axios interceptor** capturando errores de API  
✅ **Variables de entorno** configuradas en docker-compose  
⚠️ **Sentry opcional** - sistema funciona sin DSN configurado  

---

## 💡 Próximos Pasos Sugeridos

1. **Alertas:** Configurar notificaciones en Sentry para errores críticos
2. **Dashboards:** Crear dashboards personalizados en Sentry
3. **Performance:** Añadir custom spans para operaciones específicas
4. **Context:** Enriquecer errores con más información de usuario/negocio
5. **Releases:** Configurar releases tracking para vincular errores con deploys
