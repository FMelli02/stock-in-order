# Script para ejecutar tests de FEFO
# Asegúrate de tener PostgreSQL corriendo y la base de datos de test configurada

Write-Host "🧪 Ejecutando Tests de Lógica FEFO..." -ForegroundColor Cyan
Write-Host ""

# Configurar variable de entorno para la base de datos de test
$env:TEST_DATABASE_URL = "postgres://postgres:postgres@localhost:5432/stock_in_order_test?sslmode=disable"

Write-Host "📦 Descargando dependencias..." -ForegroundColor Yellow
go mod download

Write-Host ""
Write-Host "🔬 Ejecutando tests..." -ForegroundColor Green
Write-Host ""

# Ejecutar tests con verbose output
go test -v ./internal/models -run "TestConsumeStockFEFO|TestSalesOrderModel_Create"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ ¡Todos los tests pasaron!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Tests ejecutados:" -ForegroundColor Cyan
    Write-Host "  ✓ TestConsumeStockFEFO - Lógica FEFO básica (15 unidades)" -ForegroundColor White
    Write-Host "  ✓ TestConsumeStockFEFO_InsufficientStock - Rollback con stock insuficiente" -ForegroundColor White
    Write-Host "  ✓ TestConsumeStockFEFO_ExactAmount - Consumo exacto (20 unidades)" -ForegroundColor White
    Write-Host "  ✓ TestConsumeStockFEFO_SingleBatch - Consumo de un solo lote" -ForegroundColor White
    Write-Host "  ✓ TestSalesOrderModel_Create_IntegrationFEFO - Integración completa" -ForegroundColor White
    Write-Host "  ✓ TestSalesOrderModel_Create_InsufficientStock - Validación de errores" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "❌ Algunos tests fallaron. Revisa el output arriba." -ForegroundColor Red
    exit 1
}
