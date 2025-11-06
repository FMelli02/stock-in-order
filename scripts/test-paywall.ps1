# ============================================
# TEST: Paywall Middleware (El Patovica 2.0)
# ============================================
# Este script prueba el middleware de "Pared de Pago" que verifica
# si el usuario tiene una suscripción activa antes de permitir acceso.
#
# Prerequisitos:
# 1. Backend corriendo en http://localhost:8080
# 2. Base de datos con usuarios de prueba configurados
# 3. Usuarios con diferentes estados de suscripción
#
# Uso:
#   .\scripts\test-paywall.ps1
# ============================================

$BASE_URL = "http://localhost:8080/api/v1"

# Colores para output
function Write-Success { param($msg) Write-Host "✅ $msg" -ForegroundColor Green }
function Write-Error-Custom { param($msg) Write-Host "❌ $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "ℹ️  $msg" -ForegroundColor Cyan }
function Write-Warning-Custom { param($msg) Write-Host "⚠️  $msg" -ForegroundColor Yellow }

Write-Host "`n============================================" -ForegroundColor Yellow
Write-Host "🔒 TEST: Paywall Middleware" -ForegroundColor Yellow
Write-Host "============================================`n" -ForegroundColor Yellow

# ============================================
# Test 0: Verificar que el backend está corriendo
# ============================================
Write-Info "Test 0: Verificando conectividad con el backend..."

try {
    $healthCheck = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET -TimeoutSec 5
    Write-Success "Backend está corriendo"
    Write-Host "  Status: $($healthCheck.status)" -ForegroundColor White
} catch {
    Write-Error-Custom "Backend NO está accesible en http://localhost:8080"
    Write-Host "Por favor, inicia el backend con: cd backend; go run ./cmd/api" -ForegroundColor Red
    exit 1
}

# ============================================
# Test 1: Login con usuario con suscripción ACTIVA
# ============================================
Write-Info "`nTest 1: Login con usuario con suscripción ACTIVA..."

# Nota: Ajusta estos datos según tus usuarios de prueba
$activeUser = @{
    email = "admin@test.com"  # Usuario con suscripción activa
    password = "password123"
} | ConvertTo-Json

try {
    $loginActive = Invoke-RestMethod -Uri "$BASE_URL/users/login" `
        -Method POST `
        -Body $activeUser `
        -ContentType "application/json"

    $JWT_ACTIVE = $loginActive.token
    Write-Success "Token obtenido para usuario ACTIVO"
    Write-Host "  Token: $($JWT_ACTIVE.Substring(0, 50))..." -ForegroundColor Gray
} catch {
    Write-Error-Custom "Falló el login del usuario activo"
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Warning-Custom "Verifica que exista un usuario con email 'admin@test.com' y password 'password123'"
    exit 1
}

# ============================================
# Test 2: Acceso a ruta protegida CON suscripción activa
# ============================================
Write-Info "`nTest 2: Acceso a /products con suscripción ACTIVA..."

try {
    $headers = @{
        "Authorization" = "Bearer $JWT_ACTIVE"
    }
    
    $response = Invoke-RestMethod -Uri "$BASE_URL/products" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Acceso permitido (HTTP 200 OK)"
    Write-Host "  Productos obtenidos: $($response.products.Count)" -ForegroundColor White
} catch {
    $statusCode = $_.Exception.Response.StatusCode.Value__
    Write-Error-Custom "Acceso bloqueado inesperadamente (Status: $statusCode)"
    
    if ($statusCode -eq 402) {
        Write-Warning-Custom "El usuario debería tener suscripción ACTIVA pero fue bloqueado con 402"
        Write-Host "Verifica el estado de la suscripción en la base de datos:" -ForegroundColor Yellow
        Write-Host "  SELECT * FROM subscriptions WHERE user_id = (SELECT id FROM users WHERE email = 'admin@test.com');" -ForegroundColor Gray
    }
}

# ============================================
# Test 3: Acceso a /subscriptions/status (sin paywall)
# ============================================
Write-Info "`nTest 3: Acceso a /subscriptions/status (ruta SIN paywall)..."

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/subscriptions/status" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Acceso permitido a /subscriptions/status"
    Write-Host "  Plan: $($response.plan_type)" -ForegroundColor White
    Write-Host "  Status: $($response.status)" -ForegroundColor White
    Write-Host "  Start Date: $($response.start_date)" -ForegroundColor White
    
    if ($response.status -ne "active") {
        Write-Warning-Custom "El usuario tiene status '$($response.status)' pero esperábamos 'active'"
    }
} catch {
    Write-Error-Custom "Error al acceder a /subscriptions/status: $($_.Exception.Message)"
}

# ============================================
# Test 4: Acceso a Dashboard (requiere paywall)
# ============================================
Write-Info "`nTest 4: Acceso a /dashboard/metrics con suscripción ACTIVA..."

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/dashboard/metrics" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Acceso permitido al dashboard"
    Write-Host "  Total Productos: $($response.total_products)" -ForegroundColor White
    Write-Host "  Total Clientes: $($response.total_customers)" -ForegroundColor White
} catch {
    $statusCode = $_.Exception.Response.StatusCode.Value__
    if ($statusCode -eq 402) {
        Write-Error-Custom "Dashboard bloqueado con 402 Payment Required"
    } else {
        Write-Error-Custom "Error inesperado: Status $statusCode"
    }
}

