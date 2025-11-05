-- Migration: Add batch tracking fields to purchase_order_items
-- Adds lote_number and expiry_date to support batch creation on order completion

BEGIN;

ALTER TABLE purchase_order_items
ADD COLUMN lote_number TEXT,
ADD COLUMN expiry_date TIMESTAMPTZ;

COMMIT;
