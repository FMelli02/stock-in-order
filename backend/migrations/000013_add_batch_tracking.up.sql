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
-- Solo migrar productos con stock > 0
-- Usar user_id = 1 (admin) para lotes de migración
INSERT INTO product_batches (product_id, user_id, quantity, lote_number)
SELECT 
    id, 
    COALESCE(user_id, 1) as user_id, -- Si product.user_id es NULL, usar admin (1)
    quantity, 
    'LOTE_INICIAL_MIGRADO'
FROM products 
WHERE quantity > 0;

-- 4. La Cirugía: Eliminar columna quantity de products
ALTER TABLE products DROP COLUMN IF EXISTS quantity;

COMMIT;
