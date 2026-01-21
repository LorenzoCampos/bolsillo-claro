# 📚 Bolsillo Claro - API Documentation

**Base URL:** `https://api.fakerbostero.online/bolsillo/api`  
**Versión:** 2.5  
**Última actualización:** 2026-01-21

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

Actualizar cuenta (partial update).

**Headers:** `Authorization`

**Request (Ejemplo completo):**
```json
{
  "name": "Nuevo Nombre de Cuenta",
  "currency": "USD"
}
```

**Request (Solo nombre):**
```json
{
  "name": "Mi Cuenta Personal"
}
```

**Request (Solo moneda):**
```json
{
  "currency": "ARS"
}
```

**Campos actualizables (ambos opcionales):**
- `name` - Nombre de la cuenta (1-100 caracteres)
  - Debe ser único por usuario (case-insensitive)
- `currency` - Moneda primaria de la cuenta
  - Valores permitidos: `"ARS"`, `"USD"`, `"EUR"`
  - ⚠️ **Cambiar la moneda afecta a todas las operaciones futuras**

**Campos NO modificables:**
- `type` - El tipo de cuenta (personal/family) NO se puede cambiar una vez creada
- `user_id` - El propietario de la cuenta no puede cambiar

**Validaciones:**
- Al menos uno de los campos (`name` o `currency`) debe estar presente
- Si se proporciona `name`, debe tener entre 1 y 100 caracteres
- El nombre debe ser único entre todas las cuentas activas del usuario

**Response (200):**
```json
{
  "message": "Cuenta actualizada exitosamente",
  "account": {
    "id": "uuid",
    "name": "Nuevo Nombre de Cuenta",
    "type": "personal",
    "currency": "USD",
    "createdAt": "2026-01-01T00:00:00Z",
    "updatedAt": "2026-01-21T10:30:00Z"
  }
}
```

**Errors:**
- `400` - Datos inválidos o ningún campo presente
- `404` - Cuenta no encontrada o no pertenece al usuario
- `409` - Ya existe otra cuenta con ese nombre

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

Crear gasto puntual (one-time).

**Headers:** `Authorization`, `X-Account-ID`

**Request (Mínimo - ARS):**
```json
{
  "description": "Supermercado",
  "amount": 25000,
  "currency": "ARS",
  "date": "2026-01-16"
}
```

**Request (Completo - con categoría y miembro):**
```json
{
  "description": "Supermercado Carrefour",
  "amount": 25000,
  "currency": "ARS",
  "date": "2026-01-16",
  "category_id": "uuid-categoria-comida",
  "family_member_id": "uuid-miembro-papa"
}
```

**Request (Multi-Currency Modo 3):**
```json
{
  "description": "Claude Pro (tarjeta con impuestos)",
  "amount": 20,
  "currency": "USD",
  "amount_in_primary_currency": 31500,
  "date": "2026-01-16"
}
```

**Campos requeridos:**
- `description` - Descripción del gasto (no vacío)
- `amount` - Monto gastado (debe ser > 0)
- `currency` - Moneda del gasto
  - Valores permitidos: `"ARS"`, `"USD"`, `"EUR"`
- `date` - Fecha del gasto (formato: YYYY-MM-DD)

**Campos opcionales:**
- `category_id` - UUID de categoría de gasto
  - Si no se proporciona o es `null`, se usa categoría "Otro" por defecto
  - Debe existir en `expense_categories`
- `family_member_id` - UUID del miembro familiar
  - Solo válido para cuentas tipo `family`
  - Debe pertenecer a la cuenta actual
- `expense_type` - Tipo de gasto (default: `"one-time"`)
  - ⚠️ **NO uses este campo manualmente**. Se usa solo para gastos generados por recurring_expenses
  - Valores: `"one-time"` | `"recurring"`
- `end_date` - Fecha fin (formato: YYYY-MM-DD)
  - Solo para `expense_type: "recurring"` (generado por scheduler)
  - ❌ No se puede usar con `expense_type: "one-time"`

**Campos opcionales (Multi-Currency - Modo 3):**
- `exchange_rate` - Tasa de cambio manual (ej: 1575.00)
- `amount_in_primary_currency` - Monto REAL debitado en moneda primaria
  - **Modo 3 preferido:** Enviás cuántos USD gastaste Y cuántos ARS te debitaron
  - El sistema calcula automáticamente: `exchange_rate = amount_in_primary_currency / amount`
  - Ejemplo: gastaste USD 20, te debitaron ARS 31500 → exchange_rate = 1575

**Campos auto-generados:**
- `id` - UUID del gasto
- `account_id` - Heredado del header `X-Account-ID`
- `exchange_rate` - Calculado automáticamente según Modo Multi-Currency
- `amount_in_primary_currency` - Calculado automáticamente
- `created_at` - Timestamp de creación

**Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "family_member_id": null,
  "category_id": "uuid",
  "category_name": "Tecnología",
  "description": "Claude Pro",
  "amount": 20.00,
  "currency": "USD",
  "exchange_rate": 1575.00,
  "amount_in_primary_currency": 31500.00,
  "expense_type": "one-time",
  "date": "2026-01-16",
  "end_date": null,
  "created_at": "2026-01-16T10:00:00Z"
}
```

**Validaciones:**
- `amount` debe ser > 0
- `currency` debe ser ARS, USD o EUR
- `date` debe ser formato YYYY-MM-DD válido
- Si `expense_type` es `"one-time"`, NO puede tener `end_date`
- Si `expense_type` es `"recurring"` y tiene `end_date`, debe ser >= `date`
- Si `family_member_id` se proporciona, debe pertenecer a la cuenta
- Si `category_id` se proporciona, debe existir en la DB

**Multi-Currency - Modos de cálculo:**
1. **Modo 1 (Misma moneda):** `currency == primary_currency`
   - `exchange_rate = 1.0`
   - `amount_in_primary_currency = amount`

2. **Modo 2 (Tasa manual):** Proporcionás `exchange_rate`
   - `amount_in_primary_currency = amount * exchange_rate`

3. **Modo 3 (Monto real - PREFERIDO):** Proporcionás `amount_in_primary_currency`
   - `exchange_rate = amount_in_primary_currency / amount`
   - **Ejemplo:** USD 20 gastado, ARS 31500 debitado → rate = 1575

4. **Modo Auto:** Si no proporcionás nada, busca en tabla `exchange_rates`
   - Si no encuentra tasa para esa fecha, retorna **HTTP 400** pidiendo que proporcion és `exchange_rate` o `amount_in_primary_currency`

**Errors:**
- `400` - Datos inválidos, formato de fecha incorrecto, validaciones fallidas
- `400` - No se encontró tasa de cambio (proporcionar exchange_rate o amount_in_primary_currency)
- `400` - family_member_id no pertenece a la cuenta

**⚠️ Nota sobre gastos recurrentes:**
Para gastos que se repiten regularmente (Netflix, alquiler, etc.), **NO uses este endpoint**. En su lugar:
1. Usá `POST /recurring-expenses` para crear un **template**
2. El scheduler generará automáticamente los gastos reales con `expense_type: "recurring"`

Ver [docs/MULTI-CURRENCY.md](./docs/MULTI-CURRENCY.md) para más detalles del Modo 3.

---

### GET /expenses

Listar gastos.

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `month` (opcional): `YYYY-MM`
- `type` (opcional): `'one-time'`, `'recurring'`, `'all'`
- `category_id` (opcional): UUID
- `family_member_id` (opcional): UUID
- `currency` (opcional): `'ARS'`, `'USD'`, `'EUR'`, `'all'`

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

Actualizar gasto. Permite actualización parcial (solo enviás los campos que querés cambiar).

**Headers:** `Authorization`, `X-Account-ID`

**Request (Update Partial - Cambiar monto y categoría):**
```json
{
  "amount": 25.00,
  "category_id": "uuid-nueva-categoria"
}
```

**Request (Update Solo Descripción):**
```json
{
  "description": "Claude Pro - Plan Anual"
}
```

**Request (Limpiar end_date):**
```json
{
  "end_date": ""
}
```
**Nota:** Enviar `end_date: ""` (string vacío) limpia el campo (lo pone en NULL). Omitir el campo lo deja sin cambios.

**Campos actualizables (todos opcionales):**
- `description` - Nueva descripción del gasto (1-200 caracteres)
- `amount` - Nuevo monto (debe ser > 0)
- `currency` - Nueva moneda (ARS | USD | EUR)
  - ⚠️ Si cambiás la moneda, el sistema recalcula `exchange_rate` y `amount_in_primary_currency` automáticamente
- `date` - Nueva fecha del gasto (formato: YYYY-MM-DD)
  - ⚠️ Si cambiás la fecha, el sistema puede recalcular la tasa de cambio si usa tasas de la DB
- `category_id` - Nueva categoría (UUID válido o null)
- `family_member_id` - Nuevo miembro familiar (UUID válido o null)
  - Si se proporciona, debe pertenecer a la cuenta
- `end_date` - Nueva fecha fin para gastos recurrentes (formato: YYYY-MM-DD o "" para limpiar)
  - Solo válido si `expense_type` es `"recurring"`
  - Debe ser >= `date`
- `exchange_rate` - Nueva tasa de cambio manual (debe ser > 0)
  - Si se proporciona, se usa para recalcular `amount_in_primary_currency`
- `amount_in_primary_currency` - Nuevo monto en moneda primaria (debe ser > 0)
  - Si se proporciona, se usa para recalcular `exchange_rate`

**Campos NO modificables:**
- `id` - Identificador único del gasto (inmutable)
- `account_id` - Cuenta a la que pertenece (inmutable)
- `expense_type` - Tipo de gasto (inmutable - `"one-time"` o `"recurring"`)
  - No se puede cambiar porque podría violar reglas de negocio
- `recurring_expense_id` - Template que generó este gasto (inmutable)
- `created_at` - Timestamp de creación (inmutable)

**Validaciones:**
- Al menos un campo actualizable debe ser proporcionado (no se permiten updates vacíos)
- `amount` debe ser > 0
- `currency` debe ser ARS, USD o EUR
- `date` debe ser formato YYYY-MM-DD válido
- Si el gasto es `expense_type: "one-time"`, NO puede tener `end_date`
- Si el gasto es `expense_type: "recurring"` y tiene `end_date`, debe ser >= `date`
- Si `family_member_id` se proporciona, debe pertenecer a la cuenta
- Si `category_id` se proporciona, debe existir en la DB
- `exchange_rate` y `amount_in_primary_currency` deben ser > 0 si se proporcionan

**Multi-Currency - Recálculo Automático:**
Si actualizás `amount`, `currency`, o `date`, el sistema recalcula automáticamente la conversión usando:
1. **Modo 1 (Misma moneda):** `currency == primary_currency` → `exchange_rate = 1.0`
2. **Modo 2 (Tasa manual):** Si proporcionás `exchange_rate` → calcula `amount_in_primary_currency`
3. **Modo 3 (Monto real):** Si proporcionás `amount_in_primary_currency` → calcula `exchange_rate`
4. **Modo Auto:** Busca en tabla `exchange_rates` para la nueva fecha
   - Si no encuentra, retorna **HTTP 400** pidiendo que proporciones `exchange_rate` o `amount_in_primary_currency`

**Response (200):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "family_member_id": null,
  "category_id": "uuid-nueva-categoria",
  "category_name": "Tecnología",
  "description": "Claude Pro",
  "amount": 25.00,
  "currency": "USD",
  "exchange_rate": 1575.00,
  "amount_in_primary_currency": 39375.00,
  "expense_type": "one-time",
  "date": "2026-01-16",
  "end_date": null,
  "created_at": "2026-01-16T10:00:00Z"
}
```

**Errors:**
- `400` - Datos inválidos, formato de fecha incorrecto, validaciones fallidas
- `400` - No se encontró tasa de cambio (proporcionar exchange_rate o amount_in_primary_currency)
- `400` - No se proporcionaron campos para actualizar
- `400` - family_member_id no pertenece a la cuenta
- `404` - Gasto no encontrado o no pertenece a la cuenta

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

