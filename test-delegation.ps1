# Script de prueba de delegacion
Write-Host ""
Write-Host "Probando el flujo completo de delegacion..." -ForegroundColor Cyan
Write-Host ""

# 1. Login para obtener token
Write-Host "1. Login como usuario de prueba..." -ForegroundColor Yellow

$loginBody = @{
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users/login" -Method POST -Body $loginBody -ContentType "application/json"
    
    $token = $loginResponse.token
    Write-Host "Token obtenido OK" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: No se pudo hacer login" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}

# 2. Solicitar reporte de productos por email
Write-Host "2. Solicitando reporte de productos por email..." -ForegroundColor Yellow

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

try {
    $reportResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/reports/products/email" -Method POST -Headers $headers
    
    Write-Host "HTTP Status: 202 Accepted" -ForegroundColor Green
    Write-Host "Response: $($reportResponse.message)" -ForegroundColor White
    Write-Host ""
    Write-Host "Solicitud aceptada correctamente" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error en la solicitud" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}

# 3. Verificar que el worker procesó el mensaje
Write-Host "3. Verificando logs del worker..." -ForegroundColor Yellow
Write-Host ""
Start-Sleep -Seconds 2
docker logs stock_in_order_worker --tail 15

Write-Host ""
Write-Host "Prueba completada!" -ForegroundColor Green
Write-Host "El worker deberia haber generado el reporte" -ForegroundColor Cyan
Write-Host ""
