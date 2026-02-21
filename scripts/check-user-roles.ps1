# Script para verificar y actualizar roles de usuarios
# Este script ayuda a diagnosticar problemas de permisos

Write-Host "`n=== Verificación y Actualización de Roles de Usuarios ===" -ForegroundColor Cyan

# Verificar si Docker está corriendo
try {
    $dockerRunning = docker ps 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Docker no está corriendo. Inicia Docker Desktop primero." -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "❌ Docker no está instalado o no está en el PATH" -ForegroundColor Red
    exit 1
}

# Nombre del contenedor de PostgreSQL
$containerName = "stock-in-order-postgres-1"

Write-Host "`n1. Verificando contenedor de PostgreSQL..." -ForegroundColor Yellow
$container = docker ps --filter "name=postgres" --format "{{.Names}}" | Select-Object -First 1

if (-not $container) {
    Write-Host "   ❌ No se encontró contenedor de PostgreSQL corriendo" -ForegroundColor Red
    Write-Host "   Intentando con nombre alternativo..." -ForegroundColor Gray
    $container = docker ps --format "{{.Names}}" | Where-Object { $_ -match "postgres" } | Select-Object -First 1
}

if ($container) {
    Write-Host "   ✅ Contenedor encontrado: $container" -ForegroundColor Green
    $containerName = $container
}
else {
    Write-Host "   ❌ No se encontró ningún contenedor de PostgreSQL" -ForegroundColor Red
    Write-Host "   Ejecuta: docker-compose up -d" -ForegroundColor Gray
    exit 1
}

# Listar todos los usuarios y sus roles
Write-Host "`n2. Listando usuarios y sus roles..." -ForegroundColor Yellow
$query = "SELECT id, name, email, role, organization_id FROM users ORDER BY id;"
$result = docker exec -i $containerName psql -U postgres -d stockinorder -c $query 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host $result -ForegroundColor White
}
else {
    Write-Host "   ❌ Error ejecutando query: $result" -ForegroundColor Red
    exit 1
}

# Preguntar si desea actualizar algún rol
Write-Host "`n3. ¿Deseas actualizar el rol de algún usuario?" -ForegroundColor Yellow
Write-Host "   Roles disponibles: admin, vendedor, repositor" -ForegroundColor Gray
$update = Read-Host "   ¿Actualizar roles? (s/n)"

if ($update -eq "s" -or $update -eq "S") {
    $email = Read-Host "   Ingresa el email del usuario"
    Write-Host "   Roles disponibles:" -ForegroundColor Gray
    Write-Host "     - admin      (acceso completo)" -ForegroundColor Gray
    Write-Host "     - vendedor   (ventas y reportes)" -ForegroundColor Gray
    Write-Host "     - repositor  (solo inventario)" -ForegroundColor Gray
    $role = Read-Host "   Ingresa el nuevo rol"
    
    if ($role -in @("admin", "vendedor", "repositor")) {
        Write-Host "`n   Actualizando rol..." -ForegroundColor Yellow
        $updateQuery = "UPDATE users SET role = '$role' WHERE email = '$email';"
        $updateResult = docker exec -i $containerName psql -U postgres -d stockinorder -c $updateQuery 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "   ✅ Rol actualizado exitosamente" -ForegroundColor Green
            
            # Mostrar resultado
            Write-Host "`n   Usuario actualizado:" -ForegroundColor Green
            $verifyQuery = "SELECT id, name, email, role, organization_id FROM users WHERE email = '$email';"
            $verifyResult = docker exec -i $containerName psql -U postgres -d stockinorder -c $verifyQuery 2>&1
            Write-Host $verifyResult -ForegroundColor White
        }
        else {
            Write-Host "   ❌ Error actualizando rol: $updateResult" -ForegroundColor Red
        }
    }
    else {
        Write-Host "   ❌ Rol inválido. Usa: admin, vendedor o repositor" -ForegroundColor Red
    }
}

# Verificar permisos de endpoints
Write-Host "`n4. Permisos de endpoints por rol:" -ForegroundColor Yellow
Write-Host "   📦 Productos:         Todos los roles (con paywall)" -ForegroundColor Gray
Write-Host "   👥 Clientes:          Admin únicamente" -ForegroundColor Gray
Write-Host "   🏭 Proveedores:       Admin únicamente" -ForegroundColor Gray
Write-Host "   🛒 Órdenes de Venta:  Admin + Vendedor" -ForegroundColor Cyan
Write-Host "   📦 Órdenes de Compra: Todos los roles" -ForegroundColor Gray
Write-Host "   📊 Reportes:          Admin + Vendedor" -ForegroundColor Gray
Write-Host "   🔗 Integraciones:     Admin únicamente" -ForegroundColor Gray

Write-Host "`n=== Fin de la verificación ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "💡 NOTA: Si cambias el rol de un usuario, debe cerrar sesión y volver a" -ForegroundColor Yellow
Write-Host "   iniciar sesión para que el nuevo rol se refleje en el token JWT." -ForegroundColor Yellow
Write-Host ""
