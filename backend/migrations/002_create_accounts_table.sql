-- Migration 002: Crear tabla accounts
-- Las cuentas son la unidad fundamental de organización del sistema

-- Crear ENUM para tipo de cuenta
CREATE TYPE account_type AS ENUM ('personal', 'family');

-- Crear ENUM para monedas soportadas
CREATE TYPE currency AS ENUM ('ARS', 'USD');

-- Crear tabla accounts
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type account_type NOT NULL,
    currency currency NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para mejorar performance
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(type);

-- Comentarios para documentación
COMMENT ON TABLE accounts IS 'Cuentas financieras - cada cuenta pertenece a un usuario y contiene datos completamente aislados';
COMMENT ON COLUMN accounts.id IS 'Identificador único de la cuenta';
COMMENT ON COLUMN accounts.user_id IS 'Usuario propietario de esta cuenta';
COMMENT ON COLUMN accounts.name IS 'Nombre descriptivo como "Finanzas Personales" o "Gastos Familia"';
COMMENT ON COLUMN accounts.type IS 'Tipo de cuenta: personal (individual) o family (familiar con múltiples miembros)';
COMMENT ON COLUMN accounts.currency IS 'Moneda base preferida para visualizaciones consolidadas (ARS o USD)';

-- Constraint check: el nombre no puede estar vacío
ALTER TABLE accounts ADD CONSTRAINT accounts_name_not_empty CHECK (LENGTH(TRIM(name)) > 0);
