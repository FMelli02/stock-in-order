-- Migration rollback: Remove batch tracking fields from purchase_order_items

BEGIN;

ALTER TABLE purchase_order_items
DROP COLUMN IF EXISTS lote_number,
DROP COLUMN IF EXISTS expiry_date;

COMMIT;
