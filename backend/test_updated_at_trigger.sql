-- Test del trigger updated_at
-- Ejecutar con: psql -U postgres -d bolsillo_claro -f test_updated_at_trigger.sql

\echo '════════════════════════════════════════════════════════════'
\echo '🧪 TEST: Trigger updated_at automático'
\echo '════════════════════════════════════════════════════════════'
\echo ''

-- ============================================================================
-- TEST 1: Verificar que la función existe
-- ============================================================================
\echo 'TEST 1: Verificar función update_updated_at_column()'
SELECT 
    proname as function_name,
    pg_get_functiondef(oid) as definition
FROM pg_proc
WHERE proname = 'update_updated_at_column';

\echo ''
\echo 'Si la query anterior retornó 1 fila → ✓ Función existe'
\echo 'Si retornó 0 filas → ✗ Función NO existe (migración no aplicada)'
\echo ''

-- ============================================================================
-- TEST 2: Verificar que los triggers existen
-- ============================================================================
\echo 'TEST 2: Verificar triggers en todas las tablas'
SELECT 
    tablename,
    triggername
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE triggername LIKE '%update%updated_at%'
ORDER BY tablename;

\echo ''
\echo 'Triggers esperados:'
\echo '  - trigger_update_users_updated_at'
\echo '  - trigger_update_accounts_updated_at'
\echo '  - trigger_update_expenses_updated_at'
\echo '  - trigger_update_incomes_updated_at'
\echo '  - trigger_update_savings_goals_updated_at'
\echo '  - trigger_update_expense_categories_updated_at'
\echo '  - trigger_update_income_categories_updated_at'
\echo ''

-- ============================================================================
-- TEST 3: Test funcional en tabla users
-- ============================================================================
\echo 'TEST 3: Test funcional del trigger en users'
\echo ''

-- Crear un usuario de prueba si no existe
INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
VALUES (
    'test-trigger-user-id',
    'trigger_test@example.com',
    'dummy_hash',
    'Trigger Test User',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Ver el updated_at inicial
\echo 'ANTES del UPDATE:'
SELECT 
    id, 
    name, 
    updated_at,
    NOW() as current_time
FROM users 
WHERE email = 'trigger_test@example.com';

-- Esperar 2 segundos (para que el timestamp cambie visiblemente)
SELECT pg_sleep(2);

-- Hacer UPDATE (el trigger debería actualizar updated_at)
UPDATE users 
SET name = 'Trigger Test User UPDATED' 
WHERE email = 'trigger_test@example.com';

-- Ver el updated_at después
\echo ''
\echo 'DESPUÉS del UPDATE:'
SELECT 
    id, 
    name, 
    updated_at,
    NOW() as current_time,
    EXTRACT(EPOCH FROM (NOW() - updated_at)) as seconds_ago
FROM users 
WHERE email = 'trigger_test@example.com';

\echo ''
\echo 'Verificar manualmente:'
\echo '  1. updated_at cambió? (debería ser ~2 segundos más reciente)'
\echo '  2. seconds_ago debería ser cercano a 0 (menos de 1 segundo)'
\echo '  3. Si updated_at NO cambió → ✗ Trigger NO funciona'
\echo '  4. Si updated_at cambió → ✓ Trigger funciona correctamente'
\echo ''

-- ============================================================================
-- CLEANUP
-- ============================================================================
\echo 'CLEANUP: Eliminar usuario de prueba'
DELETE FROM users WHERE email = 'trigger_test@example.com';

\echo ''
\echo '════════════════════════════════════════════════════════════'
\echo 'TEST COMPLETADO'
\echo '════════════════════════════════════════════════════════════'
