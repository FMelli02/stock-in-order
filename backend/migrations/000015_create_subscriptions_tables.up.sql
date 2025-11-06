-- Migración 000015: Sistema de Suscripciones y Pagos
-- Descripción: Crea las tablas necesarias para gestionar planes, suscripciones y pagos con MercadoPago
-- Fecha: 6 de Noviembre, 2025

-- ============================================
-- 1. ENUM para estados de suscripción
-- ============================================
CREATE TYPE subscription_status AS ENUM (
    'active',      -- Suscripción activa y pagada
    'inactive',    -- Suscripción inactiva (recién creada, sin pago)
    'past_due',    -- Pago vencido, en período de gracia
    'canceled'     -- Cancelada por el usuario o por falta de pago
);

-- ============================================
-- 2. Tabla de Suscripciones
-- ============================================
CREATE TABLE subscriptions (
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
CREATE TABLE payment_history (
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
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);

-- Búsqueda de suscripciones próximas a vencer (para notificaciones)
CREATE INDEX idx_subscriptions_period_end ON subscriptions(current_period_end);

-- IDs de MercadoPago para webhooks
CREATE INDEX idx_subscriptions_mp_subscription_id ON subscriptions(mp_subscription_id);
CREATE INDEX idx_subscriptions_mp_customer_id ON subscriptions(mp_customer_id);

-- Historial de pagos
CREATE INDEX idx_payment_history_subscription_id ON payment_history(subscription_id);
CREATE INDEX idx_payment_history_user_id ON payment_history(user_id);
CREATE INDEX idx_payment_history_mp_payment_id ON payment_history(mp_payment_id);

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
