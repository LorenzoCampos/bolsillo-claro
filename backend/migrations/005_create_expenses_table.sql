-- Create expense_type ENUM
CREATE TYPE expense_type AS ENUM ('one-time', 'recurring');

-- Create expenses table
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    family_member_id UUID REFERENCES family_members(id) ON DELETE SET NULL,
    category TEXT, -- Simple text category for now, can be normalized later
    description TEXT NOT NULL,
    amount DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    currency currency NOT NULL,
    expense_type expense_type NOT NULL DEFAULT 'one-time',
    date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: one-time expenses cannot have end_date, recurring can optionally have it
    CONSTRAINT check_recurring_end_date CHECK (
        (expense_type = 'one-time' AND end_date IS NULL) OR
        (expense_type = 'recurring' AND (end_date IS NULL OR end_date >= date))
    )
);

-- Create indexes for better query performance
CREATE INDEX idx_expenses_account_id ON expenses(account_id);
CREATE INDEX idx_expenses_family_member_id ON expenses(family_member_id);
CREATE INDEX idx_expenses_date ON expenses(date);
CREATE INDEX idx_expenses_expense_type ON expenses(expense_type);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_expenses_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER expenses_updated_at
    BEFORE UPDATE ON expenses
    FOR EACH ROW
    EXECUTE FUNCTION update_expenses_updated_at();
