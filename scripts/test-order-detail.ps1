# Script para probar endpoints de detalle de órdenes
# Este script ayuda a depurar los problemas con las vistas de detalle

param(
    [Parameter(Mandatory=$false)]
    [string]$BaseUrl = "http://localhost:8080/api/v1",
    
    [Parameter(Mandatory=$true)]
    [string]$Token,
    
    [Parameter(Mandatory=$false)]
    [int]$SalesOrderId = 1,
    
    [Parameter(Mandatory=$false)]
    [int]$PurchaseOrderId = 1
)

Write-Host "`n=== Test de Endpoints de Detalle de Órdenes ===" -ForegroundColor Cyan
Write-Host "URL Base: $BaseUrl" -ForegroundColor Gray
Write-Host "Token (primeros 20 chars): $($Token.Substring(0, [Math]::Min(20, $Token.Length)))..." -ForegroundColor Gray
Write-Host ""

$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
    "Accept" = "application/json"
}

# Función helper para hacer requests
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Url,
        [string]$Description
    )
    
    Write-Host "`n--- $Description ---" -ForegroundColor Yellow
    Write-Host "  Método: $Method" -ForegroundColor Gray
    Write-Host "  URL: $Url" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $headers -ErrorAction Stop
        Write-Host "  ✅ Éxito" -ForegroundColor Green
        Write-Host "  Respuesta:" -ForegroundColor Gray
        $response | ConvertTo-Json -Depth 10 | Write-Host
        return $response
    }
    catch {
        Write-Host "  ❌ Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "  Status Code: $statusCode" -ForegroundColor Red
            
            try {
                $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
                $responseBody = $reader.ReadToEnd()
                Write-Host "  Cuerpo de respuesta: $responseBody" -ForegroundColor Red
            }
            catch {
                Write-Host "  No se pudo leer el cuerpo de la respuesta" -ForegroundColor Red
            }
        }
        return $null
    }
}

# 1. Probar endpoint de sales orders (lista)
Write-Host "`n========== SALES ORDERS ==========" -ForegroundColor Magenta
Test-Endpoint -Method "GET" -Url "$BaseUrl/sales-orders" -Description "Listar Órdenes de Venta"

# 2. Probar endpoint de sales order detail
Test-Endpoint -Method "GET" -Url "$BaseUrl/sales-orders/$SalesOrderId" -Description "Detalle de Orden de Venta #$SalesOrderId"

# 3. Probar endpoint de purchase orders (lista)
Write-Host "`n========== PURCHASE ORDERS ==========" -ForegroundColor Magenta
Test-Endpoint -Method "GET" -Url "$BaseUrl/purchase-orders" -Description "Listar Órdenes de Compra"

# 4. Probar endpoint de purchase order detail
Test-Endpoint -Method "GET" -Url "$BaseUrl/purchase-orders/$PurchaseOrderId" -Description "Detalle de Orden de Compra #$PurchaseOrderId"

# 5. Verificar token (decodificar JWT)
Write-Host "`n========== ANÁLISIS DEL TOKEN JWT ==========" -ForegroundColor Magenta
Write-Host "Intentando decodificar JWT..." -ForegroundColor Gray

try {
    $parts = $Token.Split('.')
    if ($parts.Length -eq 3) {
        # Decodificar el payload (segunda parte)
        $payload = $parts[1]
        # Agregar padding si es necesario
        $padding = 4 - ($payload.Length % 4)
        if ($padding -ne 4) {
            $payload = $payload + ("=" * $padding)
        }
        
        $payloadBytes = [System.Convert]::FromBase64String($payload)
        $payloadJson = [System.Text.Encoding]::UTF8.GetString($payloadBytes)
        $claims = $payloadJson | ConvertFrom-Json
        
        Write-Host "`nClaims del JWT:" -ForegroundColor Green
        $claims | ConvertTo-Json -Depth 10 | Write-Host
        
        Write-Host "`nVerificación de permisos:" -ForegroundColor Yellow
        Write-Host "  User ID: $($claims.user_id)" -ForegroundColor Gray
        Write-Host "  Email: $($claims.email)" -ForegroundColor Gray
        Write-Host "  Role: $($claims.role)" -ForegroundColor $(if ($claims.role -in @('admin', 'vendedor')) { 'Green' } else { 'Red' })
        Write-Host "  Organization ID: $($claims.organization_id)" -ForegroundColor Gray
        
        if ($claims.role -notin @('admin', 'vendedor')) {
            Write-Host "`n  ⚠️  WARNING: El usuario NO tiene rol de 'admin' o 'vendedor'" -ForegroundColor Red
            Write-Host "     Los endpoints de órdenes requieren estos roles." -ForegroundColor Red
        }
        else {
            Write-Host "`n  ✅ El usuario tiene el rol correcto: $($claims.role)" -ForegroundColor Green
        }
    }
    else {
        Write-Host "  ❌ Token JWT inválido (no tiene 3 partes)" -ForegroundColor Red
    }
}
catch {
    Write-Host "  ❌ Error decodificando JWT: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== Fin del test ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "USO:" -ForegroundColor Yellow
Write-Host "  .\test-order-detail.ps1 -Token 'tu-jwt-token-aqui' -SalesOrderId 1 -PurchaseOrderId 1" -ForegroundColor Gray
Write-Host ""
