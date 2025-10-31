-- Remove stock alert columns from products table
ALTER TABLE products DROP COLUMN IF EXISTS notificado;
ALTER TABLE products DROP COLUMN IF EXISTS stock_minimo;
