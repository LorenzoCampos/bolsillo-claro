-- Migration 016: Create recurring_incomes table for template-based recurring incomes
-- Date: 2026-01-19
-- Description: Implements "Recurring Templates Pattern" for incomes - separates recurring income templates
--              from actual income occurrences for better statistics, editability, and traceability
--              Mirrors the architecture of recurring_expenses (migration 013)

-- ====================
-- 1. CREATE TABLE (reuses recurrence_frequency ENUM from migration 013)
-- ====================

CREATE TABLE recurring_incomes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    
    -- Income details (same as incomes table)
    description TEXT NOT NULL,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    currency currency NOT NULL,
    category_id UUID REFERENCES income_categories(id) ON DELETE SET NULL,
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
-- 2. COMMENTS (Documentation)
-- ====================

COMMENT ON TABLE recurring_incomes IS 'Templates for recurring incomes. Actual occurrences are generated in incomes table with recurring_income_id FK';
COMMENT ON COLUMN recurring_incomes.recurrence_frequency IS 'How often: daily, weekly, monthly, yearly';
COMMENT ON COLUMN recurring_incomes.recurrence_interval IS 'Every N periods (e.g., 2 = every 2 weeks)';
COMMENT ON COLUMN recurring_incomes.recurrence_day_of_month IS 'For monthly/yearly: day of month (1-31). NULL for other frequencies';
COMMENT ON COLUMN recurring_incomes.recurrence_day_of_week IS 'For weekly: day of week (0=Sunday, 6=Saturday). NULL for other frequencies';
COMMENT ON COLUMN recurring_incomes.start_date IS 'When to start generating occurrences';
COMMENT ON COLUMN recurring_incomes.end_date IS 'When to stop generating. NULL = indefinite';
COMMENT ON COLUMN recurring_incomes.total_occurrences IS 'Max number of occurrences (e.g., 12 for 12 monthly salaries). NULL = indefinite';
COMMENT ON COLUMN recurring_incomes.current_occurrence IS 'Counter: how many occurrences have been generated so far';
COMMENT ON COLUMN recurring_incomes.is_active IS 'If false, stops generating new occurrences (soft delete)';

-- ====================
-- 3. INDEXES (for CRON performance)
-- ====================

CREATE INDEX idx_recurring_incomes_account_id ON recurring_incomes(account_id);
CREATE INDEX idx_recurring_incomes_is_active ON recurring_incomes(is_active) WHERE is_active = true;
CREATE INDEX idx_recurring_incomes_frequency ON recurring_incomes(recurrence_frequency);
CREATE INDEX idx_recurring_incomes_next_occurrence ON recurring_incomes(start_date, end_date) WHERE is_active = true;

COMMENT ON INDEX idx_recurring_incomes_is_active IS 'Optimize CRON queries for active templates only';
COMMENT ON INDEX idx_recurring_incomes_next_occurrence IS 'Optimize CRON queries to find templates that need generation';

-- ====================
-- 4. TRIGGER (auto-update updated_at)
-- ====================

CREATE TRIGGER trigger_update_recurring_incomes_updated_at
BEFORE UPDATE ON recurring_incomes
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- ====================
-- 5. ADD FK TO INCOMES TABLE
-- ====================

-- Link incomes to their recurring template (if they were auto-generated)
ALTER TABLE incomes 
ADD COLUMN recurring_income_id UUID REFERENCES recurring_incomes(id) ON DELETE SET NULL;

CREATE INDEX idx_incomes_recurring_income_id ON incomes(recurring_income_id);

COMMENT ON COLUMN incomes.recurring_income_id IS 'FK to recurring_incomes if this income was auto-generated from a template. NULL for one-time incomes';

-- ====================
-- 6. VALIDATION CONSTRAINTS (business logic)
-- ====================

-- Monthly/yearly recurrence REQUIRES day_of_month
ALTER TABLE recurring_incomes 
ADD CONSTRAINT check_monthly_requires_day_of_month 
CHECK (
    (recurrence_frequency IN ('monthly', 'yearly') AND recurrence_day_of_month IS NOT NULL)
    OR
    (recurrence_frequency NOT IN ('monthly', 'yearly') AND recurrence_day_of_month IS NULL)
);

-- Weekly recurrence REQUIRES day_of_week
ALTER TABLE recurring_incomes 
ADD CONSTRAINT check_weekly_requires_day_of_week 
CHECK (
    (recurrence_frequency = 'weekly' AND recurrence_day_of_week IS NOT NULL)
    OR
    (recurrence_frequency != 'weekly' AND recurrence_day_of_week IS NULL)
);

-- Current occurrence cannot exceed total occurrences
ALTER TABLE recurring_incomes 
ADD CONSTRAINT check_current_occurrence_within_total 
CHECK (
    total_occurrences IS NULL 
    OR 
    current_occurrence <= total_occurrences
);

COMMENT ON CONSTRAINT check_monthly_requires_day_of_month ON recurring_incomes IS 'Monthly/yearly templates must specify which day of month';
COMMENT ON CONSTRAINT check_weekly_requires_day_of_week ON recurring_incomes IS 'Weekly templates must specify which day of week';
COMMENT ON CONSTRAINT check_current_occurrence_within_total ON recurring_incomes IS 'Cannot generate more occurrences than total_occurrences limit';

-- ====================
-- 7. MIGRATE EXISTING DATA (if any)
-- ====================

-- NOTE: This migration does NOT automatically migrate existing recurring incomes
-- Reason: Current income_type='recurring' in incomes table is too simple (only has date + end_date)
-- Migration strategy: Manual or via separate data migration script
-- Users can recreate their recurring incomes using the new system

-- Future consideration: Create a data migration script that:
-- 1. Finds all incomes with income_type='recurring' and end_date IS NOT NULL
-- 2. Creates a recurring_income template for each unique (description, amount, account_id)
-- 3. Links existing income rows to the new template via recurring_income_id

-- ====================
-- MIGRATION COMPLETE
-- ====================

-- Summary of changes:
-- ✅ Created recurring_incomes table (templates) - mirrors recurring_expenses
-- ✅ Reused recurrence_frequency ENUM (daily, weekly, monthly, yearly)
-- ✅ Added indexes for CRON performance
-- ✅ Added updated_at trigger
-- ✅ Added FK recurring_income_id to incomes table
-- ✅ Added business logic constraints (day_of_month, day_of_week validation)
