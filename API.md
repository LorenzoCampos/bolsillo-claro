# 📚 Bolsillo Claro - API Documentation

**Base URL:** `https://api.fakerbostero.online/bolsillo/api`  
**Versión:** 2.0  
**Última actualización:** 2026-01-16

---

## 📋 Quick Reference

### Authentication

```bash
# Without auth
POST   /auth/register
POST   /auth/login
POST   /auth/refresh

# With JWT only
GET    /accounts
POST   /accounts
GET    /accounts/:id
PUT    /accounts/:id
DELETE /accounts/:id

# With JWT + X-Account-ID header
GET    /expenses
POST   /expenses
GET    /expenses/:id
PUT    /expenses/:id
DELETE /expenses/:id

GET    /incomes
POST   /incomes
GET    /incomes/:id
PUT    /incomes/:id
DELETE /incomes/:id

GET    /dashboard/summary
GET    /expense-categories
POST   /expense-categories
GET    /income-categories
POST   /income-categories

GET    /savings-goals
POST   /savings-goals
GET    /savings-goals/:id
PUT    /savings-goals/:id
DELETE /savings-goals/:id
POST   /savings-goals/:id/add-funds
POST   /savings-goals/:id/withdraw-funds

GET    /recurring-expenses
POST   /recurring-expenses
GET    /recurring-expenses/:id
PUT    /recurring-expenses/:id
DELETE /recurring-expenses/:id

GET    /recurring-incomes
POST   /recurring-incomes
GET    /recurring-incomes/:id
PUT    /recurring-incomes/:id
DELETE /recurring-incomes/:id
```

### Headers

**JWT only:**
```
Authorization: Bearer <access_token>
```

**JWT + Account:**
```
Authorization: Bearer <access_token>
X-Account-ID: <account_uuid>
```

### Supported Currencies

```
ARS - Peso argentino
USD - Dólar estadounidense
```

**Note:** EUR was removed in version 1.1.0 as it's not in the database ENUM.

### Account Types

```
personal - Cuenta personal (sin miembros)
family   - Cuenta familiar (requiere ≥1 miembro)
```

---

## 🔐 Authentication

### POST /auth/register

Registrar nuevo usuario (auto-login, devuelve tokens).

**Request:**
```json
{
  "email": "user@example.com",
  "password": "min8chars",
  "name": "Juan Pérez"
}
```

**Validaciones:**
- Email: formato válido, se normaliza a minúsculas automáticamente
- Password: mínimo 8 caracteres
- Name: requerido, no vacío

**Response (201):**
```json
{
  "access_token": "jwt_string",
  "refresh_token": "jwt_string",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Juan Pérez"
  }
}
```

**Errors:**
- `400` - Datos inválidos
- `409` - Email ya registrado
- `429` - Demasiados intentos (rate limit: 5 requests cada 15 minutos)

---

### POST /auth/login

