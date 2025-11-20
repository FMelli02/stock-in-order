-- Revertir migración de subscriptions
DROP TRIGGER IF EXISTS trigger_update_subscriptions_updated_at ON subscriptions;
DROP FUNCTION IF EXISTS update_subscriptions_updated_at();
DROP INDEX IF EXISTS idx_one_active_subscription_per_user;
DROP INDEX IF EXISTS idx_subscriptions_mercadopago_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_plan_id;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP TABLE IF EXISTS subscriptions;