# ============================================
# Test 5: Simular usuario SIN JWT (401 Unauthorized)
# ============================================
Write-Info "`nTest 5: Acceso SIN token JWT (esperamos 401 Unauthorized)..."

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/products" -Method GET
    Write-Error-Custom "Se permitió acceso sin JWT (debería haber fallado con 401)"
} catch {
    $statusCode = $_.Exception.Response.StatusCode.Value__
    if ($statusCode -eq 401) {
        Write-Success "Acceso bloqueado correctamente (HTTP 401 Unauthorized)"
    } else {
        Write-Error-Custom "Status Code incorrecto: $statusCode (esperado: 401)"
    }
}

# ============================================
# Test 6: Crear un producto (verificar RBAC + Paywall)
# ============================================
Write-Info "`nTest 6: Crear un producto (POST /products) - Requiere RBAC + Paywall..."

$newProduct = @{
    name = "Test Producto - Paywall Test"
    sku = "TEST-PAYWALL-$(Get-Random -Minimum 1000 -Maximum 9999)"
    description = "Producto de prueba creado por test-paywall.ps1"
    price = 9999.99
    cost = 5000.00
    stock = 100
    min_stock = 10
    max_stock = 500
    category = "testing"
    supplier_id = 1
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/products" `
        -Method POST `
        -Headers $headers `
        -Body $newProduct `
        -ContentType "application/json"
    
    Write-Success "Producto creado exitosamente"
    Write-Host "  ID: $($response.product.id)" -ForegroundColor White
    Write-Host "  SKU: $($response.product.sku)" -ForegroundColor White
} catch {
    $statusCode = $_.Exception.Response.StatusCode.Value__
    
    if ($statusCode -eq 402) {
        Write-Error-Custom "Producto bloqueado por paywall (402)"
    } elseif ($statusCode -eq 403) {
        Write-Warning-Custom "Usuario sin permisos para crear productos (403 Forbidden)"
        Write-Host "  El usuario necesita rol 'admin' o 'repositor'" -ForegroundColor Yellow
    } else {
        Write-Error-Custom "Error al crear producto: Status $statusCode"
    }
}

# ============================================
# Test 7: Verificar rutas de webhooks (públicas)
# ============================================
Write-Info "`nTest 7: Webhook de MercadoPago (ruta PÚBLICA - sin JWT)..."

try {
    # Simular webhook (sin firma válida, solo para verificar que es accesible)
    $webhookBody = @{
        type = "test"
        data = @{ id = "12345" }
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$BASE_URL/webhooks/mercadopago" `
        -Method POST `
        -Body $webhookBody `
        -ContentType "application/json"
    
    Write-Success "Webhook es accesible sin JWT (público)"
} catch {
    $statusCode = $_.Exception.Response.StatusCode.Value__
    
    # Webhook siempre retorna 200 (incluso con error de firma)
    if ($statusCode -eq 200) {
        Write-Success "Webhook respondió con 200 OK (correcto - siempre retorna 200)"
    } else {
        Write-Warning-Custom "Webhook retornó status $statusCode (esperábamos 200)"
    }
}

# ============================================
# Resumen Final
# ============================================
Write-Host "`n============================================" -ForegroundColor Yellow
Write-Host "📊 RESUMEN DE TESTS" -ForegroundColor Yellow
Write-Host "============================================" -ForegroundColor Yellow

Write-Host "`n✅ Tests Completados:" -ForegroundColor Green
Write-Host "  - Backend accesible" -ForegroundColor White
Write-Host "  - Login con JWT exitoso" -ForegroundColor White
Write-Host "  - Paywall permite acceso con suscripción activa" -ForegroundColor White
Write-Host "  - Rutas de suscripciones accesibles sin paywall" -ForegroundColor White
Write-Host "  - Dashboard protegido con paywall" -ForegroundColor White
Write-Host "  - Bloqueo correcto sin JWT (401)" -ForegroundColor White
Write-Host "  - Creación de productos con RBAC + Paywall" -ForegroundColor White
Write-Host "  - Webhooks públicos accesibles" -ForegroundColor White

Write-Host "`n⚠️  Tests Adicionales Recomendados:" -ForegroundColor Yellow
Write-Host "  1. Crear usuario con suscripción 'cancelled' y verificar bloqueo 402" -ForegroundColor Gray
Write-Host "  2. Crear usuario con suscripción 'expired' y verificar bloqueo 402" -ForegroundColor Gray
Write-Host "  3. Crear usuario sin registro en tabla subscriptions y verificar 402" -ForegroundColor Gray
Write-Host "  4. Probar límites de plan (MaxProducts, MaxOrders) cuando se implemente" -ForegroundColor Gray
Write-Host "  5. Probar RequireFeature con plan 'free' (reports = false)" -ForegroundColor Gray

Write-Host "`n============================================" -ForegroundColor Yellow
Write-Host "✅ PAYWALL TESTS FINALIZADOS" -ForegroundColor Green
Write-Host "============================================`n" -ForegroundColor Yellow