Iniciar sesión.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password"
}
```

**Nota:** El email se normaliza a minúsculas automáticamente (case-insensitive).

**Response (200):** Igual a register

**Errors:**
- `401` - Credenciales inválidas (no revela si email existe o no)
- `429` - Demasiados intentos (rate limit: 5 requests cada 15 minutos)

**Tokens:**
- Access: 15min
- Refresh: 7 días

---

### POST /auth/refresh

Renovar tokens usando el refresh token (evita re-login).

**Request:**
```json
{
  "refresh_token": "jwt_refresh_token_string"
}
```

**Response (200):**
```json
{
  "access_token": "new_jwt_access_token",
  "refresh_token": "new_jwt_refresh_token"
}
```

**Notas:**
- El refresh token viejo queda invalidado (rotación automática)
- Siempre devuelve un PAR nuevo de tokens (access + refresh)
- Verifica que el usuario siga existiendo en la DB antes de renovar

**Errors:**
- `400` - Datos inválidos (refresh_token requerido)
- `401` - Refresh token inválido, expirado o usuario no encontrado
- `429` - Demasiados intentos (rate limit: 5 requests cada 15 minutos)

**Best Practices:**
- Guardar el nuevo refresh_token y descartar el anterior
- Llamar a este endpoint cuando el access_token expira (HTTP 401)
- Implementar retry automático en el frontend para renovar tokens

---

## 💰 Accounts

### POST /accounts

Crear cuenta (personal o familiar).

**Headers:** `Authorization`

**Request (Personal):**
```json
{
  "name": "Finanzas Personales",
  "type": "personal",
  "currency": "ARS"
}
```

**Request (Family):**
```json
{
  "name": "Gastos Familia",
  "type": "family",
  "currency": "USD",
  "members": [
    { "name": "Mamá", "email": "mama@example.com" },
    { "name": "Papá" }
  ]
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "name": "Gastos Familia",
  "type": "family",
  "currency": "USD",
  "members": [
    { "id": "uuid", "name": "Mamá", "email": "mama@example.com" }
  ],
  "createdAt": "2026-01-16T10:00:00Z"
}
```

**Validations:**
- `type` obligatorio: `'personal'` o `'family'`
- `name` debe ser único por usuario (case-insensitive)
- Family requiere ≥1 miembro
- Personal no puede tener miembros
- Auto-crea meta "Ahorro General"

**Errors:**
- `400` - Datos inválidos
- `409` - Ya existe una cuenta con ese nombre

---

### GET /accounts

Listar cuentas del usuario.

**Headers:** `Authorization`

**Response (200):**
```json
{
  "accounts": [
    {
      "id": "uuid",
      "name": "Finanzas Personales",
      "type": "personal",
      "currency": "ARS",
      "createdAt": "2026-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "name": "Gastos Familia",
      "type": "family",
      "currency": "USD",
      "memberCount": 3,
      "createdAt": "2026-01-05T00:00:00Z"
    }
  ],
  "count": 2
}
```

---

### GET /accounts/:id

Detalle de cuenta.

**Headers:** `Authorization`

**Response (200):**
```json
{
  "id": "uuid",
  "name": "Gastos Familia",
  "type": "family",
  "currency": "ARS",
  "members": [
    {
      "id": "uuid",
      "name": "Mamá",
      "email": "mama@example.com",
      "isActive": true
    }
  ],
  "createdAt": "2026-01-01T00:00:00Z"
}
```

**Errors:**
- `403` - No tienes permiso
- `404` - Cuenta no encontrada

---

### PUT /accounts/:id

Actualizar cuenta.

**Headers:** `Authorization`

**Request:**
```json
{
  "name": "Nuevo Nombre",
  "currency": "USD"
}
```

**Note:** No se puede cambiar `type`

---

### DELETE /accounts/:id

Eliminar cuenta.

**Headers:** `Authorization`

**Validaciones:**
- Solo se puede eliminar si NO tiene gastos, ingresos o metas de ahorro asociadas
- Si tiene datos, retorna 409 Conflict

**Response (200):**
```json
{
  "message": "Cuenta eliminada exitosamente"
}
```

**Response (409):**
```json
{
  "error": "No se puede eliminar la cuenta porque tiene gastos, ingresos o metas asociadas"
}
```

---

### POST /accounts/:id/members

Agregar un nuevo miembro a una cuenta familiar.

**Headers:** `Authorization`

**Request:**
```json
{
  "name": "Pedro Pérez",
  "email": "pedro@example.com"
}
```

**Validaciones:**
- Solo funciona en cuentas de tipo `family`
- El nombre no puede estar vacío
- No puede existir otro miembro activo con el mismo nombre en la misma cuenta

**Response (201):**
```json
{
  "message": "Miembro agregado exitosamente",
  "member": {
    "id": "uuid",
    "name": "Pedro Pérez",
    "email": "pedro@example.com",
    "isActive": true
  }
}
```

**Errors:**
- `400` - Cuenta no es de tipo family o datos inválidos
- `404` - Cuenta no encontrada
- `409` - Ya existe un miembro activo con ese nombre

---

### PUT /accounts/:id/members/:member_id

Actualizar nombre y/o email de un miembro existente.

**Headers:** `Authorization`

**Request:**
```json
{
  "name": "Pedro García",
  "email": "pedro.garcia@example.com"
}
```

**Nota:** Al menos uno de los campos (`name` o `email`) debe estar presente.

**Validaciones:**
- El miembro debe pertenecer a la cuenta especificada
- El nombre no puede estar vacío
- Si se cambia el nombre, no puede coincidir con otro miembro activo

**Response (200):**
```json
{
  "message": "Miembro actualizado exitosamente",
  "member": {
    "id": "uuid",
    "name": "Pedro García",
    "email": "pedro.garcia@example.com",
    "isActive": true
  }
}
```

**Errors:**
- `400` - Datos inválidos o ningún campo presente
- `404` - Cuenta o miembro no encontrado
- `409` - Ya existe otro miembro activo con ese nombre

---

### PATCH /accounts/:id/members/:member_id/deactivate

Desactivar un miembro (soft delete). El miembro deja de aparecer en listados pero se preserva en la base de datos.

**Headers:** `Authorization`

**Response (200):**
```json
{
  "message": "Miembro desactivado exitosamente",
  "member": {
    "id": "uuid",
    "name": "Pedro García",
    "email": "pedro.garcia@example.com",
    "isActive": false
  }
}
```

**Errors:**
- `400` - El miembro ya está inactivo
- `404` - Cuenta o miembro no encontrado

---

### PATCH /accounts/:id/members/:member_id/reactivate

Reactivar un miembro previamente desactivado.

**Headers:** `Authorization`

**Validaciones:**
- No puede existir otro miembro activo con el mismo nombre (debe desactivarlo primero o cambiar el nombre del miembro a reactivar)

**Response (200):**
```json
{
  "message": "Miembro reactivado exitosamente",
  "member": {
    "id": "uuid",
    "name": "Pedro García",
    "email": "pedro.garcia@example.com",
    "isActive": true
  }
}
```

**Errors:**
- `400` - El miembro ya está activo
- `404` - Cuenta o miembro no encontrado
- `409` - Ya existe otro miembro activo con ese nombre

---

## 💸 Expenses

### POST /expenses

Crear gasto.

**Headers:** `Authorization`, `X-Account-ID`

**Request (One-Time):**
```json
{
  "description": "Supermercado",
  "amount": 25000,
  "currency": "ARS",
  "expense_type": "one-time",
  "date": "2026-01-16",
  "category_id": "uuid (opcional)",
  "family_member_id": "uuid (si family)"
}
```

**Request (Recurring):**
```json
{
  "description": "Netflix",
  "amount": 5000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-15",
  "end_date": null
}
```

**Request (Multi-Currency Modo 3):**
```json
{
  "description": "Claude Pro",
  "amount": 20,
  "currency": "USD",
  "amount_in_primary_currency": 31500,
  "date": "2026-01-16"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "description": "Claude Pro",
  "amount": 20.00,
  "currency": "USD",
  "exchange_rate": 1575.00,
  "amount_in_primary_currency": 31500.00,
  "expense_type": "one-time",
  "date": "2026-01-16",
  "category_name": "Tecnología",
  "created_at": "2026-01-16T10:00:00Z"
}
```

**Validations:**
- `amount` > 0
- `expense_type`: `'one-time'` o `'recurring'`
- One-time NO puede tener `end_date`
- Recurring puede tener `end_date` opcional
- Family accounts requieren `family_member_id`

**Multi-Currency:**
Ver [docs/MULTI-CURRENCY.md](./docs/MULTI-CURRENCY.md) para detalles del Modo 3.

---

### GET /expenses

Listar gastos.

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `month` (opcional): `YYYY-MM`
- `type` (opcional): `'one-time'`, `'recurring'`, `'all'`
- `category_id` (opcional): UUID
- `family_member_id` (opcional): UUID
- `currency` (opcional): `'ARS'`, `'USD'`, `'all'`

**Response (200):**
```json
{
  "expenses": [
    {
      "id": "uuid",
      "description": "Supermercado",
      "amount": 25000,
      "currency": "ARS",
      "amount_in_primary_currency": 25000,
      "expense_type": "one-time",
      "date": "2026-01-16",
      "category_name": "Alimentación"
    }
  ],
  "count": 1,
  "summary": {
    "total": 25000,
    "byType": {
      "one-time": 25000,
      "recurring": 0
    }
  }
}
```

---

### GET /expenses/:id

Detalle de gasto.

**Headers:** `Authorization`, `X-Account-ID`

**Response (200):** Similar a POST response con más detalles.

---

### PUT /expenses/:id

Actualizar gasto.

**Headers:** `Authorization`, `X-Account-ID`

**Request:** Todos los campos opcionales excepto ID

**Note:** No se puede cambiar `expense_type`

---

### DELETE /expenses/:id

Eliminar gasto.

**Headers:** `Authorization`, `X-Account-ID`

**Response (200):**
```json
{
  "message": "Gasto eliminado exitosamente"
}
```

---

## 🔁 Recurring Expenses (Templates)

**Patrón "Recurring Templates":** Los gastos recurrentes se gestionan mediante **templates** que generan automáticamente gastos reales en la tabla `expenses` vía CRON job diario (ejecuta a las 00:01 UTC).

**Ventajas:**
- Las estadísticas consultan solo `expenses` (gastos reales), sin cálculos complejos
- Editar el template preserva histórico automáticamente
- Trazabilidad perfecta (FK `recurring_expense_id` en expenses)

---

### POST /recurring-expenses

Crear template de gasto recurrente.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Monthly - Netflix):**
```json
{
  "description": "Netflix Subscription",
  "amount": 5000,
  "currency": "ARS",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 15,
  "start_date": "2026-01-01",
  "category_id": "uuid (opcional)"
}
```

**Request (Daily - Café):**
```json
{
  "description": "Café diario",
  "amount": 500,
  "currency": "ARS",
  "recurrence_frequency": "daily",
  "recurrence_interval": 1,
  "start_date": "2026-01-01"
}
```

**Request (Weekly - Gimnasio):**
```json
{
  "description": "Clase de yoga",
  "amount": 2000,
  "currency": "ARS",
  "recurrence_frequency": "weekly",
  "recurrence_day_of_week": 1,
  "start_date": "2026-01-06"
}
```
**Nota:** `recurrence_day_of_week`: 0=Domingo, 1=Lunes, ..., 6=Sábado

**Request (6 cuotas mensuales):**
```json
{
  "description": "Notebook en 6 cuotas",
  "amount": 50000,
  "currency": "ARS",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 10,
  "start_date": "2026-01-10",
  "total_occurrences": 6
}
```

**Response (201):**
```json
{
  "message": "Gasto recurrente creado exitosamente",
  "recurring_expense": {
    "id": "uuid",
    "account_id": "uuid",
    "description": "Netflix Subscription",
    "amount": 5000,
    "currency": "ARS",
    "recurrence_frequency": "monthly",
    "recurrence_interval": 1,
    "recurrence_day_of_month": 15,
    "start_date": "2026-01-01",
    "current_occurrence": 0,
    "is_active": true,
    "created_at": "2026-01-18T10:00:00Z"
  }
}
```

**Validations:**
- `recurrence_frequency`: `'daily'`, `'weekly'`, `'monthly'`, `'yearly'` (obligatorio)
- Monthly/yearly REQUIERE `recurrence_day_of_month` (1-31)
- Weekly REQUIERE `recurrence_day_of_week` (0-6)
- `recurrence_interval`: cada N períodos (default: 1)
- `amount` > 0
- `start_date`: formato YYYY-MM-DD
- `end_date`: opcional, debe ser >= start_date
- `total_occurrences`: opcional, límite de repeticiones

**Edge Cases:**
- Día 31 en meses cortos → se genera el último día del mes (ej: 28/29 feb)
- Feb 29 en años no bisiestos → se genera el 28 de febrero

---

### GET /recurring-expenses

Listar templates de gastos recurrentes.

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `is_active` (opcional): `'true'`, `'false'`, `'all'` (default: `'true'`)
- `frequency` (opcional): `'daily'`, `'weekly'`, `'monthly'`, `'yearly'`
- `page` (opcional): número de página (default: 1)
- `limit` (opcional): items por página (default: 20, max: 100)

**Response (200):**
```json
{
  "recurring_expenses": [
    {
      "id": "uuid",
      "description": "Netflix Subscription",
      "amount": 5000,
      "currency": "ARS",
      "category_name": "Entretenimiento",
      "recurrence_frequency": "monthly",
      "recurrence_interval": 1,
      "recurrence_day_of_month": 15,
      "start_date": "2026-01-01",
      "current_occurrence": 3,
      "is_active": true,
      "created_at": "2026-01-01T10:00:00Z"
    }
  ],
  "count": 1,
  "total": 10,
  "page": 1,
  "limit": 20
}
```

---

### GET /recurring-expenses/:id

Obtener detalle de un template.

**Headers:** `Authorization`, `X-Account-ID`

**Response (200):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "description": "Netflix Subscription",
  "amount": 5000,
  "currency": "ARS",
  "category_id": "uuid",
  "category_name": "Entretenimiento",
  "recurrence_frequency": "monthly",
  "recurrence_interval": 1,
  "recurrence_day_of_month": 15,
  "start_date": "2026-01-01",
  "end_date": null,
  "total_occurrences": null,
  "current_occurrence": 3,
  "exchange_rate": 1.0,
  "amount_in_primary_currency": 5000,
  "is_active": true,
  "created_at": "2026-01-01T10:00:00Z",
  "updated_at": "2026-01-18T10:00:00Z",
  "generated_expenses_count": 3
}
```

