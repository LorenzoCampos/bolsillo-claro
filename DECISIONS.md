# Decisiones Técnicas - Bolsillo Claro

Este documento registra las decisiones arquitectónicas y técnicas importantes del proyecto, con sus justificaciones.

---

## 📐 Arquitectura General

### Backend: Go + Gin + PostgreSQL

**Decisión:** Usar Go como lenguaje backend con framework Gin y PostgreSQL como base de datos.

**Razones:**
- Go: Rendimiento, simplicidad, tipado fuerte, excelente para APIs REST
- Gin: Framework minimalista pero completo, buen ecosistema
- PostgreSQL: Base de datos robusta, soporte JSON, ENUMs, transacciones ACID

**Alternativas consideradas:**
- Node.js + Express: Más familiar pero menos performante
- Python + FastAPI: Excelente pero Go es más rápido en producción

---

## 🔐 Autenticación y Seguridad

### JWT con Access + Refresh Tokens

**Decisión:** Usar JWT (HS256) con dos tipos de tokens:
- Access token: 15 minutos de vida
- Refresh token: 7 días de vida

**Razones:**
- Access token corto reduce ventana de vulnerabilidad
- Refresh token largo mejora UX (no pide login cada 15 min)
- JWT es stateless (no requiere DB lookup en cada request)

**Implementación:**
```go
AccessTokenDuration:  15 * time.Minute
RefreshTokenDuration: 7 * 24 * time.Hour
```

### Bcrypt para Passwords

**Decisión:** Cost factor de 12 para bcrypt.

**Razones:**
- Balance entre seguridad y performance
- 12 es el recomendado por OWASP en 2024+
- ~250ms por hash (aceptable para login/register)

---

## 🏦 Modelo de Datos: Users vs Accounts

### Separación Users - Accounts

**Decisión:** Users y Accounts son entidades separadas con relación 1:N.

**Razones:**
1. **Múltiples contextos financieros:**
   - Usuario puede tener "Finanzas Personales", "Gastos Familia", "Negocio"
   - Cada contexto tiene sus propios gastos/ingresos/metas

2. **Cuentas familiares:**
   - Account tipo `family` tiene múltiples `family_members`
   - Permite asignar gastos a personas específicas
   - Juan (hijo) no necesita login, pero María (madre) puede ver "cuánto gasta Juan"

3. **Escalabilidad:**
   - Futuro: Compartir cuentas con otros users (co-ownership)
   - Ejemplo: Cuenta familiar compartida entre esposos (ambos users, misma account)

**Estructura:**
```
User (maria@example.com)
├── Account 1: "Finanzas Personales" (personal, ARS)
│   ├── Expenses
│   ├── Incomes
│   └── Savings Goals
├── Account 2: "Gastos Familia" (family, ARS)
│   ├── Family Members: María, Juan, Sofía
│   ├── Expenses (asignables a miembros)
│   └── Savings Goals compartidas
└── Account 3: "Mi Negocio" (personal, USD)
    └── Finanzas del emprendimiento
```

**Alternativa rechazada:**
- User = Account (1:1): No permite múltiples contextos

---

## 💰 Multi-Currency con Modo 3: Flexibilidad Total

### Decisión: Sistema de 3 Modos para Conversión de Moneda

**Problema Real (Argentina 2026):**
Usuario compra Claude Pro por USD 20 con tarjeta.
- Dólar oficial: $900
- **Dólar tarjeta (con impuestos):** $1,575 (30% imp. PAÍS + 45% percepción ganancias)
- **Monto real debitado:** ARS $31,500

Si guardamos solo "USD 20 a tasa 900", perdemos $13,500 de diferencia con la realidad.

**Solución: Modo 3 - Flexibilidad Total**

Al crear un gasto/ingreso en moneda extranjera, el usuario puede proveer:

#### **Modo 1: Moneda Local (Automático)**
```json
POST /api/expenses
{
  "description": "Supermercado",
  "amount": 15000,
  "currency": "ARS",
  "date": "2026-01-14"
}

// Backend calcula automáticamente:
// exchange_rate = 1.0
// amount_in_primary_currency = 15000
```

#### **Modo 2: Con Exchange Rate (Conversión Estándar)**
```json
POST /api/expenses
{
  "description": "Amazon Prime",
  "amount": 10,
  "currency": "USD",
  "exchange_rate": 900,
  "date": "2026-01-14"
}

// Backend calcula:
// amount_in_primary_currency = 10 × 900 = 9000
```

