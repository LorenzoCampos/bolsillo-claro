-- Migration 010: Soporte multi-moneda con Modo 3 (Flexibilidad Total)
-- Agrega campos para soportar conversión de monedas con 3 modos:
-- Modo 1: Moneda local (ARS) - automático
-- Modo 2: Con exchange_rate provisto - sistema calcula amount_in_primary_currency
-- Modo 3: Con amount_in_primary_currency provisto - sistema calcula exchange_rate efectivo

-- ============================================================================
-- TABLA: exchange_rates
-- ============================================================================
-- Tabla para almacenar tasas de cambio históricas
-- Esto permite obtener tasas automáticas si el usuario no provee una
CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_currency currency NOT NULL,     -- Moneda origen (USD, EUR)
    to_currency currency NOT NULL,       -- Moneda destino (ARS)
    rate DECIMAL(15, 6) NOT NULL,        -- Tasa de conversión (ej: 1575.50)
    rate_date DATE NOT NULL,             -- Fecha de la tasa
    source VARCHAR(100),                 -- Fuente: 'manual', 'bcra', 'api', etc.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraint: solo una tasa por combinación de monedas por día
    UNIQUE(from_currency, to_currency, rate_date)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_exchange_rates_from_to_date 
    ON exchange_rates(from_currency, to_currency, rate_date DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_rates_date 
    ON exchange_rates(rate_date DESC);

-- Comentarios
COMMENT ON TABLE exchange_rates IS 'Tasas de cambio históricas para conversión automática';
COMMENT ON COLUMN exchange_rates.from_currency IS 'Moneda origen (USD, EUR, etc.)';
COMMENT ON COLUMN exchange_rates.to_currency IS 'Moneda destino (normalmente ARS)';
COMMENT ON COLUMN exchange_rates.rate IS 'Tasa de conversión - 1 from_currency = X to_currency';
COMMENT ON COLUMN exchange_rates.rate_date IS 'Fecha para la cual es válida esta tasa';
COMMENT ON COLUMN exchange_rates.source IS 'Origen de la tasa: manual, bcra, api, etc.';

-- ============================================================================
-- MODIFICAR: expenses
-- ============================================================================
-- Agregar campos para multi-moneda con snapshot de tasa de cambio

-- Campo: exchange_rate (tasa de cambio en el momento de la transacción)
ALTER TABLE expenses 
ADD COLUMN IF NOT EXISTS exchange_rate DECIMAL(15, 6);

-- Campo: amount_in_primary_currency (monto convertido a la moneda primaria de la cuenta)
-- Este es el monto REAL que se pagó (captura dólar tarjeta, impuestos, etc.)
ALTER TABLE expenses 
ADD COLUMN IF NOT EXISTS amount_in_primary_currency DECIMAL(15, 2);

-- Comentarios
COMMENT ON COLUMN expenses.exchange_rate IS 'Snapshot de la tasa de cambio en el momento de la transacción. Para ARS=1.0, para USD puede ser tasa oficial o dólar tarjeta (efectiva)';
COMMENT ON COLUMN expenses.amount_in_primary_currency IS 'Monto en la moneda primaria de la cuenta (normalmente ARS). Representa lo que REALMENTE se pagó, incluyendo impuestos y recargos';

-- ============================================================================
-- MODIFICAR: incomes
-- ============================================================================
-- Agregar los mismos campos para incomes

ALTER TABLE incomes 
ADD COLUMN IF NOT EXISTS exchange_rate DECIMAL(15, 6);

ALTER TABLE incomes 
ADD COLUMN IF NOT EXISTS amount_in_primary_currency DECIMAL(15, 2);

COMMENT ON COLUMN incomes.exchange_rate IS 'Snapshot de la tasa de cambio en el momento de la transacción';
COMMENT ON COLUMN incomes.amount_in_primary_currency IS 'Monto en la moneda primaria de la cuenta (normalmente ARS)';

-- ============================================================================
-- MIGRACIÓN DE DATOS EXISTENTES
-- ============================================================================
-- Actualizar registros existentes:
-- - Si currency es igual a la primary currency de la cuenta → exchange_rate = 1.0
-- - amount_in_primary_currency = amount (porque antes no había multi-moneda)

-- Para expenses
UPDATE expenses e
SET 
    exchange_rate = 1.0,
    amount_in_primary_currency = e.amount
WHERE 
    exchange_rate IS NULL;

-- Para incomes
UPDATE incomes i
SET 
    exchange_rate = 1.0,
    amount_in_primary_currency = i.amount
WHERE 
    exchange_rate IS NULL;

-- ============================================================================
-- CONSTRAINTS
-- ============================================================================
-- Hacer los campos NOT NULL ahora que tienen valores
ALTER TABLE expenses 
ALTER COLUMN exchange_rate SET NOT NULL,
ALTER COLUMN amount_in_primary_currency SET NOT NULL;

ALTER TABLE incomes 
ALTER COLUMN exchange_rate SET NOT NULL,
ALTER COLUMN amount_in_primary_currency SET NOT NULL;

-- Validar que exchange_rate sea positivo
ALTER TABLE expenses
ADD CONSTRAINT expenses_exchange_rate_positive 
CHECK (exchange_rate > 0);

ALTER TABLE incomes
ADD CONSTRAINT incomes_exchange_rate_positive 
CHECK (exchange_rate > 0);

-- Validar que amount_in_primary_currency sea positivo
ALTER TABLE expenses
ADD CONSTRAINT expenses_amount_in_primary_currency_positive 
CHECK (amount_in_primary_currency > 0);

ALTER TABLE incomes
ADD CONSTRAINT incomes_amount_in_primary_currency_positive 
CHECK (amount_in_primary_currency > 0);
