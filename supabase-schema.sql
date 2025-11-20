-- ============================================
-- Stock In Order - Schema Completo para Supabase
-- Generado: 2025-11-15 12:43:40
-- ============================================

-- Habilitar extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ============================================
-- MigraciÃ³n: 000001_create_users_table.up.sql
-- ============================================

-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================
-- MigraciÃ³n: 000002_create_products_table.up.sql
-- ============================================

-- +migrate Up
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    sku TEXT NOT NULL,
    description TEXT,
    quantity INTEGER NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, sku)
);


-- ============================================
-- MigraciÃ³n: 000003_create_suppliers_table.up.sql
-- ============================================

-- +migrate Up
CREATE TABLE IF NOT EXISTS suppliers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    contact_person TEXT,
    email TEXT,
    phone TEXT,
    address TEXT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================
-- MigraciÃ³n: 000004_create_customers_table.up.sql
-- ============================================

-- +migrate Up
CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    address TEXT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================
-- MigraciÃ³n: 000005_create_sales_orders_tables.up.sql
-- ============================================

-- +migrate Up
CREATE TABLE IF NOT EXISTS sales_orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES customers(id),
    order_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'pending',
    total_amount NUMERIC(10,2),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL
);


-- ============================================
-- MigraciÃ³n: 000006_create_purchase_orders_tables.up.sql
-- ============================================

-- Migration: Create purchase orders tables
-- Creates purchase_orders and purchase_order_items

BEGIN;

CREATE TABLE IF NOT EXISTS purchase_orders (
    id BIGSERIAL PRIMARY KEY,
    supplier_id BIGINT REFERENCES suppliers(id),
    order_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'pending',
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id BIGSERIAL PRIMARY KEY,
    purchase_order_id BIGINT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_cost NUMERIC(10, 2) NOT NULL
);

COMMIT;


-- ============================================
-- MigraciÃ³n: 000007_create_stock_movements_table.up.sql
-- ============================================

-- Migration: Create stock_movements table
-- Purpose: Ledger of inventory changes per product

BEGIN;

CREATE TABLE IF NOT EXISTS stock_movements (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity_change INTEGER NOT NULL,
    reason TEXT NOT NULL,
    reference_id TEXT,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;


-- ============================================
-- MigraciÃ³n: 000008_seed_initial_data.up.sql
-- ============================================

-- NOTA: Seed data removido para evitar conflictos con migraciones posteriores
-- Los datos de prueba deben ser insertados manualmente después de aplicar todas las migraciones

-- ============================================
-- Fin del seeding
-- ============================================


-- ============================================
-- MigraciÃ³n: 000009_add_user_roles.up.sql
-- ============================================

-- Crear tipo ENUM para los roles de usuario
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('admin', 'vendedor', 'repositor');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Añadir columna role a la tabla users
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS role user_role NOT NULL DEFAULT 'vendedor';

-- Actualizar el usuario admin existente (si existe) para que tenga rol admin
UPDATE users 
SET role = 'admin' 
WHERE email = 'admin@stockinorder.com';


-- ============================================
-- Migraciὃn: 000010_add_stock_alerts_to_products.up.sql
-- ============================================

-- Add stock alert columns to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS stock_minimo INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS notificado BOOLEAN NOT NULL DEFAULT false;


-- ============================================
-- MigraciÃ³n: 000011_create_integrations_table.up.sql
-- ============================================

CREATE TABLE IF NOT EXISTS integrations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform TEXT NOT NULL, -- Ej: 'mercadolibre', 'shopify'
    external_user_id TEXT, -- El ID del usuario en la plataforma externa
    access_token BYTEA NOT NULL, -- El token, encriptado
    refresh_token BYTEA, -- El token de refresco, encriptado (puede ser NULL para algunas plataformas)
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, platform) -- Un usuario solo puede tener una conexión por plataforma
);

