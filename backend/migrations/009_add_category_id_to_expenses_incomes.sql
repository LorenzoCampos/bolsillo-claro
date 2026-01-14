-- Add category_id column to expenses table
ALTER TABLE expenses 
    ADD COLUMN category_id UUID REFERENCES expense_categories(id) ON DELETE SET NULL;

-- Add category_id column to incomes table
ALTER TABLE incomes 
    ADD COLUMN category_id UUID REFERENCES income_categories(id) ON DELETE SET NULL;

-- Create indexes for better query performance
CREATE INDEX idx_expenses_category_id ON expenses(category_id);
CREATE INDEX idx_incomes_category_id ON incomes(category_id);

-- Migrate existing data: Match TEXT category to category name
-- For expenses
UPDATE expenses e
SET category_id = ec.id
FROM expense_categories ec
WHERE e.category IS NOT NULL 
  AND LOWER(TRIM(e.category)) = LOWER(ec.name)
  AND ec.is_system = TRUE;

-- For incomes
UPDATE incomes i
SET category_id = ic.id
FROM income_categories ic
WHERE i.category IS NOT NULL 
  AND LOWER(TRIM(i.category)) = LOWER(ic.name)
  AND ic.is_system = TRUE;

-- Drop old TEXT columns (now using category_id)
ALTER TABLE expenses DROP COLUMN category;
ALTER TABLE incomes DROP COLUMN category;
