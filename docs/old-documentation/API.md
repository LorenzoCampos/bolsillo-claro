# 📚 Bolsillo Claro - API Documentation

**Base URL:** `https://api.fakerbostero.online/bolsillo/api`

**Versión:** 1.0  
**Última actualización:** 2026-01-15

---

## 📋 Tabla de Contenidos

- [Quick Reference](#-quick-reference)
- [Autenticación](#-autenticación)
- [Cuentas](#-cuentas)
- [Gastos](#-gastos)
- [Ingresos](#-ingresos)
- [Dashboard](#-dashboard)
- [Categorías](#️-categorías)
- [Metas de Ahorro](#-metas-de-ahorro)
- [Errores Comunes](#-errores-comunes)

---

## ⚡ Quick Reference

### Endpoints sin autenticación
```
POST   /api/auth/register
POST   /api/auth/login
GET    /api/health
```

### Endpoints con JWT solamente
```
GET    /api/accounts
POST   /api/accounts
GET    /api/accounts/:id
PUT    /api/accounts/:id
DELETE /api/accounts/:id
```

### Endpoints con JWT + X-Account-ID
```
# Gastos
GET    /api/expenses
POST   /api/expenses
GET    /api/expenses/:id
PUT    /api/expenses/:id
DELETE /api/expenses/:id

# Ingresos
GET    /api/incomes
POST   /api/incomes
GET    /api/incomes/:id
PUT    /api/incomes/:id
DELETE /api/incomes/:id

# Dashboard
GET    /api/dashboard/summary?month=YYYY-MM

# Categorías
GET    /api/expense-categories
POST   /api/expense-categories
PUT    /api/expense-categories/:id
DELETE /api/expense-categories/:id
GET    /api/income-categories
POST   /api/income-categories
PUT    /api/income-categories/:id
DELETE /api/income-categories/:id

# Metas de Ahorro
GET    /api/savings-goals
POST   /api/savings-goals
GET    /api/savings-goals/:id
PUT    /api/savings-goals/:id
DELETE /api/savings-goals/:id
POST   /api/savings-goals/:id/add-funds
POST   /api/savings-goals/:id/withdraw-funds
```

### Monedas soportadas
```
ARS - Peso argentino
USD - Dólar estadounidense
EUR - Euro (solo en incomes actualmente)
```

### Account Types
```
personal - Cuenta personal (sin miembros)
family   - Cuenta familiar (requiere al menos 1 miembro)
```

### Income Types
```
one-time  - Ingreso único
recurring - Ingreso recurrente (puede tener end_date)
```

---

## 🔐 Autenticación

### Headers Requeridos

**Para endpoints protegidos:**
```
Authorization: Bearer <access_token>
```

**Para endpoints que requieren cuenta:**
```
Authorization: Bearer <access_token>
X-Account-ID: <account_uuid>
```

---

## POST /auth/register
Registrar un nuevo usuario y auto-login (devuelve tokens JWT).

**Auth:** No requerido

**Request Body:**
```json
{
  "email": "string (requerido, formato email válido)",
  "password": "string (requerido, mínimo 8 caracteres)",
  "name": "string (requerido)"
}
```

**Success Response (201):**
```json
{
  "access_token": "jwt_string",
  "refresh_token": "jwt_string",
  "user": {
    "id": "uuid",
    "email": "string",
    "name": "string"
  }
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `409`: `{ "error": "El email ya está registrado" }`
- `500`: `{ "error": "Error creando usuario", "details": "..." }`

**Edge Cases:**
- Email se normaliza a minúsculas automáticamente
- Email duplicado retorna 409 Conflict
- Password debe tener al menos 8 caracteres
- **Auto-login:** El registro devuelve tokens JWT, el usuario queda logueado automáticamente

---

## POST /auth/login
Iniciar sesión y obtener tokens JWT.

**Auth:** No requerido

**Request Body:**
```json
{
  "email": "string (requerido, formato email)",
  "password": "string (requerido)"
}
```

**Success Response (200):**
```json
{
  "access_token": "jwt_string",
  "refresh_token": "jwt_string",
  "user": {
    "id": "uuid",
    "email": "string",
    "name": "string"
  }
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `401`: `{ "error": "Credenciales inválidas" }`
- `500`: `{ "error": "Error generando token" }`

**Edge Cases:**
- Email se normaliza a minúsculas
- No se revela si el email existe o la contraseña es incorrecta (seguridad)
- Access token expira en 15 minutos
- Refresh token expira en 7 días

---

## 💰 Cuentas

### POST /accounts
Crear una nueva cuenta (personal o familiar).

**Auth:** Requerido (JWT)

**Request Body:**
```json
{
  "name": "string (requerido, 1-100 chars)",
  "type": "string (requerido, 'personal' | 'family')",
  "currency": "string (requerido, 'ARS' | 'USD')",
  "initial_balance": "number (requerido)",
  "members": [
    {
      "name": "string (requerido si type='family')",
      "email": "string (opcional)"
    }
  ]
}
```

**Success Response (201):**
```json
{
  "message": "Cuenta creada exitosamente",
  "account": {
    "id": "uuid",
    "user_id": "uuid",
    "name": "string",
    "type": "string",
    "currency": "string",
    "initial_balance": 0,
    "current_balance": 0,
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "members": [
      {
        "id": "uuid",
        "name": "string",
        "email": "string"
      }
    ]
  }
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `400`: `{ "error": "Las cuentas familiares deben tener al menos un miembro" }`
- `400`: `{ "error": "Las cuentas personales no pueden tener miembros" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error creando cuenta", "details": "..." }`

**Edge Cases:**
- Si `type='family'`, el array `members` debe tener al menos 1 elemento
- Si `type='personal'`, el array `members` debe estar vacío o no enviarse
- Se crea automáticamente una meta de "Ahorro General" al crear la cuenta
- `initial_balance` siempre se crea en 0, este campo existe pero no se usa actualmente

**IMPORTANTE:**
- El campo `type` es **OBLIGATORIO** y debe ser exactamente `"personal"` o `"family"`
- El campo `currency` solo acepta `"ARS"` o `"USD"` (más monedas pueden agregarse)

---

## GET /accounts
Listar todas las cuentas del usuario autenticado.

**Auth:** Requerido (JWT)

**Query Parameters:** Ninguno

**Success Response (200):**
```json
{
  "accounts": [
    {
      "id": "uuid",
      "name": "string",
      "type": "string",
      "currency": "string",
      "createdAt": "timestamp",
      "memberCount": 3
    }
  ],
  "count": 1
}
```

**Error Responses:**
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error obteniendo cuentas", "details": "..." }`

**Edge Cases:**
- `memberCount` solo aparece si `type='family'`
- Si el usuario no tiene cuentas, retorna array vacío `{ "accounts": [], "count": 0 }`

---

## GET /accounts/:id
Obtener detalle de una cuenta específica.

**Auth:** Requerido (JWT)

**URL Parameters:**
- `id`: UUID de la cuenta

**Success Response (200):**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "name": "string",
  "type": "string",
  "currency": "string",
  "initial_balance": 0,
  "current_balance": 0,
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "members": [
    {
      "id": "uuid",
      "name": "string",
      "email": "string",
      "is_active": true,
      "created_at": "timestamp"
    }
  ]
}
```

**Error Responses:**
- `401`: `{ "error": "Usuario no autenticado" }`
- `403`: `{ "error": "No tenés permiso para ver esta cuenta" }`
- `404`: `{ "error": "Cuenta no encontrada" }`
- `500`: `{ "error": "Error obteniendo cuenta", "details": "..." }`

**Edge Cases:**
- `members` array solo aparece si `type='family'`
- Solo se retornan miembros con `is_active=true`

---

## PUT /accounts/:id
Actualizar una cuenta existente.

**Auth:** Requerido (JWT)

**URL Parameters:**
- `id`: UUID de la cuenta

**Request Body:**
```json
{
  "name": "string (opcional)",
  "currency": "string (opcional, 'ARS' | 'USD')"
}
```

**Success Response (200):**
```json
{
  "message": "Cuenta actualizada exitosamente",
  "account": {
    "id": "uuid",
    "name": "string",
    "type": "string",
    "currency": "string",
    "updated_at": "timestamp"
  }
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `403`: `{ "error": "No tenés permiso para actualizar esta cuenta" }`
- `404`: `{ "error": "Cuenta no encontrada" }`
- `500`: `{ "error": "Error actualizando cuenta", "details": "..." }`

**Edge Cases:**
- No se puede cambiar el `type` de la cuenta
- Todos los campos son opcionales, solo se actualizan los enviados

---

## DELETE /accounts/:id
Eliminar una cuenta.

**Auth:** Requerido (JWT)

**URL Parameters:**
- `id`: UUID de la cuenta

**Success Response (200):**
```json
{
  "message": "Cuenta eliminada exitosamente"
}
```

**Error Responses:**
- `401`: `{ "error": "Usuario no autenticado" }`
- `403`: `{ "error": "No tenés permiso para eliminar esta cuenta" }`
- `404`: `{ "error": "Cuenta no encontrada" }`
- `500`: `{ "error": "Error eliminando cuenta", "details": "..." }`

**Edge Cases:**
- Al eliminar una cuenta, se eliminan en cascada:
  - Todos los gastos asociados
  - Todos los ingresos asociados
  - Todas las metas de ahorro asociadas
  - Todos los miembros familiares (si type='family')

---

## 💸 Gastos

### POST /expenses
Crear un nuevo gasto.

**Auth:** Requerido (JWT + X-Account-ID)

**Headers:**
```
Authorization: Bearer <token>
X-Account-ID: <account_uuid>
```

**Request Body:**
```json
{
  "category_id": "uuid (opcional)",
  "amount": "number (requerido, positivo)",
  "currency": "string (requerido, 'ARS' | 'USD' | 'EUR')",
  "expense_type": "string (requerido, 'one-time' | 'recurring')",
  "description": "string (requerido, 1-500 chars)",
  "date": "string (requerido, formato: YYYY-MM-DD)",
  "end_date": "string (opcional, formato: YYYY-MM-DD)",
  
  // 🔄 CAMPOS DE RECURRENCIA (solo si expense_type='recurring')
  "recurrence_frequency": "string (opcional, 'daily' | 'weekly' | 'monthly' | 'yearly')",
  "recurrence_interval": "number (opcional, default: 1, cada cuántos períodos)",
  "recurrence_day_of_month": "number (opcional, 1-31, requerido si frequency='monthly'|'yearly')",
  "recurrence_day_of_week": "number (opcional, 0-6, requerido si frequency='weekly', 0=Domingo)",
  "total_occurrences": "number (opcional, cantidad total de repeticiones, NULL=infinito)",
  "current_occurrence": "number (opcional, default: 1, número de cuota actual)"
}
```

**Ejemplos de Request:**

*Gasto único:*
```json
{
  "category_id": "uuid-alimentacion",
  "amount": 1500,
  "currency": "ARS",
  "expense_type": "one-time",
  "description": "Almuerzo",
  "date": "2026-01-16"
}
```

*Alquiler mensual (sin fin):*
```json
{
  "category_id": "uuid-hogar",
  "amount": 80000,
  "currency": "ARS",
  "expense_type": "recurring",
  "description": "Alquiler Depto Palermo",
  "date": "2026-02-05",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 5,
  "total_occurrences": null
}
```

*Compra en 6 cuotas:*
```json
{
  "category_id": "uuid-ropa",
  "amount": 8000,
  "currency": "ARS",
  "expense_type": "recurring",
  "description": "Zapatillas Nike - Cuota 1/6",
  "date": "2026-01-16",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 16,
  "total_occurrences": 6,
  "current_occurrence": 1
}
```

**Success Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "category_id": "uuid",
  "category_name": "Alimentación",
  "description": "string",
  "amount": 100.50,
  "currency": "USD",
  "exchange_rate": 1,
  "amount_in_primary_currency": 100.50,
  "expense_type": "recurring",
  "date": "YYYY-MM-DD",
  "end_date": "YYYY-MM-DD",
  
  // 🔄 CAMPOS DE RECURRENCIA
  "recurrence_frequency": "monthly",
  "recurrence_interval": 1,
  "recurrence_day_of_month": 5,
  "recurrence_day_of_week": null,
  "total_occurrences": null,
  "current_occurrence": 1,
  "parent_expense_id": null,
  
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `400`: `{ "error": "account_id not found in context" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `404`: `{ "error": "Category not found" }`
- `500`: `{ "error": "Error creando gasto", "details": "..." }`

**Edge Cases:**
- El header `X-Account-ID` es **OBLIGATORIO**
- `category_id` es **OPCIONAL** (se puede crear gasto sin categoría)
- `amount` debe ser positivo
- `amount_in_primary_currency` se calcula automáticamente según la moneda de la cuenta
- Si la cuenta usa USD y el gasto es en ARS, se convierte automáticamente

**Validaciones de Recurrencia:**
- Si `expense_type='recurring'`:
  - `recurrence_frequency` es **REQUERIDO**
  - Si `frequency='monthly'` o `'yearly'` → `recurrence_day_of_month` es **REQUERIDO** (1-31)
  - Si `frequency='weekly'` → `recurrence_day_of_week` es **REQUERIDO** (0=Domingo, 6=Sábado)
  - `recurrence_interval` default = 1 (cada 1 período)
  - `current_occurrence` default = 1
  - Si `total_occurrences` está definido → `end_date` se calcula automáticamente

- Si `expense_type='one-time'`:
  - Todos los campos de recurrencia deben ser `null` o no enviarse
  - `end_date` debe ser `null`

---

## GET /expenses
Listar gastos con filtros opcionales.

**Auth:** Requerido (JWT + X-Account-ID)

**Query Parameters:**
- `month`: string (opcional, formato: YYYY-MM) - Filtra por mes
- `category_id`: uuid (opcional) - Filtra por categoría

**Success Response (200):**
```json
{
  "expenses": [
    {
      "id": "uuid",
      "account_id": "uuid",
      "category_id": "uuid",
      "category_name": "string",
      "amount": 100.50,
      "currency": "USD",
      "amount_in_primary_currency": 100.50,
      "description": "string",
      "date": "YYYY-MM-DD",
      "created_at": "timestamp"
    }
  ],
  "count": 1
}
```

**Error Responses:**
- `400`: `{ "error": "account_id not found in context" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error obteniendo gastos", "details": "..." }`

**Edge Cases:**
- Sin filtros, retorna todos los gastos de la cuenta
- Los gastos se ordenan por fecha descendente (más recientes primero)

---

## 📊 Dashboard

### GET /dashboard/summary
Obtener resumen financiero del mes.

**Auth:** Requerido (JWT + X-Account-ID)

**Query Parameters:**
- `month`: string (opcional, formato: YYYY-MM, default: mes actual)

**Success Response (200):**
```json
{
  "period": "2026-01",
  "primary_currency": "USD",
  "total_income": 5000.00,
  "total_expenses": 3200.50,
  "total_assigned_to_goals": 500.00,
  "available_balance": 1299.50,
  "expenses_by_category": [
    {
      "category_id": "uuid",
      "category_name": "Comida",
      "category_icon": "🍔",
      "category_color": "#FF5733",
      "total": 1200.00,
      "percentage": 37.5
    }
  ],
  "top_expenses": [
    {
      "id": "uuid",
      "description": "Supermercado",
      "amount": 250.00,
      "currency": "USD",
      "amount_in_primary_currency": 250.00,
      "category_name": "Comida",
      "date": "2026-01-10"
    }
  ],
  "recent_transactions": [
    {
      "id": "uuid",
      "type": "expense",
      "description": "Supermercado",
      "amount": 250.00,
      "currency": "USD",
      "amount_in_primary_currency": 250.00,
      "category_name": "Comida",
      "date": "2026-01-10",
      "created_at": "timestamp"
    }
  ]
}
```

**Error Responses:**
- `400`: `{ "error": "account_id not found in context" }`
- `400`: `{ "error": "invalid month format, use YYYY-MM" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "failed to get account currency" }`

**Edge Cases:**
- Si no se especifica `month`, usa el mes actual
- `available_balance` = `total_income` - `total_expenses` - `total_assigned_to_goals`
- `top_expenses` retorna máximo 5 gastos más grandes
- `recent_transactions` retorna máximo 10 transacciones (expenses + incomes mezclados)
- Todos los montos se convierten a la moneda primaria de la cuenta

---

## 💰 Ingresos

### POST /incomes
Crear un nuevo ingreso.

**Auth:** Requerido (JWT + X-Account-ID)

**Headers:**
```
Authorization: Bearer <token>
X-Account-ID: <account_uuid>
```

**Request Body:**
```json
{
  "family_member_id": "uuid (opcional, solo para cuentas family)",
  "category_id": "uuid (opcional)",
  "description": "string (requerido, 1-500 chars)",
  "amount": "number (requerido, positivo)",
  "currency": "string (requerido, 'ARS' | 'USD' | 'EUR')",
  "income_type": "string (requerido, 'one-time' | 'recurring')",
  "date": "string (requerido, formato: YYYY-MM-DD)",
  "end_date": "string (opcional, formato: YYYY-MM-DD, solo si income_type='recurring')",
  "exchange_rate": "number (opcional)",
  "amount_in_primary_currency": "number (opcional)"
}
```

**Success Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "family_member_id": "uuid",
  "category_id": "uuid",
  "category_name": "string",
  "description": "string",
  "amount": 5000.00,
  "currency": "USD",
  "exchange_rate": 1.0,
  "amount_in_primary_currency": 5000.00,
  "income_type": "one-time",
  "date": "2026-01-15",
  "end_date": null,
  "created_at": "timestamp"
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `400`: `{ "error": "account_id not found in context" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error creando ingreso", "details": "..." }`

**Edge Cases:**
- Si `income_type='recurring'`, `end_date` es opcional
- `amount_in_primary_currency` se calcula automáticamente si no se provee
- Si se provee `amount_in_primary_currency`, se usa ese valor (útil para ingresos reales con comisiones)

---

## GET /incomes
Listar ingresos con filtros opcionales.

**Auth:** Requerido (JWT + X-Account-ID)

**Query Parameters:**
- `month`: string (opcional, formato: YYYY-MM) - Filtra por mes
- `category_id`: uuid (opcional) - Filtra por categoría
- `income_type`: string (opcional, 'one-time' | 'recurring')

**Success Response (200):**
```json
{
  "incomes": [
    {
      "id": "uuid",
      "account_id": "uuid",
      "category_id": "uuid",
      "category_name": "Salario",
      "description": "string",
      "amount": 5000.00,
      "currency": "USD",
      "amount_in_primary_currency": 5000.00,
      "income_type": "recurring",
      "date": "2026-01-01",
      "created_at": "timestamp"
    }
  ],
  "count": 1
}
```

**Error Responses:**
- `400`: `{ "error": "account_id not found in context" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error obteniendo ingresos", "details": "..." }`

---

## GET /incomes/:id
Obtener detalle de un ingreso específico.

**Auth:** Requerido (JWT + X-Account-ID)

**URL Parameters:**
- `id`: UUID del ingreso

**Success Response (200):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "family_member_id": "uuid",
  "category_id": "uuid",
  "category_name": "string",
  "description": "string",
  "amount": 5000.00,
  "currency": "USD",
  "exchange_rate": 1.0,
  "amount_in_primary_currency": 5000.00,
  "income_type": "one-time",
  "date": "2026-01-15",
  "end_date": null,
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Error Responses:**
- `400`: `{ "error": "account_id not found in context" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `404`: `{ "error": "Income not found" }`
- `500`: `{ "error": "Error obteniendo ingreso", "details": "..." }`

---

## PUT /incomes/:id
Actualizar un ingreso existente.

**Auth:** Requerido (JWT + X-Account-ID)

**URL Parameters:**
- `id`: UUID del ingreso

**Request Body:** (todos los campos opcionales)
```json
{
  "category_id": "uuid (opcional)",
  "description": "string (opcional)",
  "amount": "number (opcional, positivo)",
  "currency": "string (opcional)",
  "date": "string (opcional)"
}
```

**Success Response (200):**
```json
{
  "message": "Ingreso actualizado exitosamente",
  "income": { /* IncomeResponse */ }
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `404`: `{ "error": "Income not found" }`
- `500`: `{ "error": "Error actualizando ingreso", "details": "..." }`

---

## DELETE /incomes/:id
Eliminar un ingreso.

**Auth:** Requerido (JWT + X-Account-ID)

**URL Parameters:**
- `id`: UUID del ingreso

**Success Response (200):**
```json
{
  "message": "Ingreso eliminado exitosamente"
}
```

**Error Responses:**
- `401`: `{ "error": "Usuario no autenticado" }`
- `404`: `{ "error": "Income not found" }`
- `500`: `{ "error": "Error eliminando ingreso", "details": "..." }`

---

## 🎯 Metas de Ahorro

### POST /savings-goals
Crear una nueva meta de ahorro.

**Auth:** Requerido (JWT + X-Account-ID)

**Request Body:**
```json
{
  "name": "string (requerido, 1-100 chars)",
  "target_amount": "number (requerido, positivo)",
  "currency": "string (requerido, 'ARS' | 'USD')",
  "deadline": "string (opcional, formato: YYYY-MM-DD)",
  "description": "string (opcional, max 500 chars)"
}
```

**Success Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "name": "string",
  "target_amount": 10000.00,
  "current_amount": 0.00,
  "currency": "USD",
  "deadline": "2026-12-31",
  "description": "string",
  "is_general": false,
  "is_active": true,
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error creando meta", "details": "..." }`

**Edge Cases:**
- `current_amount` siempre empieza en 0
- `is_general` se crea automáticamente en false (solo la cuenta tiene una meta con is_general=true)
- `deadline` es opcional

---

## GET /savings-goals
Listar metas de ahorro de la cuenta.

**Auth:** Requerido (JWT + X-Account-ID)

**Query Parameters:**
- `is_active`: boolean (opcional, default: true) - Filtra por activas/inactivas

**Success Response (200):**
```json
{
  "goals": [
    {
      "id": "uuid",
      "name": "Vacaciones",
      "target_amount": 10000.00,
      "current_amount": 2500.00,
      "currency": "USD",
      "deadline": "2026-12-31",
      "progress_percentage": 25.0,
      "is_general": false,
      "is_active": true,
      "created_at": "timestamp"
    }
  ],
  "count": 1
}
```

---

## GET /savings-goals/:id
Obtener detalle de una meta con historial de transacciones.

**Auth:** Requerido (JWT + X-Account-ID)

**URL Parameters:**
- `id`: UUID de la meta

**Success Response (200):**
```json
{
  "id": "uuid",
  "name": "Vacaciones",
  "target_amount": 10000.00,
  "current_amount": 2500.00,
  "currency": "USD",
  "deadline": "2026-12-31",
  "description": "string",
  "progress_percentage": 25.0,
  "is_general": false,
  "is_active": true,
  "created_at": "timestamp",
  "transactions": [
    {
      "id": "uuid",
      "type": "add",
      "amount": 500.00,
      "description": "Primer ahorro",
      "created_at": "timestamp"
    }
  ]
}
```

**Error Responses:**
- `401`: `{ "error": "Usuario no autenticado" }`
- `404`: `{ "error": "Savings goal not found" }`
- `500`: `{ "error": "Error obteniendo meta", "details": "..." }`

---

## POST /savings-goals/:id/add-funds
Agregar fondos a una meta de ahorro.

**Auth:** Requerido (JWT + X-Account-ID)

**URL Parameters:**
- `id`: UUID de la meta

**Request Body:**
```json
{
  "amount": "number (requerido, positivo)",
  "description": "string (opcional, max 500 chars)"
}
```

**Success Response (200):**
```json
{
  "message": "Fondos agregados exitosamente",
  "transaction": {
    "id": "uuid",
    "savings_goal_id": "uuid",
    "type": "add",
    "amount": 500.00,
    "description": "string",
    "created_at": "timestamp"
  },
  "new_current_amount": 3000.00
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `404`: `{ "error": "Savings goal not found" }`
- `500`: `{ "error": "Error agregando fondos", "details": "..." }`

---

## POST /savings-goals/:id/withdraw-funds
Retirar fondos de una meta de ahorro.

**Auth:** Requerido (JWT + X-Account-ID)

**URL Parameters:**
- `id`: UUID de la meta

**Request Body:**
```json
{
  "amount": "number (requerido, positivo)",
  "description": "string (opcional, max 500 chars)"
}
```

**Success Response (200):**
```json
{
  "message": "Fondos retirados exitosamente",
  "transaction": {
    "id": "uuid",
    "savings_goal_id": "uuid",
    "type": "withdraw",
    "amount": 200.00,
    "description": "string",
    "created_at": "timestamp"
  },
  "new_current_amount": 2800.00
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `400`: `{ "error": "Insufficient funds in savings goal" }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `404`: `{ "error": "Savings goal not found" }`
- `500`: `{ "error": "Error retirando fondos", "details": "..." }`

**Edge Cases:**
- No se puede retirar más de `current_amount`
- `current_amount` se actualiza automáticamente

---

## 🏷️ Categorías

### GET /expense-categories
Listar categorías de gastos (predefinidas + custom de la cuenta).

**Auth:** Requerido (JWT + X-Account-ID)

**Success Response (200):**
```json
{
  "categories": [
    {
      "id": "uuid",
      "account_id": null,
      "name": "Comida",
      "icon": "🍔",
      "color": "#FF5733",
      "is_custom": false
    },
    {
      "id": "uuid",
      "account_id": "uuid",
      "name": "Mi Categoría",
      "icon": "🎯",
      "color": "#00FF00",
      "is_custom": true
    }
  ],
  "count": 2
}
```

**Error Responses:**
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error obteniendo categorías", "details": "..." }`

**Edge Cases:**
- Retorna categorías predefinidas (account_id=null) + categorías custom de la cuenta
- Las categorías predefinidas no se pueden editar/eliminar

---

### POST /expense-categories
Crear una categoría custom de gastos.

**Auth:** Requerido (JWT + X-Account-ID)

**Request Body:**
```json
{
  "name": "string (requerido, 1-50 chars)",
  "icon": "string (opcional, emoji)",
  "color": "string (opcional, hex color)"
}
```

**Success Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "name": "Mi Categoría",
  "icon": "🎯",
  "color": "#00FF00",
  "is_custom": true,
  "created_at": "timestamp"
}
```

**Error Responses:**
- `400`: `{ "error": "Datos inválidos", "details": "..." }`
- `401`: `{ "error": "Usuario no autenticado" }`
- `500`: `{ "error": "Error creando categoría", "details": "..." }`

---

### GET /income-categories
Listar categorías de ingresos (predefinidas + custom de la cuenta).

**Auth:** Requerido (JWT + X-Account-ID)

**Success Response (200):**
```json
{
  "categories": [
    {
      "id": "uuid",
      "account_id": null,
      "name": "Salario",
      "icon": "💼",
      "color": "#4CAF50",
      "is_custom": false
    }
  ],
  "count": 1
}
```

---

### POST /income-categories
Crear una categoría custom de ingresos.

**Auth:** Requerido (JWT + X-Account-ID)

**Request Body:**
```json
{
  "name": "string (requerido, 1-50 chars)",
  "icon": "string (opcional, emoji)",
  "color": "string (opcional, hex color)"
}
```

**Success Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "name": "Freelance",
  "icon": "💻",
  "color": "#2196F3",
  "is_custom": true,
  "created_at": "timestamp"
}
```

---

## ❌ Errores Comunes

### Formato de Respuestas de Error

Todos los errores siguen este formato:
```json
{
  "error": "Mensaje descriptivo del error",
  "details": "Información adicional (opcional)"
}
```

### Códigos HTTP

- `200`: Éxito
- `201`: Creado exitosamente
- `400`: Request inválido (datos mal formateados)
- `401`: No autenticado (falta o es inválido el token JWT)
- `403`: Prohibido (no tienes permisos para este recurso)
- `404`: No encontrado
- `409`: Conflicto (ej: email duplicado)
- `500`: Error interno del servidor

### Validaciones Comunes

**Campos requeridos:**
- Si falta un campo requerido → `400` con mensaje `"Datos inválidos"`

**Formatos:**
- Email inválido → `400` con details del formato esperado
- UUID inválido → `404` o `400`
- Fecha inválida → `400` con mensaje "invalid date format"

**Autenticación:**
- Sin token JWT → `401` `"Usuario no autenticado"`
- Token expirado → `401` `"Token expirado"`
- Sin X-Account-ID → `400` `"account_id not found in context"`

---

## 📝 Notas de Implementación

### Multi-Moneda (Modo 3)

El sistema implementa **Modo 3** de multi-moneda:
- Cada cuenta tiene una **moneda primaria** (definida en `accounts.currency`)
- Las transacciones se guardan en su moneda original (`expenses.currency`, `incomes.currency`)
- Se calcula y guarda automáticamente el monto convertido a la moneda primaria (`amount_in_primary_currency`)
- El dashboard siempre muestra totales en la moneda primaria de la cuenta

### Cuentas Familiares

- Las cuentas `type='family'` tienen una tabla asociada `family_members`
- Los miembros son solo informativos (nombre y email opcional)
- No hay sistema de autenticación para miembros (solo el owner de la cuenta puede ver/editar)

### Categorías

- Existen categorías predefinidas (seeds en migraciones)
- Los usuarios pueden crear categorías custom por cuenta
- Las categorías custom solo son visibles para esa cuenta específica

---

## 🚀 Ejemplos de Flujos Completos

### Flujo 1: Registro y Primera Cuenta

```bash
# 1. Registrarse
POST /api/auth/register
{
  "email": "juan@example.com",
  "password": "mipassword123",
  "name": "Juan Pérez"
}

# 2. Login
POST /api/auth/login
{
  "email": "juan@example.com",
  "password": "mipassword123"
}
# Respuesta: { access_token, refresh_token, user }

# 3. Crear cuenta personal
POST /api/accounts
Headers: Authorization: Bearer <access_token>
{
  "name": "Cuenta Principal",
  "type": "personal",
  "currency": "USD",
  "initial_balance": 0
}
# Respuesta: { account: { id: "uuid-de-la-cuenta", ... } }

# 4. IMPORTANTE: Guardar el account.id para usarlo en X-Account-ID
```

---

### Flujo 2: Registrar Gasto

```bash
# Prerequisito: Tener access_token y account_id

# 1. Obtener categorías disponibles
GET /api/expense-categories
Headers: 
  Authorization: Bearer <access_token>
  X-Account-ID: <account_uuid>
# Respuesta: { categories: [...] }

# 2. Crear gasto
POST /api/expenses
Headers:
  Authorization: Bearer <access_token>
  X-Account-ID: <account_uuid>
{
  "category_id": "uuid-de-categoria",
  "amount": 50.00,
  "currency": "USD",
  "description": "Almuerzo",
  "date": "2026-01-15"
}
```

---

### Flujo 3: Ver Dashboard del Mes

```bash
# Ver resumen financiero del mes actual
GET /api/dashboard/summary
Headers:
  Authorization: Bearer <access_token>
  X-Account-ID: <account_uuid>

# Ver resumen de un mes específico
GET /api/dashboard/summary?month=2025-12
Headers:
  Authorization: Bearer <access_token>
  X-Account-ID: <account_uuid>
```

---

### Flujo 4: Crear Cuenta Familiar

```bash
POST /api/accounts
Headers: Authorization: Bearer <access_token>
{
  "name": "Gastos Familiares",
  "type": "family",
  "currency": "ARS",
  "initial_balance": 0,
  "members": [
    {
      "name": "María Pérez",
      "email": "maria@example.com"
    },
    {
      "name": "Pedro Pérez",
      "email": "pedro@example.com"
    }
  ]
}
```

---

## 🎓 Mejores Prácticas

### Frontend Development

1. **Guardar el account_id activo en localStorage**
   ```typescript
   localStorage.setItem('activeAccountId', account.id);
   ```

2. **Agregar X-Account-ID automáticamente en Axios interceptor**
   ```typescript
   axios.interceptors.request.use(config => {
     const accountId = localStorage.getItem('activeAccountId');
     if (accountId) {
       config.headers['X-Account-ID'] = accountId;
     }
     return config;
   });
   ```

3. **Validar con Zod antes de enviar**
   ```typescript
   import { z } from 'zod';
   
   const CreateAccountSchema = z.object({
     name: z.string().min(1).max(100),
     type: z.enum(['personal', 'family']),
     currency: z.enum(['ARS', 'USD']),
     initial_balance: z.number(),
     members: z.array(z.object({
       name: z.string(),
       email: z.string().email().optional()
     })).optional()
   });
   
   // Validar antes de POST
   const data = CreateAccountSchema.parse(formData);
   ```

4. **Manejar errores de forma consistente**
   ```typescript
   try {
     await createAccount(data);
   } catch (error) {
     if (axios.isAxiosError(error)) {
       const apiError = error.response?.data;
       console.error(apiError.error, apiError.details);
     }
   }
   ```

---

## 📝 Notas Importantes para el Frontend

### Campo `type` en Accounts

**⚠️ CRÍTICO:** El campo `type` es **OBLIGATORIO** al crear cuentas.

```typescript
// ❌ MAL - Falta el campo type
{
  name: "Mi Cuenta",
  currency: "USD",
  initial_balance: 0
}

// ✅ BIEN
{
  name: "Mi Cuenta",
  type: "personal",  // ← REQUERIDO
  currency: "USD",
  initial_balance: 0
}
```

### Multi-Moneda (Modo 3)

El backend maneja automáticamente la conversión de monedas:
- Guarda el monto original en `amount` + `currency`
- Calcula y guarda `amount_in_primary_currency` automáticamente
- El dashboard siempre muestra totales en la moneda primaria de la cuenta

No necesitás calcular conversiones en el frontend, el backend lo hace.

### Cuenta Activa

El middleware `AccountMiddleware` requiere que el header `X-Account-ID` esté presente en todos los endpoints de:
- `/expenses`
- `/incomes`
- `/dashboard`
- `/expense-categories`
- `/income-categories`
- `/savings-goals`

**Sin este header, obtendrás error 400:** `"account_id not found in context"`

---

**Fin de la documentación** 🚀

---

**Creado el:** 2026-01-15  
**Última actualización:** 2026-01-15  
**Versión:** 1.0.0  
**Mantenido por:** Gentleman Programming & Lorenzo