CREATE INDEX IF NOT EXISTS idx_integrations_user_id ON integrations(user_id);
CREATE INDEX IF NOT EXISTS idx_integrations_platform ON integrations(platform);


-- ============================================
-- MigraciÃ³n: 000012_create_audit_logs_table.up.sql
-- ============================================

-- Tabla para auditoría (libro de actas)
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    user_email TEXT NOT NULL,
    user_role TEXT NOT NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER,
    details TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices para mejorar el rendimiento de consultas
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);

-- ============================================
-- MigraciÃ³n: 000013_add_batch_tracking.up.sql
-- ============================================

-- Migration: Add Batch Tracking System
-- Descripción: Refactoriza el sistema de stock de productos
-- - Crea tabla product_batches para tracking de lotes
-- - Migra stock existente a lotes iniciales
-- - Elimina columna quantity de products

BEGIN;

-- 1. Crear nueva tabla product_batches
CREATE TABLE IF NOT EXISTS product_batches (
    id BIGSERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lote_number TEXT,
    quantity INTEGER NOT NULL CHECK (quantity >= 0),
    expiry_date DATE, -- Fecha de vencimiento
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Crear índices para optimizar consultas
CREATE INDEX IF NOT EXISTS idx_product_batches_product_id ON product_batches(product_id);
CREATE INDEX IF NOT EXISTS idx_product_batches_expiry_date ON product_batches(expiry_date);
CREATE INDEX IF NOT EXISTS idx_product_batches_user_id ON product_batches(user_id);

-- 3. Migrar datos existentes de products a product_batches
-- NOTA: Este paso se omite en instalaciones nuevas porque la columna quantity
-- nunca existió en la tabla products (fue eliminada en esta misma migración)
-- Solo aplicable cuando se migra desde un sistema existente

-- 4. La Cirugía: Eliminar columna quantity de products
ALTER TABLE products DROP COLUMN IF EXISTS quantity;

COMMIT;


-- ============================================
-- Migraciὃn: 000014_add_batch_fields_to_purchase_items.up.sql
-- ============================================

-- Migration: Add batch tracking fields to purchase_order_items
-- Adds lote_number and expiry_date to support batch creation on order completion

BEGIN;

ALTER TABLE purchase_order_items
ADD COLUMN IF NOT EXISTS lote_number TEXT,
ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMPTZ;

COMMIT;


-- ============================================
-- MigraciÃ³n: 000015_create_subscriptions_table.up.sql
-- ============================================

-- NOTA: Esta migración fue reemplazada por 000016 con un esquema más completo
-- Todo el contenido se movió a la migración 000016


-- ============================================
-- MigraciÃ³n: 000016_create_subscriptions_tables.up.sql
-- ============================================

-- Migración 000015: Sistema de Suscripciones y Pagos
-- Descripción: Crea las tablas necesarias para gestionar planes, suscripciones y pagos con MercadoPago
-- Fecha: 6 de Noviembre, 2025

-- ============================================
-- 1. ENUM para estados de suscripción
-- ============================================
DO $$ BEGIN
    CREATE TYPE subscription_status AS ENUM (
        'active',      -- Suscripción activa y pagada
        'inactive',    -- Suscripción inactiva (recién creada, sin pago)
        'past_due',    -- Pago vencido, en período de gracia
        'canceled'     -- Cancelada por el usuario o por falta de pago
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ============================================
-- 2. Tabla de Suscripciones
-- ============================================
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    
    -- Relación con usuario (1 usuario = 1 suscripción)
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Plan contratado
    plan_id TEXT NOT NULL, -- 'plan_free', 'plan_basico', 'plan_pro', 'plan_enterprise'
    
    -- IDs de MercadoPago (para sincronización)
    mp_subscription_id TEXT, -- ID de la suscripción en MercadoPago
    mp_customer_id TEXT,     -- ID del cliente en MercadoPago
    mp_preapproval_id TEXT,  -- ID de pre-aprobación (para suscripciones recurrentes)
    
    -- Estado y fechas
    status subscription_status NOT NULL DEFAULT 'inactive',
    current_period_start TIMESTAMPTZ, -- Inicio del período actual
    current_period_end TIMESTAMPTZ,   -- Fin del período actual (cuándo vence)
    trial_end TIMESTAMPTZ,            -- Fin del período de prueba (si aplica)
    canceled_at TIMESTAMPTZ,          -- Fecha de cancelación
    
    -- Metadatos
    cancel_reason TEXT,               -- Razón de cancelación
    metadata JSONB,                   -- Datos adicionales (configuración, preferencias, etc.)
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================
-- 3. Tabla de Historial de Pagos
-- ============================================
CREATE TABLE IF NOT EXISTS payment_history (
    id BIGSERIAL PRIMARY KEY,
    
    -- Relación con suscripción
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Información de MercadoPago
    mp_payment_id TEXT UNIQUE,        -- ID del pago en MercadoPago
    mp_status TEXT NOT NULL,          -- Estado según MP: approved, pending, rejected, etc.
    mp_status_detail TEXT,            -- Detalle del estado
    
    -- Montos
    amount DECIMAL(10, 2) NOT NULL,   -- Monto pagado
    currency TEXT NOT NULL DEFAULT 'ARS', -- Moneda (ARS, USD, etc.)
    
    -- Descripción
    description TEXT,                 -- Descripción del pago
    plan_id TEXT NOT NULL,            -- Plan pagado
    
    -- Fechas
    payment_date TIMESTAMPTZ,         -- Fecha del pago exitoso
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================
-- 4. Índices para optimización
-- ============================================

-- Búsqueda rápida por usuario
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions(plan_id);

-- Búsqueda de suscripciones próximas a vencer (para notificaciones)
CREATE INDEX IF NOT EXISTS idx_subscriptions_period_end ON subscriptions(current_period_end);

-- IDs de MercadoPago para webhooks
CREATE INDEX IF NOT EXISTS idx_subscriptions_mp_subscription_id ON subscriptions(mp_subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_mp_customer_id ON subscriptions(mp_customer_id);

-- Historial de pagos
CREATE INDEX IF NOT EXISTS idx_payment_history_subscription_id ON payment_history(subscription_id);
CREATE INDEX IF NOT EXISTS idx_payment_history_user_id ON payment_history(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_history_mp_payment_id ON payment_history(mp_payment_id);

-- ============================================
-- 5. Función para actualizar updated_at automáticamente
-- ============================================
CREATE OR REPLACE FUNCTION update_subscriptions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para updated_at
DROP TRIGGER IF EXISTS trigger_subscriptions_updated_at ON subscriptions;
CREATE TRIGGER trigger_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_subscriptions_updated_at();

-- ============================================
-- 6. Migrar usuarios existentes al plan gratuito
-- ============================================
-- Todos los usuarios existentes obtienen automáticamente una suscripción gratuita activa
INSERT INTO subscriptions (user_id, plan_id, status, current_period_start)
SELECT 
    id,
    'plan_free',
    'active',
    NOW()
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM subscriptions WHERE subscriptions.user_id = users.id
);

-- ============================================
-- COMENTARIOS FINALES
-- ============================================
-- Esta migración crea:
-- 1. Un sistema de suscripciones con estados claros
-- 2. Integración completa con MercadoPago (IDs sincronizados)
-- 3. Historial de pagos para auditoría
-- 4. Índices optimizados para consultas rápidas
-- 5. Migración automática de usuarios existentes al plan gratuito
-- 
-- Próximos pasos:
-- - Crear modelos en Go (subscription.go)
-- - Crear repositorio (subscription_repository.go)
-- - Integrar con handlers de registro de usuarios
-- - Implementar endpoints de API para gestión de suscripciones

