CREATE TABLE integrations (
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

CREATE INDEX idx_integrations_user_id ON integrations(user_id);
CREATE INDEX idx_integrations_platform ON integrations(platform);
