-- Rollback Migration: Remove Batch Tracking System
-- Revierte los cambios de 000013_add_batch_tracking.up.sql

BEGIN;

-- 1. Restaurar columna quantity en products
ALTER TABLE products ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 0;

-- 2. Migrar stock de vuelta desde product_batches a products
-- Sumar todas las cantidades de lotes por producto
UPDATE products p
SET quantity = (
    SELECT COALESCE(SUM(pb.quantity), 0)
    FROM product_batches pb
    WHERE pb.product_id = p.id
);

-- 3. Eliminar índices
DROP INDEX IF EXISTS idx_product_batches_user_id;
DROP INDEX IF EXISTS idx_product_batches_expiry_date;
DROP INDEX IF EXISTS idx_product_batches_product_id;

-- 4. Eliminar tabla product_batches
DROP TABLE IF EXISTS product_batches;

COMMIT;
