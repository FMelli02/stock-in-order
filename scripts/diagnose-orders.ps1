# Script de Diagnóstico Completo para Problemas de Órdenes
# Verifica configuración de usuarios, roles, suscripciones y datos

Write-Host "`n╔════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║   DIAGNÓSTICO COMPLETO - STOCK IN ORDER          ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# 1. Verificar Docker
Write-Host "1️⃣  Verificando Docker..." -ForegroundColor Yellow
try {
    docker ps > $null 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "   ❌ Docker no está corriendo" -ForegroundColor Red
        exit 1
    }
    Write-Host "   ✅ Docker está corriendo" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Docker no está instalado" -ForegroundColor Red
    exit 1
}

# 2. Encontrar contenedor de PostgreSQL
Write-Host "`n2️⃣  Buscando contenedor de PostgreSQL..." -ForegroundColor Yellow
$pgContainer = docker ps --format "{{.Names}}" | Where-Object { $_ -match "postgres" } | Select-Object -First 1

if (-not $pgContainer) {
    Write-Host "   ❌ No se encontró contenedor de PostgreSQL" -ForegroundColor Red
    Write-Host "   💡 Ejecuta: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}
Write-Host "   ✅ Contenedor encontrado: $pgContainer" -ForegroundColor Green

# Función helper para ejecutar queries
function Invoke-PgQuery {
    param([string]$Query, [string]$Description)
    Write-Host "`n   $Description" -ForegroundColor Cyan
    $result = docker exec -i $pgContainer psql -U postgres -d stockinorder -c $Query 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host $result -ForegroundColor White
        return $result
    } else {
        Write-Host "   ❌ Error: $result" -ForegroundColor Red
        return $null
    }
}

# 3. Verificar usuarios y roles
Write-Host "`n3️⃣  Verificando usuarios y roles..." -ForegroundColor Yellow
$usersQuery = @"
SELECT 
    id,
    name,
    email,
    role,
    organization_id,
    created_at::DATE as created
FROM users 
ORDER BY id;
"@
Invoke-PgQuery -Query $usersQuery -Description "📋 Usuarios en el sistema:"

# 4. Verificar suscripciones
Write-Host "`n4️⃣  Verificando suscripciones..." -ForegroundColor Yellow
$subsQuery = @"
SELECT 
    u.email,
    u.role,
    s.plan_id,
    s.status,
    s.current_period_start::DATE as period_start,
    s.current_period_end::DATE as period_end
FROM users u
LEFT JOIN subscriptions s ON u.id = s.user_id
ORDER BY u.id;
"@
Invoke-PgQuery -Query $subsQuery -Description "💳 Estado de suscripciones:"

# 5. Verificar órdenes de venta y su relación con usuarios
Write-Host "`n5️⃣  Verificando órdenes de venta..." -ForegroundColor Yellow
$salesQuery = @"
SELECT 
    so.id as order_id,
    so.user_id,
    u.email as user_email,
    u.organization_id as user_org,
    so.order_date::DATE as date,
    so.status,
    COUNT(oi.id) as items_count
FROM sales_orders so
JOIN users u ON so.user_id = u.id
LEFT JOIN order_items oi ON so.id = oi.order_id
GROUP BY so.id, so.user_id, u.email, u.organization_id, so.order_date, so.status
ORDER BY so.id;
"@
Invoke-PgQuery -Query $salesQuery -Description "🛒 Órdenes de venta:"

# 6. Verificar órdenes de compra
Write-Host "`n6️⃣  Verificando órdenes de compra..." -ForegroundColor Yellow
$purchaseQuery = @"
SELECT 
    po.id as order_id,
    po.user_id,
    u.email as user_email,
    u.organization_id as user_org,
    po.order_date::DATE as date,
    po.status,
    COUNT(poi.id) as items_count
FROM purchase_orders po
JOIN users u ON po.user_id = u.id
LEFT JOIN purchase_order_items poi ON po.id = poi.purchase_order_id
GROUP BY po.id, po.user_id, u.email, u.organization_id, po.order_date, po.status
ORDER BY po.id;
"@
Invoke-PgQuery -Query $purchaseQuery -Description "📦 Órdenes de compra:"