**Nota:** `generated_expenses_count` muestra cuántos gastos se generaron desde este template.

---

### PUT /recurring-expenses/:id

Actualizar template (solo afecta FUTUROS gastos, preserva histórico).

**Headers:** `Authorization`, `X-Account-ID`

**Request (aumento de precio):**
```json
{
  "amount": 6000,
  "description": "Netflix Subscription (price increased)"
}
```

**Request (cancelar desde hoy):**
```json
{
  "end_date": "2026-01-18",
  "is_active": false
}
```

**Response (200):**
```json
{
  "message": "Gasto recurrente actualizado exitosamente",
  "updated_at": "2026-01-18T10:00:00Z",
  "note": "Los gastos ya generados NO se modifican. Solo afecta futuros gastos."
}
```

**Validaciones:**
- Partial update (solo campos enviados se actualizan)
- Frequency-specific fields validados (ej: no puedes setear day_of_month si no es monthly/yearly)
- Set a NULL: enviar campo vacío (ej: `"end_date": ""` → SET NULL)

---

### DELETE /recurring-expenses/:id

Desactivar template (soft delete - detiene generación futura).

**Headers:** `Authorization`, `X-Account-ID`

**Response (200):**
```json
{
  "message": "Gasto recurrente eliminado exitosamente",
  "generated_expenses": 3,
  "note": "Los gastos ya generados NO se eliminan. Solo se detiene la generación futura."
}
```

