-- Migration 011: Actualizar savings_goals y crear savings_goal_transactions
-- Agrega campo saved_in a savings_goals y crea tabla de transacciones

-- ============================================================================
-- MODIFICAR: savings_goals (agregar campo saved_in)
-- ============================================================================

-- Agregar campo description (opcional, descripción de la meta)
ALTER TABLE savings_goals 
ADD COLUMN IF NOT EXISTS description TEXT;

COMMENT ON COLUMN savings_goals.description IS 'Descripción opcional de la meta de ahorro';

-- Agregar campo saved_in (opcional, texto libre)
ALTER TABLE savings_goals 
ADD COLUMN IF NOT EXISTS saved_in VARCHAR(255);

COMMENT ON COLUMN savings_goals.saved_in IS 'Lugar donde se guarda el dinero: "Mercado Pago", "Plazo fijo", "Efectivo", etc.';

-- Agregar campo is_active (para soft delete)
ALTER TABLE savings_goals 
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

COMMENT ON COLUMN savings_goals.is_active IS 'Indica si la meta está activa (true) o eliminada lógicamente (false)';

-- Eliminar el constraint restrictivo que impide superar el 100%
-- (Queremos permitir que el usuario ahorre más del objetivo)
ALTER TABLE savings_goals
DROP CONSTRAINT IF EXISTS savings_goals_current_lte_target;

-- Eliminar columna is_general si existe (no la necesitamos)
ALTER TABLE savings_goals
DROP COLUMN IF EXISTS is_general CASCADE;

-- ============================================================================
-- CREAR: savings_goal_transactions
-- ============================================================================

CREATE TABLE IF NOT EXISTS savings_goal_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    savings_goal_id UUID NOT NULL REFERENCES savings_goals(id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL,                -- Monto de la transacción
    transaction_type VARCHAR(20) NOT NULL,         -- 'deposit' o 'withdrawal'
    description TEXT,                              -- Descripción opcional
    date DATE NOT NULL,                            -- Fecha de la transacción
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_savings_goal_transactions_goal_id 
    ON savings_goal_transactions(savings_goal_id);

CREATE INDEX IF NOT EXISTS idx_savings_goal_transactions_date 
    ON savings_goal_transactions(date DESC);

CREATE INDEX IF NOT EXISTS idx_savings_goal_transactions_type 
    ON savings_goal_transactions(transaction_type);

-- Constraints
ALTER TABLE savings_goal_transactions
ADD CONSTRAINT savings_goal_transactions_amount_positive 
CHECK (amount > 0);

ALTER TABLE savings_goal_transactions
ADD CONSTRAINT savings_goal_transactions_type_valid 
CHECK (transaction_type IN ('deposit', 'withdrawal'));

-- Comentarios
COMMENT ON TABLE savings_goal_transactions IS 'Historial de movimientos (depósitos y retiros) de las metas de ahorro';
COMMENT ON COLUMN savings_goal_transactions.id IS 'Identificador único de la transacción';
COMMENT ON COLUMN savings_goal_transactions.savings_goal_id IS 'Meta de ahorro a la que pertenece';
COMMENT ON COLUMN savings_goal_transactions.amount IS 'Monto de la transacción (siempre positivo, el tipo indica si es depósito o retiro)';
COMMENT ON COLUMN savings_goal_transactions.transaction_type IS 'Tipo: deposit (agregar fondos) o withdrawal (retirar fondos)';
COMMENT ON COLUMN savings_goal_transactions.description IS 'Descripción opcional: "Ahorro enero", "Adelanto para pasaje", etc.';
COMMENT ON COLUMN savings_goal_transactions.date IS 'Fecha de la transacción (no puede ser futura)';
