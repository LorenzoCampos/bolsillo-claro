-- Create expense_categories table
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE, -- NULL = predefinida (sistema)
    name TEXT NOT NULL,
    icon TEXT,  -- Emoji o nombre de icono para frontend
    color TEXT, -- Color hex para gráficos (#FF6B6B)
    is_system BOOLEAN DEFAULT FALSE, -- TRUE = categoría predefinida (no editable/borrable)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create income_categories table
CREATE TABLE income_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    icon TEXT,
    color TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_expense_categories_account_id ON expense_categories(account_id);
CREATE INDEX idx_expense_categories_is_system ON expense_categories(is_system);
CREATE INDEX idx_income_categories_account_id ON income_categories(account_id);
CREATE INDEX idx_income_categories_is_system ON income_categories(is_system);

-- Create unique constraints using expression indexes
-- For system categories (account_id IS NULL): name must be unique globally
-- For custom categories: name must be unique per account
CREATE UNIQUE INDEX unique_expense_category_system 
    ON expense_categories (name) 
    WHERE account_id IS NULL;

CREATE UNIQUE INDEX unique_expense_category_custom 
    ON expense_categories (account_id, name) 
    WHERE account_id IS NOT NULL;

CREATE UNIQUE INDEX unique_income_category_system 
    ON income_categories (name) 
    WHERE account_id IS NULL;

CREATE UNIQUE INDEX unique_income_category_custom 
    ON income_categories (account_id, name) 
    WHERE account_id IS NOT NULL;

-- Create updated_at triggers
CREATE OR REPLACE FUNCTION update_expense_categories_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER expense_categories_updated_at
    BEFORE UPDATE ON expense_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_expense_categories_updated_at();

CREATE OR REPLACE FUNCTION update_income_categories_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER income_categories_updated_at
    BEFORE UPDATE ON income_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_income_categories_updated_at();