**Comportamiento:**
- SOFT DELETE: marca `is_active = false`
- NO borra el template de la DB
- NO borra los gastos ya generados
- Detiene la generación de nuevos gastos

---

## 🔁 Recurring Incomes (Templates)

**Patrón "Recurring Templates":** Idéntico a recurring-expenses, pero genera automáticamente ingresos reales en la tabla `incomes` vía CRON job diario (ejecuta a las 00:01 UTC).

**Endpoints:** Misma estructura que `/recurring-expenses`

- `POST /recurring-incomes` - Crear template de ingreso recurrente
- `GET /recurring-incomes` - Listar templates (acepta `is_active`, `frequency`, `page`, `limit`)
- `GET /recurring-incomes/:id` - Obtener detalle con `generated_incomes_count`
- `PUT /recurring-incomes/:id` - Actualizar template (solo afecta futuros ingresos)
- `DELETE /recurring-incomes/:id` - Soft delete (marca `is_active = false`)

**Request Example (Salario mensual):**
```json
{
  "description": "Salario Mensual",
  "amount": 500000,
  "currency": "ARS",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 1,
  "start_date": "2026-01-01"
}
```

**Use Cases:**
- Salario mensual (monthly, day 1 o día de cobro)
- Ingresos por alquiler (monthly, día específico)
- Freelance recurrente (weekly/monthly)
- Rentas de inversiones (monthly/yearly)

