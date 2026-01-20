-- Migration 017: Add EUR to currency ENUM
-- Date: 2026-01-20
-- Description: Adds EUR (Euro) support to the currency ENUM type
--              This fixes the bug where handlers accept EUR but database rejects it

-- ============================================================================
-- ADD EUR TO CURRENCY ENUM
-- ============================================================================

-- PostgreSQL requires ALTER TYPE ... ADD VALUE for ENUMs
-- This is a non-blocking operation (doesn't require table locks)
-- Values are added at the end of the enum unless BEFORE/AFTER is specified

ALTER TYPE currency ADD VALUE IF NOT EXISTS 'EUR';

-- ============================================================================
-- VERIFICATION COMMENT
-- ============================================================================

COMMENT ON TYPE currency IS 'Supported currencies: ARS (Argentine Peso), USD (US Dollar), EUR (Euro)';

-- ============================================================================
-- NOTES
-- ============================================================================
-- After this migration:
-- 1. expenses.currency can accept: ARS, USD, EUR
-- 2. incomes.currency can accept: ARS, USD, EUR
-- 3. accounts.currency can accept: ARS, USD, EUR
-- 4. exchange_rates table can store EUR conversion rates
--
-- Exchange rate examples:
-- - 1 EUR = 1100 ARS (ejemplo: tasa del día)
-- - 1 USD = 1000 ARS
--
-- Multi-currency workflow:
-- 1. User creates expense in EUR (amount: 100)
-- 2. System calculates amount_in_primary_currency using exchange_rate
-- 3. If account.currency = ARS and exchange_rate = 1100:
--    amount_in_primary_currency = 100 * 1100 = 110,000 ARS