**Patrón "Recurring Templates":** Los ingresos recurrentes se gestionan mediante **templates** que generan automáticamente ingresos reales en la tabla `incomes` vía CRON job diario (ejecuta a las 00:01 UTC).

**Ventajas:**
- Las estadísticas consultan solo `incomes` (ingresos reales), sin cálculos complejos
- Editar el template preserva histórico automáticamente
- Trazabilidad perfecta (FK `recurring_income_id` en incomes)

---

### POST /recurring-incomes

Crear template de ingreso recurrente.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Monthly - Salario):**
```json
{
  "description": "Salario Mensual",
  "amount": 500000,
  "currency": "ARS",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 1,
  "start_date": "2026-01-01",
  "category_id": "uuid-categoria-salario"
}
```

**Request (Daily - Propinas):**
```json
{
  "description": "Propinas diarias",
  "amount": 2000,
  "currency": "ARS",
  "recurrence_frequency": "daily",
  "recurrence_interval": 1,
  "start_date": "2026-01-01"
}
```

**Request (Weekly - Freelance):**
```json
{
  "description": "Freelance semanal",
  "amount": 50000,
  "currency": "ARS",
  "recurrence_frequency": "weekly",
  "recurrence_day_of_week": 5,
  "start_date": "2026-01-03"
}
```
**Nota:** `recurrence_day_of_week`: 0=Domingo, 1=Lunes, ..., 6=Sábado

**Request (12 cuotas mensuales - Alquiler adelantado):**
```json
{
  "description": "Alquiler cobrado adelantado",
  "amount": 150000,
  "currency": "ARS",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 10,
  "start_date": "2026-01-10",
  "total_occurrences": 12
}
```

**Request (Multi-Currency - Freelance USA):**
```json
{
  "description": "Freelance USA mensual",
  "amount": 500,
  "currency": "USD",
  "amount_in_primary_currency": 787500,
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 15,
  "start_date": "2026-01-15"
}
```

**Campos requeridos:**
- `description` - Descripción del ingreso recurrente (1-200 caracteres)
- `amount` - Monto (debe ser > 0)
- `currency` - Moneda (ARS | USD | EUR)
- `recurrence_frequency` - Frecuencia de recurrencia
  - Valores: `"daily"` | `"weekly"` | `"monthly"` | `"yearly"`
- `start_date` - Fecha de inicio (formato: YYYY-MM-DD)
  - Primera fecha en la que se generará un ingreso

**Campos requeridos condicionales (según frecuencia):**
- `recurrence_day_of_month` - Día del mes (1-31)
  - **OBLIGATORIO** para `frequency: "monthly"` o `"yearly"`
  - ❌ NO se puede usar con `"daily"` o `"weekly"`
  - Edge case: Día 31 en meses cortos → se genera el último día del mes (28/29 feb)
- `recurrence_day_of_week` - Día de la semana (0-6)
  - **OBLIGATORIO** para `frequency: "weekly"`
  - ❌ NO se puede usar con otras frecuencias
  - 0=Domingo, 1=Lunes, 2=Martes, ..., 6=Sábado

**Campos opcionales:**
- `category_id` - UUID de categoría de ingreso (debe existir en income_categories)
- `family_member_id` - UUID de miembro familiar (debe pertenecer a la cuenta)
- `recurrence_interval` - Cada N períodos (default: 1)
  - Ejemplo: `interval: 2` con `frequency: "weekly"` = cada 2 semanas
- `end_date` - Fecha fin (formato: YYYY-MM-DD)
  - Debe ser >= `start_date`
  - El template se desactiva automáticamente cuando se alcanza
- `total_occurrences` - Límite de repeticiones (debe ser > 0)
  - Ejemplo: 12 para un año de ingresos mensuales
  - El template se desactiva automáticamente al alcanzar este número

**Campos opcionales (Multi-Currency - Modo 3):**
- `exchange_rate` - Tasa de cambio manual (ej: 1575.00)
- `amount_in_primary_currency` - Monto REAL acreditado en moneda primaria
  - **Modo 3 preferido:** Enviás cuántos USD recibís Y cuántos ARS te acreditan
  - El sistema calcula automáticamente: `exchange_rate = amount_in_primary_currency / amount`

**Campos auto-generados:**
- `id` - UUID del template
- `account_id` - Heredado del header `X-Account-ID`
- `current_occurrence` - Contador de ingresos generados (inicia en 0)
- `is_active` - Estado del template (default: true)
- `exchange_rate` - Calculado según Modo Multi-Currency (si aplica)
- `amount_in_primary_currency` - Calculado según Modo Multi-Currency (si aplica)
- `created_at` - Timestamp de creación

**Response (201):**
```json
{
  "message": "Ingreso recurrente creado exitosamente",
  "recurring_expense": {
    "id": "uuid",
    "account_id": "uuid",
    "description": "Salario Mensual",
    "amount": 500000,
    "currency": "ARS",
    "category_id": "uuid-categoria-salario",
    "category_name": "Salario",
    "family_member_id": null,
    "family_member_name": null,
    "recurrence_frequency": "monthly",
    "recurrence_interval": 1,
    "recurrence_day_of_month": 1,
    "recurrence_day_of_week": null,
    "start_date": "2026-01-01",
    "end_date": null,
    "total_occurrences": null,
    "current_occurrence": 0,
    "exchange_rate": 1.0,
    "amount_in_primary_currency": 500000.0,
    "is_active": true,
    "created_at": "2026-01-18T10:00:00Z"
  }
}
```

**Validaciones:**
- `amount` debe ser > 0
- `currency` debe ser ARS, USD o EUR
- `start_date` debe ser formato YYYY-MM-DD válido
- `end_date` (si existe) debe ser >= `start_date` y formato YYYY-MM-DD
- `recurrence_frequency` debe ser: `daily`, `weekly`, `monthly`, `yearly`
- **monthly/yearly** REQUIERE `recurrence_day_of_month` (1-31)
- **weekly** REQUIERE `recurrence_day_of_week` (0-6)
- **daily** NO debe tener `recurrence_day_of_month` ni `recurrence_day_of_week`
- `recurrence_interval` (si existe) debe ser > 0
- `total_occurrences` (si existe) debe ser > 0
- Si `family_member_id` se proporciona, debe pertenecer a la cuenta
- Si `category_id` se proporciona, debe existir en income_categories