**Nota:** Ver documentación completa en sección `/recurring-expenses` - funcionamiento idéntico.

---

## 💰 Incomes

Los endpoints de ingresos funcionan idénticamente a expenses.

### POST /incomes

**Request:**
```json
{
  "description": "Sueldo mensual",
  "amount": 200000,
  "currency": "ARS",
  "income_type": "recurring",
  "date": "2026-01-01",
  "end_date": null,
  "category_id": "uuid (opcional)",
  "family_member_id": "uuid (si family)"
}
```

**Types:**
- `one-time` - Ingreso único
- `recurring` - Ingreso recurrente

**Multi-Currency:** Soporta Modo 3 igual que expenses.

---

### GET /incomes

Query params idénticos a expenses:
- `month`, `type`, `category_id`, `family_member_id`, `currency`

---

### GET /incomes/:id

Detalle de ingreso.

---

### PUT /incomes/:id

Actualizar ingreso.

---

### DELETE /incomes/:id

Eliminar ingreso.

---

## 📊 Dashboard

### GET /dashboard/summary

Resumen financiero del mes.

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `month` (opcional): `YYYY-MM` (default: mes actual)

**Response (200):**
```json
{
  "period": "2026-01",
  "primary_currency": "ARS",
  "total_income": 200000.00,
  "total_expenses": 120000.00,
  "total_assigned_to_goals": 30000.00,
  "available_balance": 50000.00,
  "expenses_by_category": [
    {
      "category_id": "uuid",
      "category_name": "Alimentación",
      "category_icon": "🍔",
      "category_color": "#FF6B6B",
      "total": 45000.00,
      "percentage": 37.5
    }
  ],
  "top_expenses": [
    {
      "id": "uuid",
      "description": "Supermercado",
      "amount": 25000.00,
      "currency": "ARS",
      "amount_in_primary_currency": 25000.00,
      "category_id": "uuid",
      "category_name": "Alimentación",
      "category_icon": "🍔",
      "category_color": "#FF6B6B",
      "date": "2026-01-10",
      "created_at": "2026-01-10T08:30:00Z"
    }
  ],
  "recent_transactions": [
    {
      "id": "uuid",
      "type": "expense",
      "description": "Supermercado",
      "amount": 25000.00,
      "currency": "ARS",
      "amount_in_primary_currency": 25000.00,
      "category_id": "uuid",
      "category_name": "Alimentación",
      "date": "2026-01-10",
      "created_at": "2026-01-10T08:30:00Z"
    },
    {
      "id": "uuid",
      "type": "income",
      "description": "Sueldo",
      "amount": 200000.00,
      "currency": "ARS",
      "amount_in_primary_currency": 200000.00,
      "category_id": "uuid",
      "category_name": "Salario",
      "date": "2026-01-05",
      "created_at": "2026-01-05T10:00:00Z"
    }
  ]
}
```

