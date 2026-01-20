#!/bin/bash
# Script de testing para mejoras de seguridad AUTH
# Ejecutar con: bash test_auth_improvements.sh

set -e  # Exit on error

echo "════════════════════════════════════════════════════════════"
echo "🧪 AUTH SECURITY IMPROVEMENTS - TEST SUITE"
echo "════════════════════════════════════════════════════════════"
echo ""

BASE_URL="http://localhost:8080/api"
FAILED=0
PASSED=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function pass() {
    PASSED=$((PASSED + 1))
    echo -e "${GREEN}✓${NC} $1"
}

function fail() {
    FAILED=$((FAILED + 1))
    echo -e "${RED}✗${NC} $1"
}

function info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# ============================================================================
# TEST 1: Verificar que el servidor está corriendo
# ============================================================================
echo "TEST 1: Verificar servidor"
if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" | grep -q "200"; then
    pass "Servidor responde en /api/health"
else
    fail "Servidor NO responde en /api/health (¿está corriendo?)"
    echo ""
    echo "Por favor iniciar el servidor primero:"
    echo "  cd backend && go run cmd/api/main.go"
    exit 1
fi
echo ""

# ============================================================================
# TEST 2: Rate Limiting en /auth/register
# ============================================================================
echo "TEST 2: Rate Limiting en /auth/register"
info "Haciendo 6 requests seguidos (límite: 5)..."

for i in {1..6}; do
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"ratelimit_test_$i@example.com\",\"password\":\"test1234\",\"name\":\"Test\"}")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ $i -le 5 ]; then
        # Primeros 5 deberían pasar (o fallar con 409 si ya existe, pero NO 429)
        if [ "$HTTP_CODE" != "429" ]; then
            echo -e "  Request $i: ${GREEN}✓${NC} (HTTP $HTTP_CODE)"
        else
            echo -e "  Request $i: ${RED}✗${NC} Rate limit activado antes de tiempo (HTTP $HTTP_CODE)"
            FAILED=$((FAILED + 1))
        fi
    else
        # El 6to DEBE ser 429
        if [ "$HTTP_CODE" == "429" ]; then
            pass "Request 6: Bloqueado por rate limit (HTTP 429)"
            
            # Verificar que incluye retry_after en el body
            BODY=$(echo "$RESPONSE" | head -n -1)
            if echo "$BODY" | grep -q "retry_after"; then
                pass "Response incluye campo 'retry_after'"
            else
                fail "Response NO incluye campo 'retry_after'"
            fi
        else
            fail "Request 6: NO fue bloqueado (HTTP $HTTP_CODE) - rate limit no funciona"
        fi
    fi
    
    sleep 0.2  # Pequeño delay para no saturar
done
echo ""

# ============================================================================
# TEST 3: Rate Limiting en /auth/login
# ============================================================================
echo "TEST 3: Rate Limiting en /auth/login"
info "Haciendo 6 requests seguidos con credenciales incorrectas..."

for i in {1..6}; do
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"nonexistent@example.com\",\"password\":\"wrongpassword\"}")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ $i -le 5 ]; then
        # Primeros 5 deberían ser 401 (unauthorized)
        if [ "$HTTP_CODE" == "401" ]; then
            echo -e "  Request $i: ${GREEN}✓${NC} (HTTP $HTTP_CODE - unauthorized)"
        else
            echo -e "  Request $i: ${YELLOW}?${NC} (HTTP $HTTP_CODE - esperaba 401)"
        fi
    else
        # El 6to DEBE ser 429
        if [ "$HTTP_CODE" == "429" ]; then
            pass "Request 6: Bloqueado por rate limit (HTTP 429)"
        else
            fail "Request 6: NO fue bloqueado (HTTP $HTTP_CODE)"
        fi
    fi
    
    sleep 0.2
done
echo ""

