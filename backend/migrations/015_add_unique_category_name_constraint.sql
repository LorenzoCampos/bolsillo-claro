-- Migration: 015 - Add UNIQUE constraint for category names per account (case-insensitive)
-- Date: 2026-01-19
-- Description: Prevent duplicate category names for the same account (case-insensitive)
-- 
-- Rationale:
-- - Improves UX by preventing duplicate category names within the same account
-- - Case-insensitive to handle "Alimentación" vs "alimentación"
-- - Consistent with accounts module behavior (migration 014)
-- - Only applies to custom categories (is_system = false)
--
-- Example blocked scenarios:
-- Account A creates expense category "Alimentación" → OK
-- Account A tries to create "Alimentación" again → ERROR
-- Account A tries to create "alimentación" → ERROR (case-insensitive)
-- Account A creates "Transporte" → OK (different name)
-- Account B creates "Alimentación" → OK (different account)
-- System category "Alimentación" (is_system=true) → OK (not affected by constraint)

-- ============================================================================
-- ADD UNIQUE CONSTRAINTS
-- ============================================================================

-- Drop existing case-sensitive constraints first
ALTER TABLE expense_categories 
DROP CONSTRAINT IF EXISTS unique_expense_category_custom;

ALTER TABLE income_categories 
DROP CONSTRAINT IF EXISTS unique_income_category_custom;

-- Create unique case-insensitive index for expense_categories
-- Only applies to custom categories (is_system = false)
CREATE UNIQUE INDEX idx_expense_categories_unique_name_per_account 
ON expense_categories(account_id, LOWER(name)) 
WHERE is_system = false;

-- Create unique case-insensitive index for income_categories
-- Only applies to custom categories (is_system = false)
CREATE UNIQUE INDEX idx_income_categories_unique_name_per_account 
ON income_categories(account_id, LOWER(name)) 
WHERE is_system = false;

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- 
-- DROP INDEX IF EXISTS idx_expense_categories_unique_name_per_account;
-- DROP INDEX IF EXISTS idx_income_categories_unique_name_per_account;
-- 
-- -- Restore original case-sensitive constraints:
-- ALTER TABLE expense_categories 
-- ADD CONSTRAINT unique_expense_category_custom 
-- UNIQUE (account_id, name) 
-- WHERE is_system = false;
-- 
-- ALTER TABLE income_categories 
-- ADD CONSTRAINT unique_income_category_custom 
-- UNIQUE (account_id, name) 
-- WHERE is_system = false;

-- ============================================================================
-- TESTING
-- ============================================================================
-- Test duplicate prevention (expense_categories):
-- INSERT INTO expense_categories (id, account_id, name, icon, color, is_system) 
-- VALUES (uuid_generate_v4(), 'some-account-id', 'Test Category', '🔥', '#FF0000', false);
-- 
-- INSERT INTO expense_categories (id, account_id, name, icon, color, is_system) 
-- VALUES (uuid_generate_v4(), 'some-account-id', 'Test Category', '💰', '#00FF00', false);
-- ^ This should FAIL with: duplicate key value violates unique constraint
--
-- INSERT INTO expense_categories (id, account_id, name, icon, color, is_system) 
-- VALUES (uuid_generate_v4(), 'some-account-id', 'test category', '💸', '#0000FF', false);
-- ^ This should also FAIL (case-insensitive match)
--
-- Test different account (should work):
-- INSERT INTO expense_categories (id, account_id, name, icon, color, is_system) 
-- VALUES (uuid_generate_v4(), 'different-account-id', 'Test Category', '🎯', '#FFFF00', false);
-- ^ This should SUCCEED (different account_id)
--
-- Test system category (should work):
-- INSERT INTO expense_categories (id, account_id, name, icon, color, is_system) 
-- VALUES (uuid_generate_v4(), NULL, 'Test Category', '⭐', '#FF00FF', true);
-- ^ This should SUCCEED (is_system = true, not affected by constraint)