**Campos importantes:**
- `total_assigned_to_goals`: Total de fondos en metas activas (capital inmovilizado). Representa la suma del `current_amount` de todas las metas de ahorro activas, NO solo fondos agregados este mes.
- `available_balance`: Dinero disponible para gastar = `total_income - total_expenses - total_assigned_to_goals`

**Cálculo:**
```
available_balance = total_income - total_expenses - total_assigned_to_goals
```

**Notas:**
- Todos los montos en moneda primaria (conversión automática vía `amount_in_primary_currency`)
- `top_expenses`: Máximo 5 gastos más grandes del mes (incluye info de categoría si existe)
- `recent_transactions`: Máximo 10 transacciones (expenses + incomes mezclados, ordenados por `created_at DESC`)

---

## 🎯 Savings Goals

### POST /savings-goals

Crear meta de ahorro.

**Headers:** `Authorization`, `X-Account-ID`

**Request:**
```json
{
  "name": "Vacaciones en Brasil",
  "target_amount": 300000,
  "currency": "ARS",
  "deadline": "2026-06-30",
  "description": "Viaje familiar"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "name": "Vacaciones en Brasil",
  "target_amount": 300000.00,
  "current_amount": 0.00,
  "currency": "ARS",
  "deadline": "2026-06-30",
  "is_active": true,
  "progress_percentage": 0.0,
  "required_monthly_savings": 50000.00,
  "created_at": "2026-01-16T10:00:00Z"
}
```

**Fields:**
- `deadline` - Opcional (null = sin deadline, debe ser fecha futura)
- `required_monthly_savings` - Auto-calculado basado en deadline y monto faltante. Retorna `null` si no hay deadline o si el deadline ya pasó. Fórmula: `(target_amount - current_amount) / meses_restantes`

---

### GET /savings-goals

Listar metas.

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `is_active` (opcional): `true` | `false` | `all` (default: `true`)
  - `true` - Solo metas activas
  - `false` - Solo metas archivadas
  - `all` - Todas las metas (activas + archivadas)

**Response (200):**
```json
{
  "goals": [
    {
      "id": "uuid",
      "name": "Vacaciones",
      "target_amount": 300000,
      "current_amount": 50000,
      "progress_percentage": 16.67,
      "deadline": "2026-06-30",
      "required_monthly_savings": 50000.00
    }
  ],
  "count": 1
}
```

**Note:** El campo `required_monthly_savings` se calcula automáticamente para cada meta y solo aparece si tiene deadline futuro.

---

### GET /savings-goals/:id

Detalle con historial de transacciones (paginado).

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `page` (opcional): Número de página (default: 1)
- `limit` (opcional): Transacciones por página (default: 20, max: 100)

**Response (200):**
```json
{
  "id": "uuid",
  "name": "Vacaciones",
  "target_amount": 300000,
  "current_amount": 50000,
  "transactions": [
    {
      "id": "uuid",
      "amount": 30000,
      "transaction_type": "deposit",
      "description": "Ahorro enero",
      "date": "2026-01-15",
      "created_at": "2026-01-15T10:00:00Z"
    },
    {
      "id": "uuid",
      "amount": 20000,
      "transaction_type": "deposit",
      "date": "2026-01-20",
      "created_at": "2026-01-20T14:30:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 1,
    "total_count": 2,
    "limit": 20
  }
}
```

**Note:** Las transacciones de tipo `withdrawal` se muestran con `amount` negativo para facilitar la visualización.

---

### GET /savings-goals/:id/transactions

Obtener solo el historial de transacciones de una meta (endpoint dedicado).

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `page` (opcional): Número de página (default: 1)
- `limit` (opcional): Transacciones por página (default: 20, max: 100)
- `type` (opcional): `all` | `deposit` | `withdrawal` (default: `all`)
  - `all` - Todas las transacciones
  - `deposit` - Solo depósitos (fondos agregados)
  - `withdrawal` - Solo retiros (fondos retirados)

**Response (200):**
```json
{
  "transactions": [
    {
      "id": "uuid",
      "amount": 30000,
      "transaction_type": "deposit",
      "description": "Ahorro enero",
      "date": "2026-01-15",
      "created_at": "2026-01-15T10:00:00Z"
    },
    {
      "id": "uuid",
      "amount": -10000,
      "transaction_type": "withdrawal",
      "description": "Adelanto para pasajes",
      "date": "2026-01-18",
      "created_at": "2026-01-18T16:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "total_count": 47,
    "limit": 20
  }
}
```

