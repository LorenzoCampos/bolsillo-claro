-- Migration 013: Create recurring_expenses table for template-based recurring expenses
-- Date: 2026-01-18
-- Description: Implements "Recurring Templates Pattern" - separates recurring expense templates
--              from actual expense occurrences for better statistics, editability, and traceability

-- ====================
-- 1. CREATE ENUM TYPE
-- ====================

-- Define recurrence frequency options
CREATE TYPE recurrence_frequency AS ENUM ('daily', 'weekly', 'monthly', 'yearly');

COMMENT ON TYPE recurrence_frequency IS 'Frequency options for recurring expenses/incomes';

-- ====================
-- 2. CREATE TABLE
-- ====================

CREATE TABLE recurring_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    
    -- Expense details (same as expenses table)
    description TEXT NOT NULL,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    currency currency NOT NULL,
    category_id UUID REFERENCES expense_categories(id) ON DELETE SET NULL,
    family_member_id UUID REFERENCES family_members(id) ON DELETE SET NULL,
    
    -- Recurrence configuration
    recurrence_frequency recurrence_frequency NOT NULL,
    recurrence_interval INT NOT NULL DEFAULT 1 CHECK (recurrence_interval > 0),
    recurrence_day_of_month INT CHECK (recurrence_day_of_month BETWEEN 1 AND 31),
    recurrence_day_of_week INT CHECK (recurrence_day_of_week BETWEEN 0 AND 6),
    
    -- Time boundaries
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE CHECK (end_date IS NULL OR end_date >= start_date),
    total_occurrences INT CHECK (total_occurrences IS NULL OR total_occurrences > 0),
    current_occurrence INT DEFAULT 0 CHECK (current_occurrence >= 0),
    
    -- Multi-currency support (optional - for templates that specify exchange rate)
    exchange_rate NUMERIC(15,6) CHECK (exchange_rate IS NULL OR exchange_rate > 0),
    amount_in_primary_currency NUMERIC(15,2) CHECK (amount_in_primary_currency IS NULL OR amount_in_primary_currency > 0),
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ====================
-- 3. COMMENTS (Documentation)
-- ====================

COMMENT ON TABLE recurring_expenses IS 'Templates for recurring expenses. Actual occurrences are generated in expenses table with recurring_expense_id FK';
COMMENT ON COLUMN recurring_expenses.recurrence_frequency IS 'How often: daily, weekly, monthly, yearly';
COMMENT ON COLUMN recurring_expenses.recurrence_interval IS 'Every N periods (e.g., 2 = every 2 weeks)';
COMMENT ON COLUMN recurring_expenses.recurrence_day_of_month IS 'For monthly/yearly: day of month (1-31). NULL for other frequencies';
COMMENT ON COLUMN recurring_expenses.recurrence_day_of_week IS 'For weekly: day of week (0=Sunday, 6=Saturday). NULL for other frequencies';
COMMENT ON COLUMN recurring_expenses.start_date IS 'When to start generating occurrences';
COMMENT ON COLUMN recurring_expenses.end_date IS 'When to stop generating. NULL = indefinite';
COMMENT ON COLUMN recurring_expenses.total_occurrences IS 'Max number of occurrences (e.g., 6 for 6 installments). NULL = indefinite';
COMMENT ON COLUMN recurring_expenses.current_occurrence IS 'Counter: how many occurrences have been generated so far';
COMMENT ON COLUMN recurring_expenses.is_active IS 'If false, stops generating new occurrences (soft delete)';

-- ====================
-- 4. INDEXES (for CRON performance)
-- ====================

CREATE INDEX idx_recurring_expenses_account_id ON recurring_expenses(account_id);
CREATE INDEX idx_recurring_expenses_is_active ON recurring_expenses(is_active) WHERE is_active = true;
CREATE INDEX idx_recurring_expenses_frequency ON recurring_expenses(recurrence_frequency);
CREATE INDEX idx_recurring_expenses_next_occurrence ON recurring_expenses(start_date, end_date) WHERE is_active = true;

