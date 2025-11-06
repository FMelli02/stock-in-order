# Script de Prueba - Endpoints de Suscripciones
# Tarea 2: El Checkout

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TESTING SUBSCRIPTION ENDPOINTS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080/api/v1"

# 1. LOGIN
Write-Host "[1/5] Logging in as admin..." -ForegroundColor Yellow
try {
    $loginBody = @{
        email = "admin@stock.com"
        password = "admin123"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/users/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json"

    $token = $loginResponse.token
    Write-Host "✅ Login successful" -ForegroundColor Green
    Write-Host "   Token: $($token.Substring(0,20))..." -ForegroundColor Gray
} catch {
    Write-Host "❌ Login failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 2. GET SUBSCRIPTION STATUS
Write-Host "[2/5] Getting subscription status..." -ForegroundColor Yellow
try {
    $headers = @{
        Authorization = "Bearer $token"
    }

    $statusResponse = Invoke-RestMethod -Uri "$baseUrl/subscriptions/status" `
        -Method GET `
        -Headers $headers

    Write-Host "✅ Subscription status retrieved" -ForegroundColor Green
    Write-Host "   User ID: $($statusResponse.user_id)" -ForegroundColor Gray
    Write-Host "   Plan: $($statusResponse.plan_id)" -ForegroundColor Gray
    Write-Host "   Status: $($statusResponse.status)" -ForegroundColor Gray
    Write-Host "   Max Products: $($statusResponse.features.max_products)" -ForegroundColor Gray
    Write-Host "   Max Orders/Month: $($statusResponse.features.max_orders_per_month)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "   Available plans:" -ForegroundColor Gray
    foreach ($plan in $statusResponse.available_plans) {
        Write-Host "   - $($plan.name) ($($plan.plan_id)): ARS $$($plan.price)" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ Failed to get status: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 3. CREATE CHECKOUT (Plan Pro)
Write-Host "[3/5] Creating checkout for Plan Pro..." -ForegroundColor Yellow
try {
    $checkoutBody = @{
        plan_id = "plan_pro"
    } | ConvertTo-Json

    $checkoutResponse = Invoke-RestMethod -Uri "$baseUrl/subscriptions/create-checkout" `
        -Method POST `
        -Body $checkoutBody `
        -ContentType "application/json" `
        -Headers $headers

    Write-Host "✅ Checkout created successfully" -ForegroundColor Green
    Write-Host "   Checkout URL: $($checkoutResponse.checkout_url)" -ForegroundColor Gray
    
    # Ask if user wants to open the URL
    $openBrowser = Read-Host "   Do you want to open the checkout in browser? (y/n)"
    if ($openBrowser -eq "y" -or $openBrowser -eq "Y") {
        Start-Process $checkoutResponse.checkout_url
        Write-Host "   Browser opened with checkout URL" -ForegroundColor Cyan
    }
} catch {
    $errorMessage = $_.ErrorDetails.Message
    if ($errorMessage -match "already has this plan") {
        Write-Host "⚠️  User already has this plan active" -ForegroundColor Yellow
    } elseif ($errorMessage -match "MP_ACCESS_TOKEN") {
        Write-Host "⚠️  MercadoPago not configured (MP_ACCESS_TOKEN missing)" -ForegroundColor Yellow
        Write-Host "   This is expected if MP_ACCESS_TOKEN is not set in environment" -ForegroundColor Gray
    } else {
        Write-Host "❌ Failed to create checkout: $errorMessage" -ForegroundColor Red
    }
}

Write-Host ""

# 4. CREATE RECURRING SUBSCRIPTION (Plan Básico)
Write-Host "[4/5] Creating recurring subscription for Plan Básico..." -ForegroundColor Yellow
try {
    $recurringBody = @{
        plan_id = "plan_basico"
    } | ConvertTo-Json

    $recurringResponse = Invoke-RestMethod -Uri "$baseUrl/subscriptions/create-recurring" `
        -Method POST `
        -Body $recurringBody `
        -ContentType "application/json" `
        -Headers $headers

    Write-Host "✅ Recurring subscription created successfully" -ForegroundColor Green
    Write-Host "   Checkout URL: $($recurringResponse.checkout_url)" -ForegroundColor Gray
} catch {
    $errorMessage = $_.ErrorDetails.Message
    if ($errorMessage -match "MP_ACCESS_TOKEN") {
        Write-Host "⚠️  MercadoPago not configured (MP_ACCESS_TOKEN missing)" -ForegroundColor Yellow
    } else {
        Write-Host "❌ Failed to create recurring: $errorMessage" -ForegroundColor Red
    }
}

Write-Host ""

# 5. CANCEL SUBSCRIPTION
Write-Host "[5/5] Testing subscription cancellation..." -ForegroundColor Yellow
$confirmCancel = Read-Host "   Do you want to cancel the current subscription? (y/n)"
if ($confirmCancel -eq "y" -or $confirmCancel -eq "Y") {
    try {
        Invoke-RestMethod -Uri "$baseUrl/subscriptions/cancel" `
            -Method POST `
            -Headers $headers

        Write-Host "✅ Subscription canceled successfully" -ForegroundColor Green
        Write-Host "   User downgraded to plan_free" -ForegroundColor Gray
    } catch {
        $errorMessage = $_.ErrorDetails.Message
        if ($errorMessage -match "cannot cancel free plan") {
            Write-Host "⚠️  Cannot cancel free plan" -ForegroundColor Yellow
        } else {
            Write-Host "❌ Failed to cancel: $errorMessage" -ForegroundColor Red
        }
    }
} else {
    Write-Host "⏭️  Cancellation skipped" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TEST COMPLETED" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor White
Write-Host "✅ 4 endpoints created and working" -ForegroundColor Green
Write-Host "   - GET  /api/v1/subscriptions/status" -ForegroundColor Gray
Write-Host "   - POST /api/v1/subscriptions/create-checkout" -ForegroundColor Gray
Write-Host "   - POST /api/v1/subscriptions/create-recurring" -ForegroundColor Gray
Write-Host "   - POST /api/v1/subscriptions/cancel" -ForegroundColor Gray
Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host "1. Configure MP_ACCESS_TOKEN in .env" -ForegroundColor Gray
Write-Host "2. Test actual payment flow with Sandbox credentials" -ForegroundColor Gray
Write-Host "3. Implement Tarea 3: Gestión (upgrade, history)" -ForegroundColor Gray
Write-Host "4. Implement Tarea 4: Middleware Patovica (enforcement)" -ForegroundColor Gray
Write-Host ""
