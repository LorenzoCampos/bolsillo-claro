-- Create income_type ENUM (same logic as expenses)
CREATE TYPE income_type AS ENUM ('one-time', 'recurring');

-- Create incomes table
CREATE TABLE incomes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    family_member_id UUID REFERENCES family_members(id) ON DELETE SET NULL,
    category TEXT, -- Simple text category (e.g., "Salario", "Freelance", "Inversiones")
    description TEXT NOT NULL,
    amount DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    currency currency NOT NULL,
    income_type income_type NOT NULL DEFAULT 'one-time',
    date DATE NOT NULL,
    end_date DATE, -- Optional: for recurring incomes (e.g., salary contract ends)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: one-time incomes cannot have end_date, recurring can optionally have it
    CONSTRAINT check_recurring_end_date CHECK (
        (income_type = 'one-time' AND end_date IS NULL) OR
        (income_type = 'recurring' AND (end_date IS NULL OR end_date >= date))
    )
);

-- Create indexes for better query performance
CREATE INDEX idx_incomes_account_id ON incomes(account_id);
CREATE INDEX idx_incomes_family_member_id ON incomes(family_member_id);
CREATE INDEX idx_incomes_date ON incomes(date);
CREATE INDEX idx_incomes_income_type ON incomes(income_type);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_incomes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER incomes_updated_at
    BEFORE UPDATE ON incomes
    FOR EACH ROW
    EXECUTE FUNCTION update_incomes_updated_at();
