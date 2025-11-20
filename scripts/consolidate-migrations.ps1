# Script para consolidar todas las migraciones en un solo archivo para Supabase
# Este script une todos los archivos .up.sql en orden para crear el schema completo

$migrationsPath = Join-Path $PSScriptRoot "..\backend\migrations"
$outputFile = Join-Path $PSScriptRoot "..\supabase-schema.sql"

Write-Host "🔄 Consolidando migraciones..." -ForegroundColor Cyan

# Limpiar archivo de salida si existe
if (Test-Path $outputFile) {
    Remove-Item $outputFile
}

# Header del archivo
@"
-- ============================================
-- Stock In Order - Schema Completo para Supabase
-- Generado: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
-- ============================================

-- Habilitar extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

"@ | Out-File -FilePath $outputFile -Encoding UTF8

# Obtener todos los archivos .up.sql ordenados
$migrationFiles = Get-ChildItem -Path $migrationsPath -Filter "*.up.sql" | 
    Where-Object { $_.Name -notmatch "\.down\.sql$" } |
    Sort-Object Name

$count = 0
foreach ($file in $migrationFiles) {
    $count++
    Write-Host "  [$count/$($migrationFiles.Count)] Procesando: $($file.Name)" -ForegroundColor Gray
    
    # Agregar separador
    @"

-- ============================================
-- Migración: $($file.Name)
-- ============================================

"@ | Out-File -FilePath $outputFile -Append -Encoding UTF8
    
    # Leer y agregar contenido del archivo
    $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    $content | Out-File -FilePath $outputFile -Append -Encoding UTF8 -NoNewline
    
    # Agregar salto de línea al final
    "`n" | Out-File -FilePath $outputFile -Append -Encoding UTF8 -NoNewline
}

Write-Host "✅ Schema consolidado creado: $outputFile" -ForegroundColor Green
Write-Host "📊 Total de migraciones procesadas: $count" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Siguiente paso:" -ForegroundColor Yellow
Write-Host "   1. Ve a tu proyecto de Supabase (https://supabase.com/dashboard)" -ForegroundColor White
Write-Host "   2. Abre el SQL Editor" -ForegroundColor White
Write-Host "   3. Copia y pega el contenido de 'supabase-schema.sql'" -ForegroundColor White
Write-Host "   4. Ejecuta el script" -ForegroundColor White