**Validación:**
- `type` inválido → HTTP 400: `"type must be 'all', 'deposit', or 'withdrawal'"`

**Ejemplo de uso:**
```bash
# Obtener solo depósitos paginados
GET /api/savings-goals/:id/transactions?type=deposit&page=1&limit=10

# Obtener solo retiros
GET /api/savings-goals/:id/transactions?type=withdrawal
```

---

### POST /savings-goals/:id/add-funds

Agregar fondos a meta.

**Headers:** `Authorization`, `X-Account-ID`

**Request:**
```json
{
  "amount": 30000,
  "description": "Ahorro enero",
  "date": "2026-01-15"
}
```

**Validations:**
- `amount` - Requerido, debe ser > 0
- `date` - Requerido, formato YYYY-MM-DD
  - No puede ser fecha futura
  - No puede ser posterior al `deadline` de la meta (si existe)
- `description` - Opcional

**Response (200):**
```json
{
  "message": "Fondos agregados exitosamente",
  "savings_goal": {
    "id": "uuid",
    "name": "Vacaciones",
    "current_amount": 80000.00,
    "target_amount": 300000.00,
    "progress_percentage": 26.67,
    "updated_at": "2026-01-15T10:30:00Z"
  },
  "transaction": {
    "id": "uuid",
    "amount": 30000,
    "transaction_type": "deposit",
    "description": "Ahorro enero",
    "date": "2026-01-15",
    "created_at": "2026-01-15T10:30:00Z"
  }
}
```

**Error (400) - Fecha posterior al deadline:**
```json
{
  "error": "no puedes agregar fondos con una fecha posterior al deadline de la meta",
  "transaction_date": "2026-07-15",
  "goal_deadline": "2026-06-30"
}
```

**Effect:**
- Actualiza `current_amount` automáticamente
- Crea registro en `savings_goal_transactions`
- Se cuenta en `total_assigned_to_goals` del dashboard

---

### POST /savings-goals/:id/withdraw-funds

Retirar fondos de meta.

**Headers:** `Authorization`, `X-Account-ID`

**Request:**
```json
{
  "amount": 10000,
  "description": "Adelanto para pasaje",
  "date": "2026-01-18"
}
```

**Validations:**
- `amount` - Requerido, debe ser > 0 y ≤ current_amount
- `date` - Requerido, formato YYYY-MM-DD
  - No puede ser fecha futura
  - No puede ser posterior al `deadline` de la meta (si existe)
- `description` - Opcional

**Response (200):**
```json
{
  "message": "Fondos retirados exitosamente",
  "transaction": {
    "id": "uuid",
    "type": "withdraw",
    "amount": 10000
  },
  "new_current_amount": 70000.00
}
```

**Validation:**
- No se puede retirar más de `current_amount`

---

## 🏷️ Categories

### GET /expense-categories

Listar categorías de gastos (predefinidas + custom).

**Headers:** `Authorization`, `X-Account-ID`

**Response (200):**
```json
{
  "categories": [
    {
      "id": "uuid",
      "name": "Alimentación",
      "icon": "🍔",
      "color": "#FF6B6B",
      "is_custom": false
    },
    {
      "id": "uuid",
      "name": "Mi Categoría Custom",
      "icon": "🎯",
      "color": "#00FF00",
      "is_custom": true
    }
  ],
  "count": 16
}
```

**Predefined Categories (15):**
1. Alimentación 🍔 #FF6B6B
2. Transporte 🚗 #4ECDC4
3. Salud ⚕️ #95E1D3
4. Entretenimiento 🎮 #F38181
5. Educación 📚 #AA96DA
6. Hogar 🏠 #FCBAD3
7. Servicios 💡 #A8D8EA
8. Ropa 👕 #FFCCBC
9. Mascotas 🐶 #C5E1A5
10. Tecnología 💻 #90CAF9
11. Viajes ✈️ #FFAB91
12. Regalos 🎁 #F48FB1
13. Impuestos 🧾 #BCAAA4
14. Seguros 🛡️ #B39DDB
15. Otro 📦 #B0BEC5

---

### POST /expense-categories

Crear categoría custom.

**Headers:** `Authorization`, `X-Account-ID`

**Request:**
```json
{
  "name": "Veterinario",
  "icon": "🐕",
  "color": "#FF5733"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "name": "Veterinario",
  "icon": "🐕",
  "color": "#FF5733",
  "is_system": false,
  "created_at": "2026-01-19T01:30:00Z"
}
```

**Response (409) - Duplicate name:**
```json
{
  "error": "Ya existe una categoría con ese nombre en esta cuenta"
}
```