#### **Modo 3: Con Monto Real Pagado (Dólar Tarjeta) ⭐**
```json
POST /api/expenses
{
  "description": "Claude Pro - Enero 2026",
  "amount": 20,
  "currency": "USD",
  "amount_in_primary_currency": 31500,  ← Lo que REALMENTE pagaste
  "date": "2026-01-14"
}

// Backend calcula la tasa efectiva:
// exchange_rate = 31500 / 20 = 1575
```

**✅ Captura perfecta del dólar tarjeta argentino!**

### Estructura de Base de Datos

```sql
-- Campos agregados a expenses e incomes
ALTER TABLE expenses ADD COLUMN exchange_rate DECIMAL(15, 6) NOT NULL;
ALTER TABLE expenses ADD COLUMN amount_in_primary_currency DECIMAL(15, 2) NOT NULL;

ALTER TABLE incomes ADD COLUMN exchange_rate DECIMAL(15, 6) NOT NULL;
ALTER TABLE incomes ADD COLUMN amount_in_primary_currency DECIMAL(15, 2) NOT NULL;

-- Tabla de tasas de cambio (para fallback automático)
CREATE TABLE exchange_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_currency currency NOT NULL,
    to_currency currency NOT NULL,
    rate DECIMAL(15, 6) NOT NULL,
    rate_date DATE NOT NULL,
    source VARCHAR(100),  -- 'manual', 'bcra', etc.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(from_currency, to_currency, rate_date)
);
```

### Lógica del Backend (Pseudo-código)

```go
if currency == primaryCurrency {
    // Modo 1: Moneda local
    exchangeRate = 1.0
    amountInPrimaryCurrency = amount
    
} else if amountInPrimaryCurrency != nil {
    // Modo 3: Usuario proveyó monto real (PRIORIDAD)
    exchangeRate = amountInPrimaryCurrency / amount  // Calcula tasa efectiva
    
} else if exchangeRate != nil {
    // Modo 2: Usuario proveyó tasa
    amountInPrimaryCurrency = amount * exchangeRate
    
} else {
    // Fallback: Buscar en tabla exchange_rates
    rate = fetchRateFromDB(currency, primaryCurrency, date)
    if rate == nil {
        return ERROR "Debe proveer exchange_rate o amount_in_primary_currency"
    }
    exchangeRate = rate
    amountInPrimaryCurrency = amount * rate
}
```

### Ventajas del Modo 3

✅ **Realidad Argentina:** Captura dólar tarjeta, dólar MEP, CCL, cripto  
✅ **Flexibilidad Total:** Usuario elige qué dato tiene disponible  
✅ **Trazabilidad:** Siempre sabés cuánto pagaste REALMENTE  
✅ **Reportes Precisos:** Balance mensual exacto en moneda primaria  
✅ **Auditoría:** "Pagué USD 20 pero me cobraron ARS $31,500 (tasa efectiva 1575)"  

### Ejemplo de Datos Guardados

```sql
SELECT description, amount, currency, exchange_rate, amount_in_primary_currency
FROM expenses WHERE date = '2026-01-14';
```

| description | amount | currency | exchange_rate | amount_in_primary_currency |
|-------------|--------|----------|---------------|----------------------------|
| Supermercado | 15000 | ARS | 1.0 | 15000 |
| Amazon Prime | 10 | USD | 900.0 | 9000 |
| Claude Pro | 20 | USD | **1575.0** | **31500** |

**Total en ARS:** $55,500 (conversión exacta a la realidad del usuario)

---

## 📊 Exchange Rates: Semi-Automático

### Decisión: Carga Manual Diaria por Admin

**Opciones evaluadas:**

**A) Manual por usuario cada vez:**
- ❌ UX horrible (usuario mete tipo de cambio en cada gasto)

**B) API externa automática:**
- ❌ Dependencia de servicio externo
- ❌ ¿Qué dólar usar? (oficial, blue, MEP, CCL, cripto)
- ❌ Complejidad innecesaria

**C) Semi-automático (ELEGIDA):**
- ✅ Admin carga tipo de cambio 1 vez por día
- ✅ Backend usa ese valor para todos los gastos del día
- ✅ Si no existe, pide al usuario que lo ingrese manualmente
- ✅ Flexible y simple

**Implementación:**
```sql
CREATE TABLE exchange_rates (
    id UUID PRIMARY KEY,
    from_currency currency NOT NULL,
    to_currency currency NOT NULL,
    rate DECIMAL(15, 6) NOT NULL,
    date DATE NOT NULL,
    source VARCHAR(50) DEFAULT 'manual',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_currency, to_currency, date)
);

-- Ejemplo de datos
INSERT INTO exchange_rates (from_currency, to_currency, rate, date, source) VALUES
('USD', 'ARS', 900.00, '2026-01-13', 'manual'),
('USD', 'ARS', 905.50, '2026-01-14', 'manual'),
('EUR', 'ARS', 980.00, '2026-01-13', 'manual');
```