COMMENT ON INDEX idx_recurring_expenses_is_active IS 'Optimize CRON queries for active templates only';
COMMENT ON INDEX idx_recurring_expenses_next_occurrence IS 'Optimize CRON queries to find templates that need generation';

-- ====================
-- 5. TRIGGER (auto-update updated_at)
-- ====================

CREATE TRIGGER trigger_update_recurring_expenses_updated_at
BEFORE UPDATE ON recurring_expenses
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- ====================
-- 6. ADD FK TO EXPENSES TABLE
-- ====================

-- Link expenses to their recurring template (if they were auto-generated)
ALTER TABLE expenses 
ADD COLUMN recurring_expense_id UUID REFERENCES recurring_expenses(id) ON DELETE SET NULL;

CREATE INDEX idx_expenses_recurring_expense_id ON expenses(recurring_expense_id);

COMMENT ON COLUMN expenses.recurring_expense_id IS 'FK to recurring_expenses if this expense was auto-generated from a template. NULL for one-time expenses';

-- ====================
-- 7. VALIDATION CONSTRAINTS (business logic)
-- ====================

-- Monthly/yearly recurrence REQUIRES day_of_month
ALTER TABLE recurring_expenses 
ADD CONSTRAINT check_monthly_requires_day_of_month 
CHECK (
    (recurrence_frequency IN ('monthly', 'yearly') AND recurrence_day_of_month IS NOT NULL)
    OR
    (recurrence_frequency NOT IN ('monthly', 'yearly') AND recurrence_day_of_month IS NULL)
);

-- Weekly recurrence REQUIRES day_of_week
ALTER TABLE recurring_expenses 
ADD CONSTRAINT check_weekly_requires_day_of_week 
CHECK (
    (recurrence_frequency = 'weekly' AND recurrence_day_of_week IS NOT NULL)
    OR
    (recurrence_frequency != 'weekly' AND recurrence_day_of_week IS NULL)
);

-- Current occurrence cannot exceed total occurrences
ALTER TABLE recurring_expenses 
ADD CONSTRAINT check_current_occurrence_within_total 
CHECK (
    total_occurrences IS NULL 
    OR 
    current_occurrence <= total_occurrences
);

-- If has total_occurrences, must have end_date OR will calculate based on frequency
-- (This is a soft constraint - handled in application logic)

COMMENT ON CONSTRAINT check_monthly_requires_day_of_month ON recurring_expenses IS 'Monthly/yearly templates must specify which day of month';
COMMENT ON CONSTRAINT check_weekly_requires_day_of_week ON recurring_expenses IS 'Weekly templates must specify which day of week';
COMMENT ON CONSTRAINT check_current_occurrence_within_total ON recurring_expenses IS 'Cannot generate more occurrences than total_occurrences limit';

-- ====================
-- 8. MIGRATE EXISTING DATA (if any)
-- ====================

-- NOTE: This migration does NOT automatically migrate existing recurring expenses
-- Reason: Current expense_type='recurring' in expenses table is too simple (only has date + end_date)
-- Migration strategy: Manual or via separate data migration script
-- Users can recreate their recurring expenses using the new system

-- Future consideration: Create a data migration script that:
-- 1. Finds all expenses with expense_type='recurring' and end_date IS NOT NULL
-- 2. Creates a recurring_expense template for each unique (description, amount, account_id)
-- 3. Links existing expense rows to the new template via recurring_expense_id

-- ====================
-- MIGRATION COMPLETE
-- ====================

-- Summary of changes:
-- ✅ Created recurrence_frequency ENUM (daily, weekly, monthly, yearly)
-- ✅ Created recurring_expenses table (templates)
-- ✅ Added indexes for CRON performance
-- ✅ Added updated_at trigger
-- ✅ Added FK recurring_expense_id to expenses table
-- ✅ Added business logic constraints (day_of_month, day_of_week validation)