**Multi-Currency - Modos de cálculo:**
1. **Modo 1 (Misma moneda):** `currency == primary_currency` → `exchange_rate = 1.0`
2. **Modo 2 (Tasa manual):** Proporcionás `exchange_rate` → calcula `amount_in_primary_currency`
3. **Modo 3 (Monto real - PREFERIDO):** Proporcionás `amount_in_primary_currency` → calcula `exchange_rate`
4. **Modo Auto:** Busca en tabla `exchange_rates` para `start_date`
   - Si no encuentra, retorna **HTTP 400** pidiendo `exchange_rate` o `amount_in_primary_currency`

**Edge Cases:**
- Día 31 en meses cortos → se genera el último día del mes (ej: 28/29 feb)
- Feb 29 en años no bisiestos → se genera el 28 de febrero

**Errors:**
- `400` - Datos inválidos, formato de fecha incorrecto, validaciones fallidas
- `400` - monthly/yearly requiere recurrence_day_of_month (1-31)
- `400` - weekly requiere recurrence_day_of_week (0=Domingo, 6=Sábado)
- `400` - recurrence_day_of_week solo aplica a frequency=weekly
- `400` - recurrence_day_of_month solo aplica a frequency=monthly/yearly
- `400` - No se encontró tasa de cambio (proporcionar exchange_rate o amount_in_primary_currency)
- `400` - family_member_id no pertenece a esta cuenta

**Use Cases:**
- Salario mensual (monthly, día de cobro)
- Ingresos por alquiler (monthly, día específico)
- Freelance recurrente (weekly/monthly)
- Rentas de inversiones (monthly/yearly)
- Propinas diarias (daily)

---

### GET /recurring-incomes

Listar templates de ingresos recurrentes.

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `is_active` (opcional): `'true'`, `'false'`, `'all'` (default: `'true'`)
- `frequency` (opcional): `'daily'`, `'weekly'`, `'monthly'`, `'yearly'`
- `page` (opcional): número de página (default: 1)
- `limit` (opcional): items por página (default: 20, max: 100)

**Response (200):**
```json
{
  "recurring_incomes": [
    {
      "id": "uuid",
      "description": "Salario Mensual",
      "amount": 500000,
      "currency": "ARS",
      "category_name": "Salario",
      "recurrence_frequency": "monthly",
      "recurrence_interval": 1,
      "recurrence_day_of_month": 1,
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

### GET /recurring-incomes/:id

Obtener detalle de un template.

**Headers:** `Authorization`, `X-Account-ID`

**Response (200):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "description": "Salario Mensual",
  "amount": 500000,
  "currency": "ARS",
  "category_id": "uuid",
  "category_name": "Salario",
  "family_member_id": null,
  "recurrence_frequency": "monthly",
  "recurrence_interval": 1,
  "recurrence_day_of_month": 1,
  "start_date": "2026-01-01",
  "end_date": null,
  "total_occurrences": null,
  "current_occurrence": 3,
  "is_active": true,
  "created_at": "2026-01-01T10:00:00Z",
  "generated_incomes_count": 3
}
```

---

### PUT /recurring-incomes/:id

Actualizar template de ingreso recurrente. **IMPORTANTE:** Actualizar el template NO afecta ingresos ya generados (histórico preservado). Solo afecta FUTUROS ingresos que se generen.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Update Partial - Cambiar monto):**
```json
{
  "amount": 550000
}
```

**Request (Update Completo - Cambiar día y categoría):**
```json
{
  "recurrence_day_of_month": 5,
  "category_id": "uuid-nueva-categoria"
}
```

**Request (Desactivar template):**
```json
{
  "is_active": false
}
```

**Request (Limpiar end_date):**
```json
{
  "end_date": ""
}
```
**Nota:** Enviar `end_date: ""` (string vacío) limpia el campo (lo pone en NULL).

**Request (Limpiar category_id):**
```json
{
  "category_id": ""
}
```
**Nota:** Enviar `category_id: ""` o `family_member_id: ""` limpia el campo (NULL).

**Campos actualizables (todos opcionales):**
- `description` - Nueva descripción (1-200 caracteres)
- `amount` - Nuevo monto (debe ser > 0)
- `currency` - Nueva moneda (ARS | USD | EUR)
- `category_id` - Nueva categoría (UUID válido o "" para limpiar → NULL)
- `family_member_id` - Nuevo miembro familiar (UUID válido o "" para limpiar → NULL)
  - Si se proporciona UUID, debe pertenecer a la cuenta
- `recurrence_interval` - Nuevo intervalo (debe ser > 0)
- `recurrence_day_of_month` - Nuevo día del mes (1-31)
  - Solo válido si `recurrence_frequency` actual es `monthly` o `yearly`
- `recurrence_day_of_week` - Nuevo día de la semana (0-6)
  - Solo válido si `recurrence_frequency` actual es `weekly`
- `end_date` - Nueva fecha fin (YYYY-MM-DD o "" para limpiar → NULL)
  - Debe ser >= `start_date`
- `total_occurrences` - Nuevo límite de repeticiones (debe ser > 0)
- `is_active` - Activar/desactivar template (true | false)
  - `false` = detiene generación de futuros ingresos (soft delete)

**Campos NO modificables:**
- `id` - Identificador único del template (inmutable)
- `account_id` - Cuenta a la que pertenece (inmutable)
- `recurrence_frequency` - Frecuencia (inmutable - cambiar requiere crear nuevo template)
  - No se puede cambiar porque afectaría la lógica del scheduler
- `start_date` - Fecha de inicio (inmutable - histórico)
- `current_occurrence` - Contador automático (inmutable)
- `created_at` - Timestamp de creación (inmutable)

**Validaciones:**
- Al menos un campo actualizable debe ser proporcionado (no se permiten updates vacíos)
- `amount` debe ser > 0
- `currency` debe ser ARS, USD o EUR
- `end_date` debe ser formato YYYY-MM-DD válido y >= `start_date`
- `recurrence_interval` debe ser > 0
- `recurrence_day_of_month` (1-31) solo válido si frequency actual es monthly/yearly
- `recurrence_day_of_week` (0-6) solo válido si frequency actual es weekly
- `total_occurrences` debe ser > 0
- Si `family_member_id` se proporciona (y no es ""), debe pertenecer a la cuenta
- Si `category_id` se proporciona (y no es ""), debe existir en income_categories