**Endpoint admin:**
```
POST /api/exchange-rates
{
  "from_currency": "USD",
  "to_currency": "ARS",
  "rate": 900,
  "date": "2026-01-13"
}
```

**Lógica en CREATE expense/income:**
```go
if req.Currency != account.PrimaryCurrency {
    // Buscar exchange rate del día
    rate, err := getExchangeRate(req.Currency, account.PrimaryCurrency, req.Date)
    
    if err != nil {
        // No existe, pedir al usuario
        return gin.H{
            "error": "exchange_rate_required",
            "message": "Please provide exchange rate for this date",
            "from": req.Currency,
            "to": account.PrimaryCurrency,
            "date": req.Date
        }
    }
    
    // Guardar snapshot
    expense.ExchangeRate = rate
    expense.AmountInPrimary = req.Amount * rate
}
```

---

## 💵 Savings Goals: Descuento Virtual (No Expenses Reales)

### Decisión: Metas NO crean expenses, solo descuentan del balance disponible

**Problema:**
Usuario asigna $30,000 a meta "Viaje a Brasil".
¿Cómo representar esto?

**Opción A: Crear expense real (RECHAZADA)**
```json
POST /api/expenses
{
  "description": "Ahorro para viaje",
  "category_id": "ahorro",
  "amount": 30000,
  "linked_to_goal_id": "uuid"
}
```

**Problemas:**
- ❌ Reportes de "gastos" incluyen "ahorros" (conceptualmente incorrecto)
- ❌ Si sacás plata de la meta, ¿creás un income negativo?
- ❌ Contamina la tabla expenses con datos que no son gastos reales

**Opción B: Descuento virtual (ELEGIDA)**

**Estructura:**
```sql
CREATE TABLE savings_goals (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    name TEXT NOT NULL,
    target_amount DECIMAL(15, 2) NOT NULL,
    current_amount DECIMAL(15, 2) DEFAULT 0,
    currency currency NOT NULL,
    deadline DATE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE savings_goal_transactions (
    id UUID PRIMARY KEY,
    savings_goal_id UUID REFERENCES savings_goals(id),
    amount DECIMAL(15, 2) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL, -- 'deposit' o 'withdrawal'
    description TEXT,
    date DATE NOT NULL,
    created_at TIMESTAMP
);
```

**Endpoints:**
```
POST /api/savings-goals/:id/add-funds
{
  "amount": 30000,
  "description": "Ahorro enero",
  "date": "2026-01-13"
}

POST /api/savings-goals/:id/withdraw-funds
{
  "amount": 5000,
  "description": "Adelanto para pasaje",
  "date": "2026-01-15"
}
```

**Dashboard calcula:**
```
Ingresos del mes: $150,000
Gastos del mes: $85,000
Asignado a metas: $30,000
─────────────────────────────
Balance disponible: $35,000
```

**Ventajas:**
- ✅ Separación conceptual clara
- ✅ Reportes de gastos no incluyen ahorros
- ✅ Fácil revertir (withdraw funds)
- ✅ Historial de movimientos en savings_goal_transactions

---

## 🏷️ Categorías: Predefinidas + Custom por Cuenta

### Decisión: Sistema híbrido (predefinidas + personalizables)

**Estructura:**
```sql
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY,
    account_id UUID REFERENCES accounts(id),  -- NULL = predefinida
    name TEXT NOT NULL,
    icon TEXT,
    color TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    UNIQUE (COALESCE(account_id::text, 'SYSTEM'), name)
);
```

**Lógica:**
- `account_id = NULL` y `is_system = TRUE` → Categoría predefinida (global)
- `account_id = <uuid>` y `is_system = FALSE` → Categoría custom (específica de cuenta)

**Query para listar:**
```sql
SELECT * FROM expense_categories 
WHERE account_id IS NULL OR account_id = $1
ORDER BY is_system DESC, name ASC
```

**Resultado:**
```
- Alimentación (system)
- Transporte (system)
- Veterinario (custom de esta cuenta)
- ... resto de system
```

**Ventajas:**
- ✅ Usuario nuevo tiene categorías listas para usar (onboarding fácil)
- ✅ Power users pueden crear categorías específicas ("Veterinario", "Clases de tango")
- ✅ Reportes consistentes (mayoría usa predefinidas)
- ✅ Escalable (cada cuenta personaliza independientemente)

