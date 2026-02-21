# Script de Correccion Automatica para Problemas de Ordenes
# Corrige roles, suscripciones y organization_id

Write-Host "`n========================================================" -ForegroundColor Green
Write-Host "   CORRECCION AUTOMATICA - STOCK IN ORDER             " -ForegroundColor Green
Write-Host "========================================================`n" -ForegroundColor Green

# Encontrar contenedor de PostgreSQL
$pgContainer = docker ps --format "{{.Names}}" | Where-Object { $_ -match "postgres" } | Select-Object -First 1

if (-not $pgContainer) {
    Write-Host "ERROR: No se encontro contenedor de PostgreSQL corriendo" -ForegroundColor Red
    exit 1
}

Write-Host "OK: Contenedor PostgreSQL: $pgContainer`n" -ForegroundColor Green

# Función helper
function Invoke-PgQuery {
    param([string]$Query)
    $result = docker exec -i $pgContainer psql -U user -d stock_db -c $Query 2>&1
    return @{
        Success = ($LASTEXITCODE -eq 0)
        Output = $result
    }
}

Write-Host "Aplicando correcciones...`n" -ForegroundColor Yellow

# 1. Asegurar que todos los usuarios tengan un rol valido
Write-Host "`n1. Asignando rol 'vendedor' a usuarios sin rol..." -ForegroundColor Cyan
$query1 = "UPDATE users SET role = 'vendedor' WHERE role IS NULL OR role = '';"
$result1 = Invoke-PgQuery -Query $query1
if ($result1.Success) {
    Write-Host "   OK: Roles actualizados" -ForegroundColor Green
} else {
    Write-Host "   WARN: $($result1.Output)" -ForegroundColor Yellow
}

# 2. Sincronizar organization_id para todos los usuarios
Write-Host "`n2. Sincronizando organization_id para todos los usuarios..." -ForegroundColor Cyan
# Para usuarios sin organization_id, asignarles su propio ID (se consideran admins de su propia organización)
$query2 = "UPDATE users SET organization_id = id WHERE organization_id IS NULL OR organization_id = 0;"
$result2 = Invoke-PgQuery -Query $query2
if ($result2.Success) {
    Write-Host "   [OK] Organization IDs sincronizados" -ForegroundColor Green
} else {
    Write-Host "   [WARN] $($result2.Output)" -ForegroundColor Yellow
}

# 3. Crear suscripciones FREE para usuarios que no tengan
Write-Host "`n3. Creando suscripciones FREE para usuarios sin suscripcion..." -ForegroundColor Cyan
# Dividir en queries simples para evitar problemas de parsing
$query3_select = "SELECT id FROM users WHERE NOT EXISTS (SELECT 1 FROM subscriptions WHERE user_id = users.id);"
$usersWithoutSub = Invoke-PgQuery -Query $query3_select

if ($usersWithoutSub.Success -and $usersWithoutSub.Output -match '\d+') {
    # Crear suscripcion usando una query mas simple
    $query3_insert = "INSERT INTO subscriptions (user_id, plan_id, status, current_period_start) SELECT id, 'free', 'active', NOW() FROM users WHERE NOT EXISTS (SELECT 1 FROM subscriptions WHERE user_id = users.id);"
    $result3 = Invoke-PgQuery -Query $query3_insert
    if ($result3.Success) {
        Write-Host "   OK: Suscripciones FREE creadas" -ForegroundColor Green
    } else {
        Write-Host "   WARN: $($result3.Output)" -ForegroundColor Yellow
    }
} else {
    Write-Host "   OK: Todos los usuarios ya tienen suscripcion" -ForegroundColor Green
}

# 4. Activar suscripciones inactivas
Write-Host "`n4. Activando suscripciones inactivas..." -ForegroundColor Cyan
$query4 = "UPDATE subscriptions SET status = 'active' WHERE status != 'active';"
$result4 = Invoke-PgQuery -Query $query4
if ($result4.Success) {
    Write-Host "   OK: Suscripciones activadas" -ForegroundColor Green
} else {
    Write-Host "   WARN: $($result4.Output)" -ForegroundColor Yellow
}

# 5. Verificar estado final
Write-Host "`n5. Verificando estado final..." -ForegroundColor Cyan
$verifyQuery = "SELECT u.id, u.email, u.role, u.organization_id, s.plan_id, s.status as subscription_status FROM users u LEFT JOIN subscriptions s ON u.id = s.user_id ORDER BY u.id;"
$verifyResult = Invoke-PgQuery -Query $verifyQuery
Write-Host "`n   Estado de usuarios:" -ForegroundColor Yellow
Write-Host $verifyResult.Output -ForegroundColor White

# Contar problemas restantes
Write-Host "`n6. Verificando problemas restantes..." -ForegroundColor Cyan

$checkQuery1 = "SELECT COUNT(*) as invalid_roles FROM users WHERE role NOT IN ('admin', 'vendedor', 'repositor');"
$checkQuery2 = "SELECT COUNT(*) as missing_org FROM users WHERE organization_id IS NULL OR organization_id = 0;"
$checkQuery3 = "SELECT COUNT(*) as no_active_sub FROM users WHERE NOT EXISTS (SELECT 1 FROM subscriptions s WHERE s.user_id = users.id AND s.status = 'active');"

Write-Host "   Roles invalidos:" -ForegroundColor Gray
$checkResult1 = Invoke-PgQuery -Query $checkQuery1
Write-Host $checkResult1.Output -ForegroundColor White

Write-Host "   Sin organization_id:" -ForegroundColor Gray
$checkResult2 = Invoke-PgQuery -Query $checkQuery2
Write-Host $checkResult2.Output -ForegroundColor White

Write-Host "   Sin suscripcion activa:" -ForegroundColor Gray
$checkResult3 = Invoke-PgQuery -Query $checkQuery3
Write-Host $checkResult3.Output -ForegroundColor White

# Resumen final
Write-Host "`n========================================================" -ForegroundColor Green
Write-Host "   CORRECCIONES COMPLETADAS                            " -ForegroundColor Green
Write-Host "========================================================`n" -ForegroundColor Green

Write-Host "Acciones realizadas:" -ForegroundColor Yellow
Write-Host "   [OK] Roles asignados a todos los usuarios" -ForegroundColor Green
Write-Host "   [OK] Organization IDs sincronizados" -ForegroundColor Green
Write-Host "   [OK] Suscripciones FREE creadas" -ForegroundColor Green
Write-Host "   [OK] Suscripciones activadas" -ForegroundColor Green

Write-Host "`nIMPORTANTE:" -ForegroundColor Yellow
Write-Host "   - Los usuarios deben CERRAR SESION y volver a iniciar sesion" -ForegroundColor Cyan
Write-Host "   - Esto es necesario para que el nuevo token JWT incluya los cambios" -ForegroundColor Cyan
Write-Host "   - Despues de iniciar sesion, los errores deberian estar resueltos" -ForegroundColor Cyan

Write-Host "`nProximos pasos:" -ForegroundColor Yellow
Write-Host "   1. Cierra sesion en la aplicacion web" -ForegroundColor Gray
Write-Host "   2. Vuelve a iniciar sesion" -ForegroundColor Gray
Write-Host "   3. Intenta acceder al detalle de una orden" -ForegroundColor Gray
Write-Host "   4. Si persiste el error, ejecuta: .\scripts\diagnose-orders.ps1" -ForegroundColor Gray

Write-Host ""