**Response (200):**
```json
{
  "message": "Ingreso recurrente actualizado exitosamente",
  "updated_at": "2026-01-21T15:30:00Z",
  "note": "Los ingresos ya generados NO se modifican. Solo afecta futuros ingresos."
}
```

**Errors:**
- `400` - Datos inválidos, formato de fecha incorrecto, validaciones fallidas
- `400` - No hay campos para actualizar
- `400` - recurrence_day_of_month solo aplica a frequency=monthly/yearly
- `400` - recurrence_day_of_week solo aplica a frequency=weekly
- `400` - family_member_id no pertenece a esta cuenta
- `404` - Ingreso recurrente no encontrado

---

### DELETE /recurring-incomes/:id

Eliminar template de ingreso recurrente (soft delete - marca `is_active = false`).

**Headers:** `Authorization`, `X-Account-ID`

**Response (200):**
```json
{
  "message": "Ingreso recurrente eliminado exitosamente (marcado como inactivo)"
}
```

**Nota:** Los ingresos ya generados NO se eliminan (histórico preservado).

---

## 💰 Incomes

Los endpoints de ingresos funcionan idénticamente a expenses.

### POST /incomes

Registrar un ingreso único (one-time). Para ingresos recurrentes (sueldo, alquiler), usar `/recurring-incomes`.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Mínimo - Ingreso simple):**
```json
{
  "description": "Sueldo mensual",
  "amount": 200000,
  "currency": "ARS",
  "date": "2026-01-01"
}
```

**Request (Completo - Con categoría y miembro):**
```json
{
  "description": "Freelance USA",
  "amount": 100,
  "currency": "USD",
  "amount_in_primary_currency": 157500,
  "date": "2026-01-20",
  "category_id": "uuid-categoria-freelance",
  "family_member_id": "uuid-miembro-familia"
}
```

**Campos requeridos:**
- `description` - Descripción del ingreso (1-200 caracteres)
  - Ejemplo: "Sueldo enero", "Freelance proyecto X", "Venta de auto"
- `amount` - Monto del ingreso (debe ser > 0)
  - Ejemplo: 200000 (ARS), 100 (USD)
- `currency` - Moneda del ingreso
  - Valores: `"ARS"` | `"USD"` | `"EUR"`
- `date` - Fecha del ingreso (formato: YYYY-MM-DD)
  - Ejemplo: "2026-01-20"

**Campos opcionales:**
- `category_id` - UUID de la categoría de ingreso (debe existir en income_categories)
  - Si no se proporciona, el ingreso queda sin categoría (null)
- `family_member_id` - UUID del miembro familiar (solo para cuentas tipo "family")
  - Si no se proporciona, el ingreso no está asignado a ningún miembro (null)
  - Si se proporciona, debe pertenecer a la cuenta
- `income_type` - Tipo de ingreso (DEFAULT: `"one-time"`)
  - ⚠️ **NO uses este campo manualmente**. Se usa solo para ingresos generados por recurring_incomes
  - Valores: `"one-time"` | `"recurring"`
- `end_date` - Fecha fin (formato: YYYY-MM-DD)
  - Solo para `income_type: "recurring"` (generado por scheduler)
  - ❌ No se puede usar con `income_type: "one-time"`

**Campos opcionales (Multi-Currency - Modo 3):**
- `exchange_rate` - Tasa de cambio manual (ej: 1575.00)
- `amount_in_primary_currency` - Monto REAL acreditado en moneda primaria
  - **Modo 3 preferido:** Enviás cuántos USD recibiste Y cuántos ARS te acreditaron
  - El sistema calcula automáticamente: `exchange_rate = amount_in_primary_currency / amount`
  - Ejemplo: recibiste USD 100, te acreditaron ARS 157500 → exchange_rate = 1575

**Campos auto-generados:**
- `id` - UUID del ingreso
- `account_id` - Heredado del header `X-Account-ID`
- `exchange_rate` - Calculado automáticamente según Modo Multi-Currency
- `amount_in_primary_currency` - Calculado automáticamente
- `created_at` - Timestamp de creación