# ============================================================================
# TEST 4: Verificar logging (manual - revisar stdout del server)
# ============================================================================
echo "TEST 4: Verificar logging estructurado"
info "Este test es MANUAL - verificar los logs del servidor"
echo ""
echo "  Los logs deberían mostrar (en stdout del server):"
echo "  1. Eventos 'auth.register.failed' (email_already_exists)"
echo "  2. Eventos 'auth.login.failed' (user_not_found / invalid_password)"
echo "  3. Eventos 'ratelimit.exceeded' en los requests 6"
echo "  4. Formato JSON estructurado con timestamp, level, event, data"
echo ""
info "Revisar manualmente los logs del servidor y confirmar"
echo ""

# ============================================================================
# TEST 5: Trigger updated_at (requiere DB access)
# ============================================================================
echo "TEST 5: Trigger updated_at"
info "Este test requiere acceso a PostgreSQL"
echo ""
echo "  Para testearlo manualmente:"
echo "  1. psql -U postgres -d bolsillo_claro"
echo "  2. SELECT id, name, updated_at FROM users LIMIT 1;"
echo "  3. UPDATE users SET name = 'Test Update' WHERE id = '<ese_id>';"
echo "  4. SELECT id, name, updated_at FROM users WHERE id = '<ese_id>';"
echo "  5. Verificar que updated_at cambió"
echo ""
info "Si querés, puedo generar el SQL para ejecutar"
echo ""

# ============================================================================
# TEST 6: Email normalization
# ============================================================================
echo "TEST 6: Email normalization (case-insensitive)"
info "Intentando registrar mismo email con diferentes cases..."

# Registrar con email en mayúsculas
RESPONSE1=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"NORMALIZATION@EXAMPLE.COM","password":"test1234","name":"Test1"}')

HTTP_CODE1=$(echo "$RESPONSE1" | tail -n1)

if [ "$HTTP_CODE1" == "201" ] || [ "$HTTP_CODE1" == "409" ]; then
    info "Primer registro: HTTP $HTTP_CODE1"
    
    # Intentar registrar con minúsculas (debería fallar con 409)
    RESPONSE2=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"email":"normalization@example.com","password":"test1234","name":"Test2"}')
    
    HTTP_CODE2=$(echo "$RESPONSE2" | tail -n1)
    
    if [ "$HTTP_CODE2" == "409" ]; then
        pass "Email normalization funciona (detecta duplicado case-insensitive)"
    else
        fail "Email normalization NO funciona (HTTP $HTTP_CODE2 - esperaba 409)"
    fi
else
    fail "No se pudo testear normalization (primer request falló con HTTP $HTTP_CODE1)"
fi
echo ""

# ============================================================================
# TEST 7: Headers de rate limit
# ============================================================================
echo "TEST 7: Headers informativos de rate limit"
info "Verificando headers X-RateLimit-Limit y Retry-After..."

# Hacer suficientes requests para activar rate limit
for i in {1..6}; do
    curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"headers_test@example.com","password":"wrong"}' \
        > /dev/null 2>&1
    sleep 0.1
done

# El siguiente debería tener los headers
HEADERS=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"headers_test@example.com","password":"wrong"}' 2>&1)

if echo "$HEADERS" | grep -qi "X-RateLimit-Limit"; then
    pass "Header X-RateLimit-Limit presente"
else
    fail "Header X-RateLimit-Limit NO encontrado"
fi

if echo "$HEADERS" | grep -qi "Retry-After"; then
    pass "Header Retry-After presente"
else
    fail "Header Retry-After NO encontrado"
fi
echo ""

# ============================================================================
# RESUMEN
# ============================================================================
echo "════════════════════════════════════════════════════════════"
echo "RESUMEN DE TESTS"
echo "════════════════════════════════════════════════════════════"
echo -e "${GREEN}Pasados:${NC} $PASSED"
echo -e "${RED}Fallados:${NC} $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ TODOS LOS TESTS PASARON${NC}"
    exit 0
else
    echo -e "${RED}✗ ALGUNOS TESTS FALLARON${NC}"
    echo ""
    echo "Revisar los errores arriba y fixear antes de mergear"
    exit 1
fi
