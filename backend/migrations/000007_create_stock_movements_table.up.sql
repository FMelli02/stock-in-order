-- Migration: Create stock_movements table
-- Purpose: Ledger of inventory changes per product

BEGIN;

CREATE TABLE IF NOT EXISTS stock_movements (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity_change INTEGER NOT NULL,
    reason TEXT NOT NULL,
    reference_id TEXT,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
