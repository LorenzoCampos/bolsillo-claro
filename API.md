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
EUR - Euro
```

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
- Family requiere ≥1 miembro
- Personal no puede tener miembros
- Auto-crea meta "Ahorro General"

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
      "amount": 25000,
      "date": "2026-01-10"
    }
  ],
  "recent_transactions": [
    {
      "id": "uuid",
      "type": "expense",
      "description": "Supermercado",
      "amount": 25000,
      "date": "2026-01-10"
    }
  ]
}
```

**Cálculo:**
```
available_balance = total_income - total_expenses - total_assigned_to_goals
```

**Notas:**
- Todos los montos en moneda primaria
- `top_expenses`: Máximo 5
- `recent_transactions`: Máximo 10 (expenses + incomes mezclados)

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
  "is_general": false,
  "is_active": true,
  "progress_percentage": 0.0,
  "required_monthly_savings": 50000.00,
  "created_at": "2026-01-16T10:00:00Z"
}
```

**Fields:**
- `deadline` - Opcional (null = sin deadline)
- `is_general` - Auto-false (solo 1 meta general por cuenta)
- `required_monthly_savings` - Solo si tiene deadline

---

### GET /savings-goals

Listar metas.

**Headers:** `Authorization`, `X-Account-ID`

**Query Params:**
- `is_active` (opcional): `true` / `false` (default: `true`)

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
      "deadline": "2026-06-30"
    }
  ],
  "count": 1
}
```

---

### GET /savings-goals/:id

Detalle con historial de transacciones.

**Headers:** `Authorization`, `X-Account-ID`

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
      "type": "add",
      "amount": 30000,
      "description": "Ahorro enero",
      "date": "2026-01-15",
      "created_at": "2026-01-15T10:00:00Z"
    },
    {
      "id": "uuid",
      "type": "add",
      "amount": 20000,
      "date": "2026-01-20"
    }
  ]
}
```

---

### POST /savings-goals/:id/add-funds

Agregar fondos a meta.

**Headers:** `Authorization`, `X-Account-ID`

**Request:**
```json
{
  "amount": 30000,
  "description": "Ahorro enero"
}
```

**Response (200):**
```json
{
  "message": "Fondos agregados exitosamente",
  "transaction": {
    "id": "uuid",
    "type": "add",
    "amount": 30000
  },
  "new_current_amount": 80000.00
}
```

**Effect:**
- Actualiza `current_amount` automáticamente
- Se cuenta en `total_assigned_to_goals` del dashboard

---

### POST /savings-goals/:id/withdraw-funds

Retirar fondos de meta.

**Headers:** `Authorization`, `X-Account-ID`

**Request:**
```json
{
  "amount": 10000,
  "description": "Adelanto para pasaje"
}
```

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
  "name": "Veterinario",
  "icon": "🐕",
  "color": "#FF5733",
  "is_custom": true
}
```

**Restrictions:**
- No se pueden editar/borrar predefinidas
- No se pueden borrar custom con expenses asociados

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

Crear categoría custom de ingresos (misma estructura que expense-categories).

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
| `El email ya está registrado` | Email duplicado | Usar otro email o login |

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
**Última actualización:** 2026-01-16  
**Versión:** 2.0 (Consolidada)  
**Mantenido por:** Gentleman Programming & Lorenzo