**Response (201):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "family_member_id": "uuid-miembro-familia",
  "category_id": "uuid-categoria-freelance",
  "category_name": "Freelance",
  "description": "Freelance USA",
  "amount": 100.00,
  "currency": "USD",
  "exchange_rate": 1575.00,
  "amount_in_primary_currency": 157500.00,
  "income_type": "one-time",
  "date": "2026-01-20",
  "end_date": null,
  "created_at": "2026-01-20T10:00:00Z"
}
```

**Validaciones:**
- `amount` debe ser > 0
- `currency` debe ser ARS, USD o EUR
- `date` debe ser formato YYYY-MM-DD válido
- Si `income_type` es `"one-time"`, NO puede tener `end_date`
- Si `income_type` es `"recurring"` y tiene `end_date`, debe ser >= `date`
- Si `family_member_id` se proporciona, debe pertenecer a la cuenta
- Si `category_id` se proporciona, debe existir en la DB

**Multi-Currency - Modos de cálculo:**
1. **Modo 1 (Misma moneda):** `currency == primary_currency`
   - `exchange_rate = 1.0`
   - `amount_in_primary_currency = amount`

2. **Modo 2 (Tasa manual):** Proporcionás `exchange_rate`
   - `amount_in_primary_currency = amount * exchange_rate`

3. **Modo 3 (Monto real - PREFERIDO):** Proporcionás `amount_in_primary_currency`
   - `exchange_rate = amount_in_primary_currency / amount`
   - **Ejemplo:** USD 100 recibido, ARS 157500 acreditado → rate = 1575

4. **Modo Auto:** Si no proporcionás nada, busca en tabla `exchange_rates`
   - Si no encuentra tasa para esa fecha, retorna **HTTP 400** pidiendo que proporciones `exchange_rate` o `amount_in_primary_currency`

**Errors:**
- `400` - Datos inválidos, formato de fecha incorrecto, validaciones fallidas
- `400` - No se encontró tasa de cambio (proporcionar exchange_rate o amount_in_primary_currency)
- `400` - family_member_id no pertenece a la cuenta

**⚠️ Nota sobre ingresos recurrentes:**
Para ingresos que se repiten regularmente (sueldo, alquiler, pensión, etc.), **NO uses este endpoint**. En su lugar:
1. Usá `POST /recurring-incomes` para crear un **template**
2. El scheduler generará automáticamente los ingresos reales con `income_type: "recurring"`

Ver [docs/MULTI-CURRENCY.md](./docs/MULTI-CURRENCY.md) para más detalles del Modo 3.

---

### GET /incomes

Query params idénticos a expenses:
- `month`, `type`, `category_id`, `family_member_id`, `currency`

---

### GET /incomes/:id

Detalle de ingreso.

---

### PUT /incomes/:id

Actualizar ingreso. Permite actualización parcial (solo enviás los campos que querés cambiar).

**Headers:** `Authorization`, `X-Account-ID`

**Request (Update Partial - Cambiar monto y categoría):**
```json
{
  "amount": 220000,
  "category_id": "uuid-nueva-categoria"
}
```

**Request (Update Solo Descripción):**
```json
{
  "description": "Sueldo enero + bonus"
}
```

**Request (Limpiar end_date):**
```json
{
  "end_date": ""
}
```
**Nota:** Enviar `end_date: ""` (string vacío) limpia el campo (lo pone en NULL). Omitir el campo lo deja sin cambios.

**Campos actualizables (todos opcionales):**
- `description` - Nueva descripción del ingreso (1-200 caracteres)
- `amount` - Nuevo monto (debe ser > 0)
- `currency` - Nueva moneda (ARS | USD | EUR)
  - ⚠️ Si cambiás la moneda, el sistema recalcula `exchange_rate` y `amount_in_primary_currency` automáticamente
- `date` - Nueva fecha del ingreso (formato: YYYY-MM-DD)
  - ⚠️ Si cambiás la fecha, el sistema puede recalcular la tasa de cambio si usa tasas de la DB
- `category_id` - Nueva categoría (UUID válido o null)
- `family_member_id` - Nuevo miembro familiar (UUID válido o null)
  - Si se proporciona, debe pertenecer a la cuenta
- `end_date` - Nueva fecha fin para ingresos recurrentes (formato: YYYY-MM-DD o "" para limpiar)
  - Solo válido si `income_type` es `"recurring"`
  - Debe ser >= `date`
- `exchange_rate` - Nueva tasa de cambio manual (debe ser > 0)
  - Si se proporciona, se usa para recalcular `amount_in_primary_currency`
- `amount_in_primary_currency` - Nuevo monto en moneda primaria (debe ser > 0)
  - Si se proporciona, se usa para recalcular `exchange_rate`

**Campos NO modificables:**
- `id` - Identificador único del ingreso (inmutable)
- `account_id` - Cuenta a la que pertenece (inmutable)
- `income_type` - Tipo de ingreso (inmutable - `"one-time"` o `"recurring"`)
  - No se puede cambiar porque podría violar reglas de negocio
- `recurring_income_id` - Template que generó este ingreso (inmutable)
- `created_at` - Timestamp de creación (inmutable)

**Validaciones:**
- Al menos un campo actualizable debe ser proporcionado (no se permiten updates vacíos)
- `amount` debe ser > 0
- `currency` debe ser ARS, USD o EUR
- `date` debe ser formato YYYY-MM-DD válido
- Si el ingreso es `income_type: "one-time"`, NO puede tener `end_date`
- Si el ingreso es `income_type: "recurring"` y tiene `end_date`, debe ser >= `date`
- Si `family_member_id` se proporciona, debe pertenecer a la cuenta
- Si `category_id` se proporciona, debe existir en la DB
- `exchange_rate` y `amount_in_primary_currency` deben ser > 0 si se proporcionan

**Multi-Currency - Recálculo Automático:**
Si actualizás `amount`, `currency`, o `date`, el sistema recalcula automáticamente la conversión usando:
1. **Modo 1 (Misma moneda):** `currency == primary_currency` → `exchange_rate = 1.0`
2. **Modo 2 (Tasa manual):** Si proporcionás `exchange_rate` → calcula `amount_in_primary_currency`
3. **Modo 3 (Monto real):** Si proporcionás `amount_in_primary_currency` → calcula `exchange_rate`
4. **Modo Auto:** Busca en tabla `exchange_rates` para la nueva fecha
   - Si no encuentra, retorna **HTTP 400** pidiendo que proporciones `exchange_rate` o `amount_in_primary_currency`

**Response (200):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "family_member_id": "uuid-miembro-familia",
  "category_id": "uuid-nueva-categoria",
  "category_name": "Freelance",
  "description": "Sueldo enero + bonus",
  "amount": 220000.00,
  "currency": "ARS",
  "exchange_rate": 1.0,
  "amount_in_primary_currency": 220000.00,
  "income_type": "one-time",
  "date": "2026-01-20",
  "end_date": null,
  "created_at": "2026-01-20T10:00:00Z"
}
```

**Errors:**
- `400` - Datos inválidos, formato de fecha incorrecto, validaciones fallidas
- `400` - No se encontró tasa de cambio (proporcionar exchange_rate o amount_in_primary_currency)
- `400` - No se proporcionaron campos para actualizar
- `400` - family_member_id no pertenece a la cuenta
- `404` - Ingreso no encontrado o no pertenece a la cuenta

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

**Request (Completo):**
```json
{
  "name": "Vacaciones en Brasil",
  "target_amount": 300000,
  "deadline": "2026-06-30",
  "description": "Viaje familiar a la playa",
  "saved_in": "Cuenta de ahorros Banco Galicia"
}
```

**Request (Mínimo):**
```json
{
  "name": "Vacaciones en Brasil",
  "target_amount": 300000
}
```

