-- ZarishLog — Add safety_stock to products (Phase 6)
-- PostgreSQL 18 — Adds safety_stock column for reorder planning.
-- Idempotent: uses IF NOT EXISTS via information_schema.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='safety_stock') THEN
    ALTER TABLE products ADD COLUMN safety_stock numeric(12,3) DEFAULT 0;
  END IF;
END;
$$;

CREATE INDEX IF NOT EXISTS idx_products_safety_stock ON products(safety_stock);
