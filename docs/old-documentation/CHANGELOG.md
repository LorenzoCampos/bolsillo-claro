# Changelog - Bolsillo Claro

## [MVP v1.0] - En Desarrollo

### Decisiones de Arquitectura y Cambios

#### 2026-01-16 - Sistema Completo de Recurrencia (EN PROGRESO)

**Feature:** Implementación de sistema avanzado de recurrencia para gastos

**Motivación:**
- Casos de uso reales: alquileres mensuales, servicios periódicos, compras en cuotas
- Ejemplo: Zapatillas en 6 cuotas de $8,000 c/u

**Diseño Técnico:**
- Ver documento completo: `/docs/RECURRENCE-SYSTEM-DESIGN.md`

**Nuevos Campos en `expenses` table:**
1. `recurrence_frequency` - Frecuencia: daily, weekly, monthly, yearly
2. `recurrence_interval` - Cada cuántos períodos (ej: cada 2 semanas)
3. `recurrence_day_of_month` - Día del mes (1-31) para frecuencia mensual/anual
4. `recurrence_day_of_week` - Día de semana (0-6) para frecuencia semanal
5. `total_occurrences` - Cantidad total de repeticiones (NULL = infinito)
6. `current_occurrence` - Número de ocurrencia actual (para mostrar "3/6")
7. `parent_expense_id` - ID del gasto padre (para gastos auto-generados)

**Casos de Uso Soportados:**
- ✅ Alquiler mensual sin fin (ej: día 5 de cada mes)
- ✅ Compras en cuotas (ej: 6 cuotas mensuales)
- ✅ Suscripciones anuales (ej: Netflix cada año)
- ✅ Gastos semanales (ej: gym todos los lunes)
- ✅ Gastos diarios (ej: café todas las mañanas)

**Estado:** 📝 Diseño completo → 🚧 Implementación en progreso

---

#### 2026-01-13 - Definición MVP Final

**Decisiones tomadas:**
1. **Wishlist removida del MVP** - Se pospone para v1.1
2. **Multi-currency con snapshot histórico** - Implementación semi-automática
3. **Exchange rates manuales/semi-automáticos** - Admin carga 1 vez por día
4. **Savings Goals integradas en balance** - Descuento virtual (no crea expenses reales)
5. **Onboarding de primera cuenta** - Manejado por frontend, backend provee `has_accounts` flag

**Alcance MVP v1.0:**
- ✅ Autenticación (JWT)
- ✅ Cuentas (CRUD completo)
- ✅ Gastos (CRUD + filtros + multi-currency)
- ✅ Ingresos (CRUD + filtros + multi-currency)
- ✅ Categorías (predefinidas + custom)
- ⏳ Dashboard básico (balance, gastos por categoría, top gastos)
- ⏳ Exchange Rates (manual/semi-automático)
- ⏳ Savings Goals (CRUD + add/withdraw funds)

**Pospuesto para v1.1:**
- ❌ Wishlist vinculada a metas
- ❌ Dashboard con tendencias (6 meses)
- ❌ API externa de exchange rates
- ❌ Account settings (theme, language)
- ❌ Notificaciones
- ❌ Budgets (presupuestos)
- ❌ Exports (CSV/Excel)

---

## [Fase 3] - 2026-01-13 - Categorías Completadas

### Implementado
- ✅ Tabla `expense_categories` (15 predefinidas + custom por cuenta)
- ✅ Tabla `income_categories` (10 predefinidas + custom por cuenta)
- ✅ CRUD completo de categorías custom
- ✅ Migración de expenses/incomes: columna `category` TEXT → `category_id` UUID
- ✅ Datos existentes migrados con JOIN a categorías predefinidas
- ✅ Responses incluyen `category_id` + `category_name` para facilitar frontend

### Categorías Predefinidas

**Expense Categories (15):**
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

**Income Categories (10):**
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

### Decisiones Técnicas
- Categorías system tienen `account_id = NULL` e `is_system = TRUE`
- Categorías custom tienen `account_id = <uuid>` e `is_system = FALSE`
- No se pueden editar/borrar categorías system
- No se pueden borrar categorías custom que tengan expenses/incomes asociados
- Unique constraint: nombre único por scope (global para system, por cuenta para custom)

---

## [Fase 2] - 2026-01-13 - CRUD Expenses Completado

### Implementado
- ✅ POST /api/expenses (crear one-time o recurring)
- ✅ GET /api/expenses (listar con filtros: fecha, tipo, categoría, miembro, paginación)
- ✅ GET /api/expenses/:id (detalle individual)
- ✅ PUT /api/expenses/:id (actualización parcial)
- ✅ DELETE /api/expenses/:id (eliminación)

### Validaciones
- Expense type: `one-time` no puede tener `end_date`, `recurring` puede tenerlo (opcional)
- Fechas: formato YYYY-MM-DD, end_date >= date
- Family members: validación de que pertenezcan a la cuenta
- Ownership: solo puedes ver/modificar tus propios gastos

---

## [Fase 2] - 2026-01-13 - CRUD Incomes Completado

### Implementado
- ✅ POST /api/incomes (crear one-time o recurring)
- ✅ GET /api/incomes (listar con filtros idénticos a expenses)
- ✅ GET /api/incomes/:id (detalle individual)
- ✅ PUT /api/incomes/:id (actualización parcial)
- ✅ DELETE /api/incomes/:id (eliminación)

### Estructura idéntica a Expenses
- Misma lógica de tipos (one-time/recurring)
- Misma lógica de end_date opcional
- Mismos filtros y paginación

---

## [Fase 1] - 2026-01-13 - Foundation

### Implementado
- ✅ Autenticación con JWT (access + refresh tokens)
- ✅ Bcrypt para passwords (cost factor 12)
- ✅ Cuentas: POST /api/accounts (personal + family)
- ✅ Cuentas: GET /api/accounts (listar)
- ✅ Middleware: AuthMiddleware (JWT validation)
- ✅ Middleware: AccountMiddleware (X-Account-ID validation)

### Base de Datos
- ✅ users (id, email, password_hash)
- ✅ accounts (id, user_id, name, account_type, currency)
- ✅ family_members (id, account_id, name, role)
- ✅ savings_goals (id, account_id, name, target_amount, current_amount)
- ✅ expenses (id, account_id, family_member_id, category_id, description, amount, currency, expense_type, date, end_date)
- ✅ incomes (id, account_id, family_member_id, category_id, description, amount, currency, income_type, date, end_date)
- ✅ expense_categories (id, account_id, name, icon, color, is_system)
- ✅ income_categories (id, account_id, name, icon, color, is_system)

### Decisiones de Arquitectura
- **Users vs Accounts:** Separados para permitir múltiples contextos financieros por usuario
- **Account Types:** `personal` (individual) y `family` (con múltiples miembros)
- **Family Members:** Solo existen para cuentas tipo `family`, permiten asignar gastos/ingresos a personas específicas
- **Currency:** Enum con ARS, USD, EUR (extensible)
- **Expense/Income Types:** `one-time` (gasto único) y `recurring` (recurrente como suscripciones)

---

## Roadmap

### v1.0 MVP (En curso - ~7-11 horas restantes)
1. ⏳ CRUD completo de Accounts
2. ⏳ Multi-currency con snapshot histórico
3. ⏳ Dashboard básico
4. ⏳ Savings Goals CRUD

### v1.1 (Futuro)
- Wishlist vinculada a metas
- Dashboard con tendencias
- API externa de exchange rates
- Account settings

### v2.0 (Futuro lejano)
- Budgets (presupuestos)
- Notificaciones
- Exports
- Reports avanzados
- Mobile app
