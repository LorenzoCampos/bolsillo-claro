-- Migration: 014 - Add UNIQUE constraint for account names per user
-- Date: 2026-01-18
-- Description: Prevent duplicate account names for the same user (case-insensitive)
-- 
-- Rationale:
-- - Improves UX by forcing users to use descriptive, unique names
-- - Prevents confusion in UI when listing accounts
-- - Case-insensitive to handle "Finanzas Personales" vs "finanzas personales"
--
-- Example blocked scenario:
-- User creates "Gastos Familia" → OK
-- User tries to create "Gastos Familia" again → ERROR
-- User tries to create "gastos familia" → ERROR (case-insensitive)
-- User creates "Gastos Trabajo" → OK (different name)

-- ============================================================================
-- ADD UNIQUE CONSTRAINT
-- ============================================================================

-- Create unique partial index on (user_id, LOWER(name))
-- This prevents duplicate names per user (case-insensitive)
CREATE UNIQUE INDEX idx_accounts_unique_name_per_user 
ON accounts(user_id, LOWER(name));

-- ============================================================================
-- ROLLBACK INSTRUCTIONS
-- ============================================================================
-- To rollback this migration:
-- DROP INDEX idx_accounts_unique_name_per_user;

-- ============================================================================
-- TESTING
-- ============================================================================
-- Test duplicate prevention:
-- INSERT INTO accounts (id, user_id, name, type, currency) 
-- VALUES (uuid_generate_v4(), 'some-user-id', 'Test Account', 'personal', 'ARS');
-- 
-- INSERT INTO accounts (id, user_id, name, type, currency) 
-- VALUES (uuid_generate_v4(), 'some-user-id', 'Test Account', 'personal', 'ARS');
-- ^ This should FAIL with: duplicate key value violates unique constraint
--
-- INSERT INTO accounts (id, user_id, name, type, currency) 
-- VALUES (uuid_generate_v4(), 'some-user-id', 'test account', 'personal', 'ARS');
-- ^ This should also FAIL (case-insensitive match)