**Response (201):**
```json
{
  "message": "Meta de ahorro creada exitosamente",
  "savings_goal": {
    "id": "uuid",
    "account_id": "uuid",
    "name": "Vacaciones en Brasil",
    "description": "Viaje familiar a la playa",
    "target_amount": 300000.00,
    "current_amount": 0.00,
    "currency": "ARS",
    "saved_in": "Cuenta de ahorros Banco Galicia",
    "deadline": "2026-06-30",
    "progress_percentage": 0.0,
    "required_monthly_savings": 50000.00,
    "is_active": true,
    "created_at": "2026-01-16T10:00:00Z",
    "updated_at": "2026-01-16T10:00:00Z"
  }
}
```

**Campos requeridos:**
- `name` - Nombre de la meta (1-255 caracteres, único por cuenta)
- `target_amount` - Monto objetivo (debe ser > 0)

**Campos opcionales:**
- `description` - Descripción de la meta
- `deadline` - Fecha límite (YYYY-MM-DD, debe ser futura)
- `saved_in` - Dónde se guarda el dinero físicamente (ej: "Cuenta Banco X", "Alcancía")

**Campos auto-generados:**
- `currency` - Hereda la moneda de la cuenta (ARS/USD)
- `current_amount` - Siempre inicia en 0
- `is_active` - Siempre inicia en `true`
- `progress_percentage` - Siempre inicia en 0
- `required_monthly_savings` - Auto-calculado si hay deadline. Fórmula: `(target_amount - current_amount) / meses_restantes`. Retorna `null` si no hay deadline.

**Validaciones:**
- El nombre debe ser único entre metas activas de la misma cuenta (case-insensitive)
- Si se proporciona deadline, debe ser fecha futura
- La moneda se hereda automáticamente de la cuenta (no se puede especificar)

**Errors:**
- `400` - Datos inválidos (ej: deadline en el pasado, target_amount ≤ 0)
- `409` - Ya existe una meta activa con ese nombre

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

### PUT /savings-goals/:id

Actualizar meta de ahorro.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Partial Update):**
```json
{
  "name": "Nuevo nombre de la meta",
  "target_amount": 500000,
  "deadline": "2026-12-31",
  "description": "Meta actualizada",
  "is_active": true
}
```

**Request (Clear deadline):**
```json
{
  "deadline": ""
}
```

**Response (200):**
```json
{
  "message": "Meta de ahorro actualizada exitosamente",
  "savings_goal": {
    "id": "uuid",
    "account_id": "uuid",
    "name": "Nuevo nombre de la meta",
    "description": "Meta actualizada",
    "target_amount": 500000.00,
    "current_amount": 80000.00,
    "currency": "ARS",
    "saved_in": null,
    "deadline": "2026-12-31",
    "is_active": true,
    "progress_percentage": 16.0,
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-01-21T14:30:00Z"
  }
}
```

**Campos actualizables (todos opcionales):**
- `name` - Nombre de la meta (1-255 caracteres)
  - Debe ser único por cuenta (case-insensitive)
- `description` - Descripción de la meta
- `target_amount` - Monto objetivo (debe ser > 0)
- `saved_in` - Dónde se guarda el dinero (ej: "Cuenta de ahorros Banco X")
- `deadline` - Fecha límite (YYYY-MM-DD)
  - Debe ser fecha futura
  - Enviar string vacío `""` para limpiar el deadline
- `is_active` - Estado de la meta (true/false)
  - `false` = archivada (deja de aparecer en listados activos)

**Validaciones:**
- Partial update: Solo los campos enviados se actualizan
- El nombre debe ser único entre metas activas de la misma cuenta
- El `current_amount` NO se puede modificar (usar add-funds/withdraw-funds)
- La `currency` NO se puede modificar
- Si se actualiza `target_amount`, recalcula automáticamente `progress_percentage`

**Errors:**
- `400` - Datos inválidos (ej: deadline en el pasado, target_amount ≤ 0)
- `404` - Meta no encontrada
- `409` - Ya existe otra meta activa con ese nombre

**Nota:** Esta operación NO afecta las transacciones (add-funds/withdraw-funds) ya realizadas.

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
- `date` - **Opcional**, formato YYYY-MM-DD (default: fecha actual)
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
- `date` - **Opcional**, formato YYYY-MM-DD (default: fecha actual)
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

Crear categoría custom de gastos.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Mínimo):**
```json
{
  "name": "Veterinario"
}
```

**Request (Completo):**
```json
{
  "name": "Veterinario",
  "icon": "🐕",
  "color": "#FF5733"
}
```

**Campos requeridos:**
- `name` - Nombre de la categoría
  - Debe ser único por cuenta (case-insensitive)
  - Ejemplo: "Alimentación" y "alimentación" son duplicados
  - "Alimentación" en Cuenta A puede coexistir con "Alimentación" en Cuenta B

**Campos opcionales:**
- `icon` - Emoji representativo (ej: "🐕", "🏥", "🎮")
  - Si no se proporciona, se guarda como NULL
- `color` - Color en formato hexadecimal (ej: "#FF5733", "#4CAF50")
  - Si no se proporciona, se guarda como NULL

**Campos auto-generados:**
- `id` - UUID de la categoría
- `account_id` - Heredado del header `X-Account-ID`
- `is_system` - Siempre `false` para categorías custom
- `created_at` - Timestamp de creación

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

**Validaciones:**
- `name` es requerido (no puede estar vacío)
- `name` debe ser único por cuenta (comparación case-insensitive)
- `icon` (si se proporciona) debe ser un emoji válido
- `color` (si se proporciona) debe ser formato hexadecimal válido (ej: "#FF5733")

**Errors:**
- `400` - Datos inválidos, name vacío o formato incorrecto
- `409` - Ya existe una categoría con ese nombre en esta cuenta

**Restrictions:**
- No se pueden editar/borrar categorías del sistema (`is_system = true`)
- No se pueden borrar categorías custom con gastos asociados
- Nombres únicos por cuenta (sin importar mayúsculas/minúsculas)

---

### PUT /expense-categories/:id

Actualizar categoría custom de gastos.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Update name):**
```json
{
  "name": "Veterinaria y Mascotas"
}
```

**Request (Update icon y color):**
```json
{
  "icon": "🐾",
  "color": "#8BC34A"
}
```