**Validation Rules:**
- `name`: Required, must be unique per account (case-insensitive)
  - "Alimentación" and "alimentación" are considered duplicates
  - "Alimentación" in Account A can exist alongside "Alimentación" in Account B
- `icon`: Required, emoji character
- `color`: Required, hex color code (e.g., "#FF5733")

**Restrictions:**
- No se pueden editar/borrar categorías del sistema (is_system = true)
- No se pueden borrar categorías custom con gastos asociados
- Nombres únicos por cuenta (sin importar mayúsculas/minúsculas)

---

### GET /income-categories

Listar categorías de ingresos.

**Predefined (10):**
1. Salario 💼 #66BB6A
2. Freelance 💻 #42A5F5
3. Inversiones 📈 #AB47BC
4. Negocio 🏢 #FFA726
5. Alquiler 🏘️ #26C6DA
6. Regalo 🎁 #EC407A
7. Venta 🏷️ #78909C
8. Intereses 💰 #9CCC65
9. Reembolso ↩️ #7E57C2
10. Otro 💵 #8D6E63

---

### POST /income-categories

Crear categoría custom de ingresos.

**Headers:** `Authorization`, `X-Account-ID`

**Request:**
```json
{
  "name": "Bonus Anual",
  "icon": "💎",
  "color": "#4CAF50"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "name": "Bonus Anual",
  "icon": "💎",
  "color": "#4CAF50",
  "is_system": false,
  "created_at": "2026-01-19T01:30:00Z"
}
```

**Response (409) - Duplicate name:**
```json
{
  "error": "Ya existe una categoría con ese nombre en esta cuenta"
}
```

**Validation Rules:**
- Same as expense-categories (unique name per account, case-insensitive)
- Icons and colors should reflect income context

---

## ❌ Error Responses

Todas las respuestas de error siguen este formato:

```json
{
  "error": "Mensaje descriptivo",
  "details": "Información adicional (opcional)"
}
```

### HTTP Status Codes

- `200` - Éxito
- `201` - Creado
- `400` - Request inválido
- `401` - No autenticado
- `403` - Sin permisos
- `404` - No encontrado
- `409` - Conflicto (ej: email duplicado)
- `500` - Error del servidor

### Common Errors

| Error | Causa | Solución |
|-------|-------|----------|
| `account_id not found in context` | Falta header `X-Account-ID` | Agregar header |
| `Usuario no autenticado` | Token JWT inválido/faltante | Verificar Authorization |
| `Datos inválidos` | Campo requerido faltante o formato incorrecto | Validar payload |
| `El email ya está registrado` | Email duplicado en registro | Usar otro email o login |
| `Ya existe una cuenta con ese nombre` | Nombre de cuenta duplicado (case-insensitive) | Usar otro nombre de cuenta |
| `Ya existe una categoría con ese nombre en esta cuenta` | Nombre de categoría duplicado en la misma cuenta (case-insensitive) | Usar otro nombre de categoría |

---

## 🎓 Best Practices

### Frontend

**1. Guardar account activo en localStorage:**
```typescript
localStorage.setItem('activeAccountId', account.id);
```

**2. Axios interceptor para X-Account-ID:**
```typescript
axios.interceptors.request.use(config => {
  const accountId = localStorage.getItem('activeAccountId');
  if (accountId) {
    config.headers['X-Account-ID'] = accountId;
  }
  return config;
});
```

**3. Validar con Zod antes de enviar:**
```typescript
const CreateExpenseSchema = z.object({
  amount: z.number().positive(),
  currency: z.enum(['ARS', 'USD', 'EUR']),
  description: z.string().min(1).max(500),
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
});

const data = CreateExpenseSchema.parse(formData);
```

**4. Manejar errores consistentemente:**
```typescript
try {
  await createExpense(data);
} catch (error) {
  if (axios.isAxiosError(error)) {
    const apiError = error.response?.data;
    toast.error(apiError.error);
  }
}
```

---

## 📚 See Also

- [FEATURES.md](./FEATURES.md) - Guía narrativa de funcionalidades
- [STACK.md](./STACK.md) - Stack tecnológico
- [docs/DATABASE.md](./docs/DATABASE.md) - Schema de base de datos
- [docs/MULTI-CURRENCY.md](./docs/MULTI-CURRENCY.md) - Sistema multi-moneda
- [docs/RECURRENCE.md](./docs/RECURRENCE.md) - Sistema de recurrencia

---

**Creado:** 2026-01-15  
**Última actualización:** 2026-01-18 (Recurring Expenses Templates added)
**Versión:** 2.0 (Consolidada)  
**Mantenido por:** Gentleman Programming & Lorenzo