**Reglas de negocio:**
- No se pueden editar/borrar categorías system
- No se pueden borrar categorías custom con expenses/incomes asociados
- Nombre único por scope (global para system, por cuenta para custom)

---

## 🔄 Expense/Income Types: One-time vs Recurring

### Decisión: end_date opcional para recurring

**Tipos:**
- `one-time`: Gasto único (compra supermercado)
- `recurring`: Gasto recurrente (Netflix, gimnasio)

**Regla de end_date:**
- `one-time`: **NO** puede tener end_date
- `recurring`: **PUEDE** tener end_date (opcional)

**Razón del cambio:**
Inicialmente `recurring` requería `end_date`, pero en la práctica:
- Netflix: No sabés cuándo lo vas a cancelar
- Spotify: Suscripción indefinida
- Gimnasio: Contrato de 6 meses SÍ tiene end_date

**Constraint DB:**
```sql
CONSTRAINT check_recurring_end_date CHECK (
    (expense_type = 'one-time' AND end_date IS NULL) OR
    (expense_type = 'recurring' AND (end_date IS NULL OR end_date >= date))
)
```

**Casos de uso:**
```json
// Netflix indefinido
{
  "description": "Netflix",
  "expense_type": "recurring",
  "date": "2026-01-01",
  "end_date": null
}

// Gimnasio 6 meses
{
  "description": "Gimnasio",
  "expense_type": "recurring",
  "date": "2026-01-01",
  "end_date": "2026-06-30"
}
```

---

## 🔒 Seguridad: AccountMiddleware

### Decisión: Header X-Account-ID para context switching

**Problema:**
Usuario tiene 3 cuentas. ¿Cómo sabe el backend en qué cuenta está operando?

**Solución:**
```
POST /api/expenses
Headers:
  Authorization: Bearer <jwt>        ← Identifica al USER
  X-Account-ID: <uuid-de-cuenta>     ← Identifica la ACCOUNT
```

**Middleware valida:**
1. JWT válido (AuthMiddleware)
2. Account existe y pertenece al user (AccountMiddleware)
3. Si pasa, guarda `account_id` en contexto

**Ventajas:**
- ✅ Explicit context (no asumes "cuenta por defecto")
- ✅ Frontend puede hacer "switch" de cuenta fácilmente
- ✅ Seguridad: No puedes operar en cuentas ajenas

**Alternativas rechazadas:**
- URL: `/api/accounts/:account_id/expenses` → URLs muy largas
- Default account: ¿Y si tiene múltiples? ¿Cuál es default?

---

## 📝 Nomenclatura API

### Decisión: REST con nombres en inglés

**Formato:**
```
GET    /api/expenses
POST   /api/expenses
GET    /api/expenses/:id
PUT    /api/expenses/:id
DELETE /api/expenses/:id
```

**Razones:**
- Nombres en inglés (estándar internacional)
- REST conventions (verbos HTTP + sustantivos plurales)
- Consistente con ecosistema Go/Gin

---

## 🎨 Frontend Considerations

### Onboarding de Primera Cuenta

**Backend provee:**
```
GET /api/users/me
{
  "id": "uuid",
  "email": "maria@example.com",
  "has_accounts": false,  ← Frontend detecta esto
  "created_at": "..."
}
```

**Frontend maneja:**
- Modal/wizard si `has_accounts === false`
- Guía paso a paso para crear primera cuenta
- Selector de cuenta en navbar si tiene múltiples

**Separación de responsabilidades:**
- Backend: Lógica de negocio, validaciones, seguridad
- Frontend: UX, wizards, visualización

---

## 🚀 Filosofía de Desarrollo

### Incremental y Validado

**Enfoque:**
1. Implementar feature completa (CRUD completo)
2. Probar cada endpoint inmediatamente
3. No pasar a siguiente feature hasta que actual esté 100%

**Razón:**
Evitar "medio implementado". Mejor tener 5 features completas que 10 a medias.

### API-First, Frontend-Ready

**Decisión:**
- Responses incluyen info redundante para facilitar frontend
- Ejemplo: `category_id` + `category_name` (frontend no hace segundo request)

```json
{
  "id": "uuid",
  "category_id": "uuid-cat",
  "category_name": "Alimentación"  ← Redundante pero útil
}
```

**Razón:**
- Reduce requests HTTP
- Frontend más simple
- Mejor UX (respuesta más rápida)

---

## 📚 Referencias

- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [REST API Design Best Practices](https://restfulapi.net/)
- [PostgreSQL ENUMs](https://www.postgresql.org/docs/current/datatype-enum.html)

---

**Última actualización:** 2026-01-14 (Implementación Modo 3 Multi-Currency)