**Campos actualizables (todos opcionales):**
- `name` - Nuevo nombre (debe ser único por cuenta, case-insensitive)
- `icon` - Nuevo emoji
- `color` - Nuevo color hexadecimal

**Campos NO modificables:**
- `id` - Identificador único (inmutable)
- `account_id` - Cuenta a la que pertenece (inmutable)
- `is_system` - Flag de sistema (inmutable)
- `created_at` - Timestamp de creación (inmutable)

**Validaciones:**
- Solo se pueden editar categorías custom (`is_system = false`)
- La categoría debe pertenecer a la cuenta del header `X-Account-ID`
- `name` (si se proporciona) debe ser único por cuenta (case-insensitive)

**Response (200):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "name": "Veterinaria y Mascotas",
  "icon": "🐾",
  "color": "#8BC34A",
  "is_system": false,
  "created_at": "2026-01-19T01:30:00Z"
}
```

**Errors:**
- `400` - Datos inválidos, formato incorrecto
- `403` - No se pueden editar categorías del sistema
- `403` - La categoría no pertenece a esta cuenta
- `404` - Categoría no encontrada

---

### DELETE /expense-categories/:id

Eliminar categoría custom de gastos.

**Headers:** `Authorization`, `X-Account-ID`

**Validaciones:**
- Solo se pueden eliminar categorías custom (`is_system = false`)
- La categoría debe pertenecer a la cuenta del header `X-Account-ID`
- La categoría NO debe tener gastos asociados

**Response (200):**
```json
{
  "message": "category deleted successfully",
  "id": "uuid"
}
```

**Errors:**
- `400` - category_id es requerido
- `403` - No se pueden eliminar categorías del sistema
- `403` - La categoría no pertenece a esta cuenta
- `404` - Categoría no encontrada
- `409` - No se puede eliminar categoría con gastos asociados
  ```json
  {
    "error": "cannot delete category with associated expenses",
    "expense_count": 15
  }
  ```

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

**Request (Mínimo):**
```json
{
  "name": "Bonus Anual"
}
```

**Request (Completo):**
```json
{
  "name": "Bonus Anual",
  "icon": "💎",
  "color": "#4CAF50"
}
```

**Campos requeridos:**
- `name` - Nombre de la categoría
  - Debe ser único por cuenta (case-insensitive)
  - Ejemplo: "Salario" y "salario" son duplicados
  - "Salario" en Cuenta A puede coexistir con "Salario" en Cuenta B

**Campos opcionales:**
- `icon` - Emoji representativo (ej: "💎", "💼", "📈")
  - Si no se proporciona, se guarda como NULL
- `color` - Color en formato hexadecimal (ej: "#4CAF50", "#66BB6A")
  - Si no se proporciona, se guarda como NULL

**Campos auto-generados:**
- `id` - UUID de la categoría
- `account_id` - Heredado del header `X-Account-ID`
- `is_system` - Siempre `false` para categorías custom
- `created_at` - Timestamp de creación

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

**Validaciones:**
- `name` es requerido (no puede estar vacío)
- `name` debe ser único por cuenta (comparación case-insensitive)
- `icon` (si se proporciona) debe ser un emoji válido
- `color` (si se proporciona) debe ser formato hexadecimal válido (ej: "#4CAF50")

**Errors:**
- `400` - Datos inválidos, name vacío o formato incorrecto
- `409` - Ya existe una categoría con ese nombre en esta cuenta

**Restrictions:**
- No se pueden editar/borrar categorías del sistema (`is_system = true`)
- No se pueden borrar categorías custom con ingresos asociados
- Nombres únicos por cuenta (sin importar mayúsculas/minúsculas)

---

### PUT /income-categories/:id

Actualizar categoría custom de ingresos.

**Headers:** `Authorization`, `X-Account-ID`

**Request (Update name):**
```json
{
  "name": "Bonus y Comisiones"
}
```

**Request (Update icon y color):**
```json
{
  "icon": "💰",
  "color": "#9CCC65"
}
```

**Campos actualizables (todos opcionales):**
- `name` - Nuevo nombre (debe ser único por cuenta, case-insensitive)
- `icon` - Nuevo emoji
- `color` - Nuevo color hexadecimal

**Campos NO modificables:**
- `id` - Identificador único (inmutable)
- `account_id` - Cuenta a la que pertenece (inmutable)
- `is_system` - Flag de sistema (inmutable)
- `created_at` - Timestamp de creación (inmutable)

**Validaciones:**
- Solo se pueden editar categorías custom (`is_system = false`)
- La categoría debe pertenecer a la cuenta del header `X-Account-ID`
- `name` (si se proporciona) debe ser único por cuenta (case-insensitive)

**Response (200):**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "name": "Bonus y Comisiones",
  "icon": "💰",
  "color": "#9CCC65",
  "is_system": false,
  "created_at": "2026-01-19T01:30:00Z"
}
```

**Errors:**
- `400` - Datos inválidos, formato incorrecto
- `403` - No se pueden editar categorías del sistema
- `403` - La categoría no pertenece a esta cuenta
- `404` - Categoría no encontrada

---

### DELETE /income-categories/:id

Eliminar categoría custom de ingresos.

**Headers:** `Authorization`, `X-Account-ID`

**Validaciones:**
- Solo se pueden eliminar categorías custom (`is_system = false`)
- La categoría debe pertenecer a la cuenta del header `X-Account-ID`
- La categoría NO debe tener ingresos asociados

**Response (200):**
```json
{
  "message": "category deleted successfully",
  "id": "uuid"
}
```

**Errors:**
- `400` - category_id es requerido
- `403` - No se pueden eliminar categorías del sistema
- `403` - La categoría no pertenece a esta cuenta
- `404` - Categoría no encontrada
- `409` - No se puede eliminar categoría con ingresos asociados
  ```json
  {
    "error": "cannot delete category with associated incomes",
    "income_count": 8
  }
  ```

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
**Última actualización:** 2026-01-21 (Documentados todos los endpoints de Savings Goals + campos opcionales)
**Versión:** 2.3 (Consolidada)  
**Mantenido por:** Gentleman Programming & Lorenzo
