-- Migración 000015 DOWN: Revertir sistema de suscripciones
-- Fecha: 6 de Noviembre, 2025

-- Eliminar trigger
DROP TRIGGER IF EXISTS trigger_subscriptions_updated_at ON subscriptions;

-- Eliminar función
DROP FUNCTION IF EXISTS update_subscriptions_updated_at();

-- Eliminar índices
DROP INDEX IF EXISTS idx_payment_history_mp_payment_id;
DROP INDEX IF EXISTS idx_payment_history_user_id;
DROP INDEX IF EXISTS idx_payment_history_subscription_id;

DROP INDEX IF EXISTS idx_subscriptions_mp_customer_id;
DROP INDEX IF EXISTS idx_subscriptions_mp_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_period_end;
DROP INDEX IF EXISTS idx_subscriptions_plan_id;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user_id;

-- Eliminar tablas (en orden inverso por dependencias)
DROP TABLE IF EXISTS payment_history;
DROP TABLE IF EXISTS subscriptions;

-- Eliminar tipos
DROP TYPE IF EXISTS subscription_status;
