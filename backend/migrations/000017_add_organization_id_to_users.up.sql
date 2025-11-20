-- Migration: Add organization_id to users table
-- Descripción: Permite que múltiples usuarios trabajen en el mismo inventario
-- Un admin es dueño de la organización (organization_id = su propio id)
-- Vendedores y repositores pertenecen a la organización del admin que los creó

BEGIN;

-- Agregar columna organization_id
ALTER TABLE users ADD COLUMN IF NOT EXISTS organization_id BIGINT;

-- Crear foreign key hacia users (auto-referencia)
ALTER TABLE users ADD CONSTRAINT fk_users_organization 
    FOREIGN KEY (organization_id) REFERENCES users(id) ON DELETE CASCADE;

-- Para usuarios admin existentes, su organization_id es su propio id
UPDATE users SET organization_id = id WHERE role = 'admin' AND organization_id IS NULL;

-- Crear índice para mejorar performance en consultas por organización
CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id);

COMMIT;
