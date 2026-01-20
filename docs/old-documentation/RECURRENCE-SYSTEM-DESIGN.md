# 🔄 Sistema de Recurrencia - Diseño Técnico

**Versión:** 1.0  
**Fecha:** 2026-01-16  
**Autor:** Gentleman Programming & Lorenzo  
**Status:** 📝 En Diseño

---

## 📋 Tabla de Contenidos

- [Objetivo](#objetivo)
- [Casos de Uso](#casos-de-uso)
- [Diseño de Base de Datos](#diseño-de-base-de-datos)
- [Lógica de Negocio](#lógica-de-negocio)
- [API Changes](#api-changes)
- [Frontend Changes](#frontend-changes)
- [Ejemplos de Uso](#ejemplos-de-uso)
- [Consideraciones Técnicas](#consideraciones-técnicas)

---

## 🎯 Objetivo

Implementar un sistema completo de recurrencia para gastos que permita:

1. **Gastos periódicos automáticos** (alquiler, suscripciones, servicios)
2. **Compras en cuotas** (6 cuotas de $8000)
3. **Gastos con frecuencias personalizables** (diario, semanal, mensual, anual)
4. **Límite de ocurrencias** (finito o infinito)

---

## 💡 Casos de Uso

### **Caso 1: Alquiler Mensual**
- **Descripción:** Pagar alquiler cada mes, día 5
- **Configuración:**
  - `expense_type: 'recurring'`
  - `recurrence_frequency: 'monthly'`
  - `recurrence_day_of_month: 5`
  - `recurrence_interval: 1`
  - `total_occurrences: null` (infinito)
  - `date: '2026-01-05'`
  - `end_date: null`

### **Caso 2: Zapatillas en 6 Cuotas**
- **Descripción:** Compra de $48,000 en 6 cuotas de $8,000
- **Configuración:**
  - `expense_type: 'recurring'`
  - `recurrence_frequency: 'monthly'`
  - `recurrence_day_of_month: 16`
  - `recurrence_interval: 1`
  - `total_occurrences: 6`
  - `current_occurrence: 1`
  - `date: '2026-01-16'`
  - `end_date: '2026-06-16'` (calculado automáticamente)

### **Caso 3: Gym 3 veces por semana**
- **Descripción:** Pagar gym cada lunes, miércoles y viernes
- **Configuración (alternativa):**
  - Crear 3 gastos recurrentes separados, uno por cada día
  - O agregar campo `recurrence_days_of_week: [1, 3, 5]` (feature futura)

### **Caso 4: Suscripción Netflix Anual**
- **Descripción:** Pagar Netflix una vez al año
- **Configuración:**
  - `expense_type: 'recurring'`
  - `recurrence_frequency: 'yearly'`
  - `recurrence_day_of_month: 15`
  - `recurrence_interval: 1`
  - `total_occurrences: null`

---

## 🗄️ Diseño de Base de Datos

### **Migración: Agregar Columnas de Recurrencia**

```sql
-- Archivo: backend/migrations/012_add_recurrence_fields.sql

-- Agregar campos de recurrencia a expenses table
ALTER TABLE expenses
  ADD COLUMN recurrence_frequency TEXT 
    CHECK (recurrence_frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
  ADD COLUMN recurrence_interval INT DEFAULT 1 
    CHECK (recurrence_interval > 0),
  ADD COLUMN recurrence_day_of_month INT 
    CHECK (recurrence_day_of_month BETWEEN 1 AND 31),
  ADD COLUMN recurrence_day_of_week INT 
    CHECK (recurrence_day_of_week BETWEEN 0 AND 6),  -- 0=Domingo, 6=Sábado
  ADD COLUMN total_occurrences INT 
    CHECK (total_occurrences > 0 OR total_occurrences IS NULL),
  ADD COLUMN current_occurrence INT DEFAULT 1 
    CHECK (current_occurrence > 0),
  ADD COLUMN parent_expense_id UUID 
    REFERENCES expenses(id) ON DELETE CASCADE;

-- Índices para mejor performance
CREATE INDEX idx_expenses_recurrence_frequency ON expenses(recurrence_frequency);
CREATE INDEX idx_expenses_parent_expense_id ON expenses(parent_expense_id);

-- Constraint: Si es recurrente, debe tener frecuencia
ALTER TABLE expenses
  ADD CONSTRAINT check_recurring_has_frequency 
  CHECK (
    (expense_type = 'one-time' AND recurrence_frequency IS NULL) OR
    (expense_type = 'recurring' AND recurrence_frequency IS NOT NULL)
  );

-- Constraint: Si es mensual, debe tener día del mes
ALTER TABLE expenses
  ADD CONSTRAINT check_monthly_has_day 
  CHECK (
    (recurrence_frequency != 'monthly') OR
    (recurrence_frequency = 'monthly' AND recurrence_day_of_month IS NOT NULL)
  );

-- Constraint: Si es semanal, debe tener día de la semana
ALTER TABLE expenses
  ADD CONSTRAINT check_weekly_has_day 
  CHECK (
    (recurrence_frequency != 'weekly') OR
    (recurrence_frequency = 'weekly' AND recurrence_day_of_week IS NOT NULL)
  );

-- Constraint: current_occurrence no puede exceder total_occurrences
ALTER TABLE expenses
  ADD CONSTRAINT check_current_within_total 
  CHECK (
    total_occurrences IS NULL OR 
    current_occurrence <= total_occurrences
  );

COMMENT ON COLUMN expenses.recurrence_frequency IS 'Frecuencia de repetición: daily, weekly, monthly, yearly';
COMMENT ON COLUMN expenses.recurrence_interval IS 'Cada cuántos períodos se repite (ej: cada 2 semanas = interval:2)';
COMMENT ON COLUMN expenses.recurrence_day_of_month IS 'Día del mes para recurrencia mensual/anual (1-31)';
COMMENT ON COLUMN expenses.recurrence_day_of_week IS 'Día de la semana para recurrencia semanal (0=Domingo, 6=Sábado)';
COMMENT ON COLUMN expenses.total_occurrences IS 'Cantidad total de repeticiones. NULL = infinito';
COMMENT ON COLUMN expenses.current_occurrence IS 'Número de ocurrencia actual (para cuotas: 1/6, 2/6, etc.)';
COMMENT ON COLUMN expenses.parent_expense_id IS 'ID del gasto padre (para gastos generados automáticamente)';
```

### **Nuevos Campos Explicados:**

| Campo | Tipo | Descripción | Ejemplo |
|-------|------|-------------|---------|
| `recurrence_frequency` | TEXT | Frecuencia: `daily`, `weekly`, `monthly`, `yearly` | `'monthly'` |
| `recurrence_interval` | INT | Cada cuántos períodos (cada 2 semanas = 2) | `1` |
| `recurrence_day_of_month` | INT | Día del mes (1-31) | `5` (día 5) |
| `recurrence_day_of_week` | INT | Día de semana (0=Dom, 6=Sáb) | `1` (Lunes) |
| `total_occurrences` | INT | Total de repeticiones (NULL = infinito) | `6` o `NULL` |
| `current_occurrence` | INT | Ocurrencia actual (para mostrar "3/6") | `1` |
| `parent_expense_id` | UUID | ID del gasto padre (si fue auto-generado) | `uuid` o `NULL` |

---

## 🔧 Lógica de Negocio

### **Reglas de Validación**

1. **Gastos one-time:**
   - `recurrence_frequency` debe ser `NULL`
   - No puede tener `total_occurrences`
   - No puede tener `parent_expense_id`

2. **Gastos recurring:**
   - `recurrence_frequency` es **REQUERIDO**
   - Si `frequency = 'monthly'` → `recurrence_day_of_month` es **REQUERIDO**
   - Si `frequency = 'weekly'` → `recurrence_day_of_week` es **REQUERIDO**
   - `recurrence_interval` default = 1
   - `current_occurrence` default = 1

3. **Cuotas (caso especial):**
   - `total_occurrences` debe ser mayor a 0
   - `current_occurrence` debe estar entre 1 y `total_occurrences`
   - `end_date` se calcula automáticamente

### **Cálculo de Próximas Ocurrencias**

```go
// Pseudo-código para calcular próxima fecha
func CalculateNextOccurrence(expense Expense) time.Time {
    switch expense.RecurrenceFrequency {
    case "daily":
        return expense.Date.AddDays(expense.RecurrenceInterval)
    
    case "weekly":
        return expense.Date.AddWeeks(expense.RecurrenceInterval)
    
    case "monthly":
        // Agregar meses y ajustar al día especificado
        nextMonth := expense.Date.AddMonths(expense.RecurrenceInterval)
        return SetDayOfMonth(nextMonth, expense.RecurrenceDayOfMonth)
    
    case "yearly":
        return expense.Date.AddYears(expense.RecurrenceInterval)
    }
}
```

### **Generación Automática de Gastos Futuros**

**Opción A: On-Demand (recomendada para MVP)**
- No se crean gastos futuros automáticamente
- Se calculan "virtualmente" al listar
- Endpoint: `GET /expenses?include_future=true&months=3`

**Opción B: CRON Job (feature futura)**
- Job diario que crea gastos del mes siguiente
- Mejora performance
- Requiere más infraestructura

---

## 📡 API Changes

### **POST /expenses - Request Body (NUEVO)**

```json
{
  "category_id": "uuid (opcional)",
  "amount": 100.00,
  "currency": "ARS",
  "expense_type": "recurring",
  "description": "Alquiler Depto",
  "date": "2026-01-05",
  
  // ⭐ NUEVOS CAMPOS DE RECURRENCIA
  "recurrence_frequency": "monthly",      // "daily" | "weekly" | "monthly" | "yearly"
  "recurrence_interval": 1,               // Cada cuántos (default: 1)
  "recurrence_day_of_month": 5,           // Requerido si frequency='monthly' o 'yearly'
  "recurrence_day_of_week": null,         // Requerido si frequency='weekly' (0=Dom, 6=Sáb)
  "total_occurrences": null,              // NULL = infinito, 6 = 6 cuotas
  "end_date": null                        // Calculado automáticamente si total_occurrences está definido
}
```

### **Response - Expense Object (NUEVO)**

```json
{
  "id": "uuid",
  "account_id": "uuid",
  "category_id": "uuid",
  "category_name": "Vivienda",
  "description": "Alquiler Depto",
  "amount": 50000.00,
  "currency": "ARS",
  "exchange_rate": 1,
  "amount_in_primary_currency": 50000.00,
  "expense_type": "recurring",
  "date": "2026-01-05",
  "end_date": null,
  
  // ⭐ NUEVOS CAMPOS
  "recurrence_frequency": "monthly",
  "recurrence_interval": 1,
  "recurrence_day_of_month": 5,
  "recurrence_day_of_week": null,
  "total_occurrences": null,
  "current_occurrence": 1,
  "parent_expense_id": null,
  
  "created_at": "2026-01-16T12:00:00Z",
  "updated_at": "2026-01-16T12:00:00Z"
}
```

---

## 🎨 Frontend Changes

### **1. Actualizar Types**

```typescript
// frontend/src/types/expense.ts

export const ExpenseSchema = z.object({
  // ... campos existentes ...
  
  // Nuevos campos de recurrencia
  recurrence_frequency: z.enum(['daily', 'weekly', 'monthly', 'yearly']).optional().nullable(),
  recurrence_interval: z.number().int().positive().optional(),
  recurrence_day_of_month: z.number().int().min(1).max(31).optional().nullable(),
  recurrence_day_of_week: z.number().int().min(0).max(6).optional().nullable(),
  total_occurrences: z.number().int().positive().optional().nullable(),
  current_occurrence: z.number().int().positive().optional(),
  parent_expense_id: z.string().uuid().optional().nullable(),
});
```

### **2. Mejorar ExpenseForm**

Campos condicionales:

- **Si `expense_type = 'recurring'`:**
  - Mostrar selector de frecuencia (Diario, Semanal, Mensual, Anual)
  - Mostrar "Cada cuántos" (interval)
  
- **Si `frequency = 'monthly'` o `'yearly'`:**
  - Mostrar selector de día del mes (1-31)
  
- **Si `frequency = 'weekly'`:**
  - Mostrar selector de día de semana (Dom-Sáb)
  
- **Mostrar opciones de fin:**
  - Radio: "Sin fin" / "Cantidad de repeticiones"
  - Si "Cantidad": input numérico (ej: 6 cuotas)

### **3. Mejorar Visualización en Lista**

Mostrar:
- Badge con frecuencia: "📅 Mensual" / "📅 Cada 2 semanas"
- Si tiene cuotas: "Cuota 3/6"
- Próxima fecha (calculada)

---

## 📚 Ejemplos de Uso

### **Ejemplo 1: Crear Alquiler Mensual**

```bash
POST /expenses
{
  "description": "Alquiler Depto Palermo",
  "amount": 80000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-02-05",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 5,
  "total_occurrences": null
}
```

**Resultado:**
- Se crea 1 gasto el 2026-02-05
- Aparecerá automáticamente todos los meses día 5
- Sin fecha de fin

### **Ejemplo 2: Compra en 6 Cuotas**

```bash
POST /expenses
{
  "description": "Zapatillas Nike - Cuota 1/6",
  "amount": 8000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-16",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 16,
  "total_occurrences": 6,
  "current_occurrence": 1
}
```

**Resultado:**
- Se crea 1 gasto hoy
- Se calculan automáticamente las próximas 5 cuotas
- `end_date` = 2026-06-16 (calculado)

---

## ⚠️ Consideraciones Técnicas

### **Performance**

- **Índices:** Agregar índices en `recurrence_frequency` y `parent_expense_id`
- **Paginación:** Mantener límite de 20 items por página
- **Cache:** React Query cache de 5 minutos para lista de expenses

### **Edge Cases**

1. **Día 31 en meses de 30 días:**
   - Ajustar al último día del mes (30 o 28/29)
   
2. **Cambio de mes con diferente cantidad de días:**
   - Ejemplo: Recurrencia día 31, próximo mes tiene 30 días
   - Solución: Usar día 30 para ese mes

3. **Año bisiesto:**
   - Manejar correctamente 29 de febrero

4. **Timezone:**
   - Todas las fechas en UTC
   - Conversión a local en frontend

### **Migraciones y Retrocompatibilidad**

- Gastos existentes (`one-time` y `recurring` viejos):
  - `recurrence_frequency = NULL`
  - Funciona normalmente
  
- No se requiere migración de datos existentes

---

## 🚀 Plan de Implementación

### **Fase 1: Base de Datos** ✅
- [ ] Crear migración `012_add_recurrence_fields.sql`
- [ ] Aplicar migración en desarrollo
- [ ] Testear constraints

### **Fase 2: Backend** 🔄
- [ ] Actualizar structs en `internal/handlers/expenses/create.go`
- [ ] Agregar validaciones de recurrencia
- [ ] Implementar cálculo de `end_date`
- [ ] Actualizar respuestas de API
- [ ] Tests unitarios

### **Fase 3: Frontend** 🔄
- [ ] Actualizar types en `types/expense.ts`
- [ ] Mejorar `ExpenseForm.tsx` con campos condicionales
- [ ] Actualizar visualización en `Expenses.tsx`
- [ ] Agregar helpers para calcular próximas fechas

### **Fase 4: Documentación** ✅
- [x] Este documento de diseño
- [ ] Actualizar `API.md`
- [ ] Actualizar `README.md`
- [ ] Crear ejemplos de uso

### **Fase 5 (Futura): Auto-generación**
- [ ] CRON job para crear gastos futuros
- [ ] Endpoint `GET /expenses/upcoming`
- [ ] Notificaciones de gastos próximos

---

**Fin del diseño técnico**

---

**Creado:** 2026-01-16  
**Última actualización:** 2026-01-16  
**Estado:** 📝 En Diseño → 🚧 Listo para implementar
