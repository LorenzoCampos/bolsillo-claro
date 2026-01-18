-- Migration 012: Add trigger to automatically update updated_at column
-- Este trigger garantiza que el campo updated_at se actualice automáticamente
-- cada vez que se modifica un registro en la tabla users

-- ============================================================================
-- PASO 1: Crear función genérica para actualizar updated_at
-- ============================================================================
-- Esta función puede reutilizarse para cualquier tabla que tenga updated_at

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    -- NEW es el registro después del UPDATE
    -- Actualizamos su campo updated_at al timestamp actual
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- PASO 2: Crear trigger en tabla users
-- ============================================================================
-- Este trigger se ejecuta ANTES de cada UPDATE en la tabla users

CREATE TRIGGER trigger_update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- PASO 3: Crear triggers en otras tablas que tienen updated_at
-- ============================================================================

-- Trigger para accounts
CREATE TRIGGER trigger_update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para expenses
CREATE TRIGGER trigger_update_expenses_updated_at
    BEFORE UPDATE ON expenses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para incomes
CREATE TRIGGER trigger_update_incomes_updated_at
    BEFORE UPDATE ON incomes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para savings_goals
CREATE TRIGGER trigger_update_savings_goals_updated_at
    BEFORE UPDATE ON savings_goals
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para expense_categories
CREATE TRIGGER trigger_update_expense_categories_updated_at
    BEFORE UPDATE ON expense_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para income_categories
CREATE TRIGGER trigger_update_income_categories_updated_at
    BEFORE UPDATE ON income_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- NOTAS:
-- ============================================================================
-- 1. El trigger se ejecuta ANTES del UPDATE (BEFORE UPDATE)
-- 2. Se ejecuta para CADA FILA modificada (FOR EACH ROW)
-- 3. La función es reutilizable para todas las tablas
-- 4. NOW() en PostgreSQL retorna el timestamp del inicio de la transacción
-- 5. Si necesitás precisión de microsegundos en cada trigger, usá CURRENT_TIMESTAMP
-- 
-- Ventajas:
-- - Automático: No necesitás recordar actualizar updated_at en el código
-- - Consistente: Siempre se actualiza, sin importar cómo se hace el UPDATE
-- - Performante: BEFORE trigger no genera queries adicionales
-- - Auditable: Sabés exactamente cuándo fue la última modificación
--
-- Uso:
-- UPDATE users SET name = 'Nuevo Nombre' WHERE id = 'uuid';
-- → updated_at se actualiza automáticamente a NOW()