# 7. Detectar problemas comunes
Write-Host "`n7️⃣  Analizando problemas potenciales..." -ForegroundColor Yellow

# 7.1 Usuarios sin suscripción activa
$noSubQuery = @"
SELECT email, role 
FROM users u 
WHERE NOT EXISTS (
    SELECT 1 FROM subscriptions s 
    WHERE s.user_id = u.id AND s.status = 'active'
);
"@
$noSubResult = docker exec -i $pgContainer psql -U postgres -d stockinorder -t -c $noSubQuery 2>&1
if ($noSubResult -and $noSubResult.Trim() -ne "") {
    Write-Host "   ⚠️  Usuarios SIN suscripción activa:" -ForegroundColor Red
    Write-Host $noSubResult -ForegroundColor Red
} else {
    Write-Host "   ✅ Todos los usuarios tienen suscripción activa" -ForegroundColor Green
}

# 7.2 Usuarios sin role admin o vendedor
$noRoleQuery = @"
SELECT email, role 
FROM users 
WHERE role NOT IN ('admin', 'vendedor');
"@
$noRoleResult = docker exec -i $pgContainer psql -U postgres -d stockinorder -t -c $noRoleQuery 2>&1
if ($noRoleResult -and $noRoleResult.Trim() -ne "") {
    Write-Host "`n   ⚠️  Usuarios sin acceso a órdenes (no son admin ni vendedor):" -ForegroundColor Red
    Write-Host $noRoleResult -ForegroundColor Red
} else {
    Write-Host "`n   ✅ Todos los usuarios pueden acceder a órdenes" -ForegroundColor Green
}

# 7.3 Organization ID mismatch
$orgMismatchQuery = @"
SELECT 
    'Sales Orders' as tipo,
    COUNT(*) as count
FROM sales_orders so
JOIN users u ON so.user_id = u.id
WHERE so.user_id != u.organization_id
UNION ALL
SELECT 
    'Purchase Orders' as tipo,
    COUNT(*) as count
FROM purchase_orders po
JOIN users u ON po.user_id = u.id
WHERE po.user_id != u.organization_id;
"@
$orgMismatch = docker exec -i $pgContainer psql -U postgres -d stockinorder -t -c $orgMismatchQuery 2>&1
Write-Host "`n   📊 Órdenes con organization_id diferente:" -ForegroundColor Cyan
Write-Host $orgMismatch -ForegroundColor White

# 8. Sugerencias de corrección
Write-Host "`n8️⃣  Sugerencias de corrección..." -ForegroundColor Yellow

Write-Host "`n   Para corregir problemas comunes, puedes ejecutar:" -ForegroundColor Cyan
Write-Host "   
   # Actualizar rol de usuario a 'vendedor':
   UPDATE users SET role = 'vendedor' WHERE email = 'tu@email.com';
   
   # Crear suscripción FREE para usuario:
   INSERT INTO subscriptions (user_id, plan_id, status, current_period_start)
   VALUES ((SELECT id FROM users WHERE email = 'tu@email.com'), 'free', 'active', NOW());
   
   # Sincronizar organization_id con user_id (para admins):
   UPDATE users SET organization_id = id WHERE role = 'admin' AND organization_id IS NULL;
" -ForegroundColor Gray

# 9. Resumen final
Write-Host "`n╔════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║   RESUMEN DEL DIAGNÓSTICO                        ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

Write-Host "📌 Puntos clave verificados:" -ForegroundColor Yellow
Write-Host "   • Usuarios y roles" -ForegroundColor Gray
Write-Host "   • Suscripciones activas" -ForegroundColor Gray
Write-Host "   • Órdenes de venta y compra" -ForegroundColor Gray
Write-Host "   • Problemas de permisos" -ForegroundColor Gray

Write-Host "`n💡 Para resolver errores en detalle de órdenes:" -ForegroundColor Yellow
Write-Host "   1. Asegúrate de tener rol 'admin' o 'vendedor'" -ForegroundColor Gray
Write-Host "   2. Verifica que tengas una suscripción activa" -ForegroundColor Gray
Write-Host "   3. Cierra sesión y vuelve a iniciar después de cambios" -ForegroundColor Gray
Write-Host "   4. Usa el script check-user-roles.ps1 para actualizar roles" -ForegroundColor Gray

Write-Host ""
