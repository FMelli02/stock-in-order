-- Add stock alert columns to products table
ALTER TABLE products ADD COLUMN stock_minimo INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN notificado BOOLEAN NOT NULL DEFAULT false;
