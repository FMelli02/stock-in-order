-- Crear tabla de suscripciones
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id TEXT NOT NULL DEFAULT 'free',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'expired', 'pending')),
    mercadopago_subscription_id TEXT,
    mercadopago_payment_id TEXT,
    start_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date TIMESTAMPTZ,
    auto_renew BOOLEAN NOT NULL DEFAULT false,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices para mejorar performance
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_mercadopago_subscription_id ON subscriptions(mercadopago_subscription_id) WHERE mercadopago_subscription_id IS NOT NULL;

-- Constraint: Solo una suscripción activa por usuario
CREATE UNIQUE INDEX idx_one_active_subscription_per_user 
ON subscriptions(user_id) 
WHERE status = 'active';

-- Trigger para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_subscriptions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_subscriptions_updated_at();

-- Crear suscripciones FREE para todos los usuarios existentes
INSERT INTO subscriptions (user_id, plan_id, status, start_date, auto_renew)
SELECT 
    id,
    'free',
    'active',
    NOW(),
    false
FROM users
ON CONFLICT DO NOTHING;
