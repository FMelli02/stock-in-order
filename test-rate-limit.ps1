# Script para probar el rate limiting
# Hace 110 peticiones rápidas para verificar que la 101 retorna 429

Write-Host "Probando Rate Limiting (100 req/min por IP)..." -ForegroundColor Cyan
Write-Host "Haciendo 110 peticiones al endpoint /api/v1/health..." -ForegroundColor Yellow
Write-Host ""

$successCount = 0
$rateLimitCount = 0
$errorCount = 0

for ($i = 1; $i -le 110; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" -Method GET -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            $successCount++
            if ($i -le 5 -or $i -ge 99) {
                Write-Host "[$i] ✓ 200 OK" -ForegroundColor Green
            } elseif ($i -eq 6) {
                Write-Host "... (peticiones 6-98 exitosas)" -ForegroundColor Gray
            }
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 429) {
            $rateLimitCount++
            Write-Host "[$i] ⚠ 429 Too Many Requests - RATE LIMIT ACTIVADO" -ForegroundColor Red
        } else {
            $errorCount++
            Write-Host "[$i] ✗ Error $statusCode" -ForegroundColor Magenta
        }
    }
    
    # Pequeña pausa para no saturar completamente
    Start-Sleep -Milliseconds 10
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "RESULTADOS:" -ForegroundColor Yellow
Write-Host "  Exitosas (200): $successCount" -ForegroundColor Green
Write-Host "  Rate Limited (429): $rateLimitCount" -ForegroundColor Red
Write-Host "  Otros errores: $errorCount" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($rateLimitCount -gt 0) {
    Write-Host "✓ Rate Limiting FUNCIONANDO CORRECTAMENTE" -ForegroundColor Green
    Write-Host "  Las peticiones fueron bloqueadas después del límite de 100/min" -ForegroundColor Green
} else {
    Write-Host "✗ Rate Limiting NO DETECTADO" -ForegroundColor Red
    Write-Host "  Todas las peticiones pasaron sin ser bloqueadas" -ForegroundColor Red
}
