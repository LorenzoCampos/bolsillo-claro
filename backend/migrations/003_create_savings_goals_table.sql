-- Migration 003: Crear tabla savings_goals
-- Almacena las metas de ahorro de cada cuenta

CREATE TABLE IF NOT EXISTS savings_goals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_amount DECIMAL(15,2) NOT NULL CHECK (target_amount > 0),
    current_amount DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (current_amount >= 0),
    currency currency NOT NULL,
    deadline DATE,
    is_general BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_savings_goals_account_id ON savings_goals(account_id);
CREATE INDEX IF NOT EXISTS idx_savings_goals_is_general ON savings_goals(is_general);

-- Constraint: solo puede haber una meta general por cuenta
CREATE UNIQUE INDEX IF NOT EXISTS idx_savings_goals_one_general_per_account 
    ON savings_goals(account_id) WHERE is_general = true;

-- Constraint: current_amount no puede ser mayor que target_amount
ALTER TABLE savings_goals ADD CONSTRAINT savings_goals_current_lte_target 
    CHECK (current_amount <= target_amount);

-- Comentarios
COMMENT ON TABLE savings_goals IS 'Metas de ahorro - cada cuenta tiene una meta general más metas específicas opcionales';
COMMENT ON COLUMN savings_goals.id IS 'Identificador único de la meta';
COMMENT ON COLUMN savings_goals.account_id IS 'Cuenta a la que pertenece esta meta';
COMMENT ON COLUMN savings_goals.name IS 'Nombre de la meta como "Vacaciones", "Auto nuevo", "Ahorro General"';
COMMENT ON COLUMN savings_goals.target_amount IS 'Monto objetivo a alcanzar';
COMMENT ON COLUMN savings_goals.current_amount IS 'Monto actualmente ahorrado - se actualiza automáticamente';
COMMENT ON COLUMN savings_goals.currency IS 'Moneda de la meta (ARS o USD)';
COMMENT ON COLUMN savings_goals.deadline IS 'Fecha objetivo - null significa sin deadline';
COMMENT ON COLUMN savings_goals.is_general IS 'True solo para la meta de Ahorro General de cada cuenta';
