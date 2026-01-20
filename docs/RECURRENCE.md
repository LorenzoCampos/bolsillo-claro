# 🔄 Sistema de Recurrencia Avanzado

**Status:** 📝 En Diseño / Roadmap v1.1  
**Versión:** 1.0  
**Fecha:** 2026-01-16

---

## ⚠️ Nota Importante

Este documento describe el sistema de recurrencia **AVANZADO** que está planeado para v1.1.

**Estado actual (v1.0):**
- ✅ Gastos/Ingresos recurring básicos (con `date` y `end_date`)
- ❌ NO implementados: campos de día específico, límite de ocurrencias, contador de cuotas

**Ver estado actual:** [FEATURES.md](../FEATURES.md#módulo-de-gastos)

---

## 📋 Índice

- [Objetivo](#objetivo)
- [Casos de Uso](#casos-de-uso)
- [Diseño de Base de Datos](#diseño-de-base-de-datos)
- [Lógica de Negocio](#lógica-de-negocio)
- [API](#api)
- [Ejemplos](#ejemplos)
- [Implementación](#implementación)

---

## 🎯 Objetivo

Extender el sistema actual de recurrencia básico para soportar:

1. **Frecuencias granulares:** Daily, weekly, monthly, yearly
2. **Día específico:** "Todos los días 5 del mes", "Todos los lunes"
3. **Intervalos:** "Cada 2 semanas", "Cada 3 meses"
4. **Límite de ocurrencias:** Compras en cuotas (6/6, 12/12)
5. **Tracking de cuotas:** Mostrar "Cuota 3/6"

---

## 💡 Casos de Uso

### Caso 1: Alquiler Mensual (Sin Fin)

**Escenario:** Alquiler de $80,000 que se paga el día 5 de cada mes, indefinidamente.

**Configuración:**
```json
{
  "description": "Alquiler Depto Palermo",
  "amount": 80000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-02-05",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 5,
  "recurrence_interval": 1,
  "total_occurrences": null
}
```

**Comportamiento:**
- Se cobra día 5 de febrero, marzo, abril, mayo... indefinidamente
- No tiene `end_date`
- Aparece en todos los meses futuros al consultar gastos

---

### Caso 2: Zapatillas en 6 Cuotas

**Escenario:** Compra de zapatillas por $48,000 en 6 cuotas de $8,000 c/u, vencimiento día 16.

**Configuración:**
```json
{
  "description": "Zapatillas Nike - Cuota 1/6",
  "amount": 8000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-16",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 16,
  "recurrence_interval": 1,
  "total_occurrences": 6,
  "current_occurrence": 1
}
```

**Comportamiento:**
- Cuota 1: 16-ene-2026
- Cuota 2: 16-feb-2026
- Cuota 3: 16-mar-2026
- Cuota 4: 16-abr-2026
- Cuota 5: 16-may-2026
- Cuota 6: 16-jun-2026
- `end_date` se calcula automáticamente: `2026-06-16`
- UI muestra: "Cuota 3/6" en marzo

---

### Caso 3: Gimnasio Todos los Lunes

**Escenario:** Clase de gym que se paga $2,000 todos los lunes.

**Configuración:**
```json
{
  "description": "Clase Gym - Lunes",
  "amount": 2000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-06",
  "recurrence_frequency": "weekly",
  "recurrence_day_of_week": 1,
  "recurrence_interval": 1,
  "total_occurrences": null
}
```

**Comportamiento:**
- Se repite todos los lunes (día 1 = lunes, 0 = domingo)
- Sin fin

---

### Caso 4: Suscripción Anual

**Escenario:** Netflix que se paga una vez al año, día 15 de enero.

**Configuración:**
```json
{
  "description": "Netflix Premium - Anual",
  "amount": 60000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-15",
  "recurrence_frequency": "yearly",
  "recurrence_day_of_month": 15,
  "recurrence_interval": 1,
  "total_occurrences": null
}
```

**Comportamiento:**
- Se cobra una vez al año: 15-ene-2026, 15-ene-2027, 15-ene-2028...

---

### Caso 5: Pago Quincenal

**Escenario:** Servicio que se paga cada 2 semanas.

**Configuración:**
```json
{
  "description": "Servicio de Limpieza",
  "amount": 15000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-06",
  "recurrence_frequency": "weekly",
  "recurrence_day_of_week": 1,
  "recurrence_interval": 2,
  "total_occurrences": null
}
```

**Comportamiento:**
- 06-ene (lunes), 20-ene (lunes +2 semanas), 03-feb, 17-feb...

---

## 🗄️ Diseño de Base de Datos

### Nuevos Campos (Migración 012)

```sql
-- backend/migrations/012_add_recurrence_fields.sql

ALTER TABLE expenses
  ADD COLUMN recurrence_frequency TEXT 
    CHECK (recurrence_frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
  ADD COLUMN recurrence_interval INT DEFAULT 1 
    CHECK (recurrence_interval > 0),
  ADD COLUMN recurrence_day_of_month INT 
    CHECK (recurrence_day_of_month BETWEEN 1 AND 31),
  ADD COLUMN recurrence_day_of_week INT 
    CHECK (recurrence_day_of_week BETWEEN 0 AND 6),
  ADD COLUMN total_occurrences INT 
    CHECK (total_occurrences > 0 OR total_occurrences IS NULL),
  ADD COLUMN current_occurrence INT DEFAULT 1 
    CHECK (current_occurrence > 0),
  ADD COLUMN parent_expense_id UUID 
    REFERENCES expenses(id) ON DELETE CASCADE;

-- Índices
CREATE INDEX idx_expenses_recurrence_frequency ON expenses(recurrence_frequency);
CREATE INDEX idx_expenses_parent_expense_id ON expenses(parent_expense_id);
```

### Constraints

```sql
-- Si es recurrente, debe tener frecuencia
ALTER TABLE expenses
  ADD CONSTRAINT check_recurring_has_frequency 
  CHECK (
    (expense_type = 'one-time' AND recurrence_frequency IS NULL) OR
    (expense_type = 'recurring' AND recurrence_frequency IS NOT NULL)
  );

-- Si es mensual/anual, debe tener día del mes
ALTER TABLE expenses
  ADD CONSTRAINT check_monthly_has_day 
  CHECK (
    (recurrence_frequency NOT IN ('monthly', 'yearly')) OR
    (recurrence_frequency IN ('monthly', 'yearly') AND recurrence_day_of_month IS NOT NULL)
  );

-- Si es semanal, debe tener día de la semana
ALTER TABLE expenses
  ADD CONSTRAINT check_weekly_has_day 
  CHECK (
    (recurrence_frequency != 'weekly') OR
    (recurrence_frequency = 'weekly' AND recurrence_day_of_week IS NOT NULL)
  );

-- current_occurrence no puede exceder total_occurrences
ALTER TABLE expenses
  ADD CONSTRAINT check_current_within_total 
  CHECK (
    total_occurrences IS NULL OR 
    current_occurrence <= total_occurrences
  );
```

### Descripción de Campos

| Campo | Tipo | Descripción | Ejemplo |
|-------|------|-------------|---------|
| `recurrence_frequency` | TEXT | Frecuencia: `daily`, `weekly`, `monthly`, `yearly` | `'monthly'` |
| `recurrence_interval` | INT | Cada cuántos períodos (default: 1) | `2` (cada 2 semanas) |
| `recurrence_day_of_month` | INT | Día del mes (1-31), requerido si frequency=monthly/yearly | `5` |
| `recurrence_day_of_week` | INT | Día semana (0-6, 0=Domingo), requerido si frequency=weekly | `1` (Lunes) |
| `total_occurrences` | INT | Total de repeticiones. NULL = infinito | `6` o `NULL` |
| `current_occurrence` | INT | Ocurrencia actual (para mostrar "3/6") | `3` |
| `parent_expense_id` | UUID | ID del gasto padre (para gastos auto-generados) | UUID o `NULL` |

---

## 🔧 Lógica de Negocio

### Validaciones al Crear

**Si `expense_type = 'recurring'`:**

1. `recurrence_frequency` es **REQUERIDO**
2. Si `frequency = 'monthly'` o `'yearly'`:
   - `recurrence_day_of_month` es **REQUERIDO** (1-31)
3. Si `frequency = 'weekly'`:
   - `recurrence_day_of_week` es **REQUERIDO** (0-6)
4. `recurrence_interval` default = 1
5. `current_occurrence` default = 1
6. Si `total_occurrences` está definido:
   - `end_date` se calcula automáticamente
   - `current_occurrence` <= `total_occurrences`

**Si `expense_type = 'one-time'`:**
- Todos los campos de recurrencia deben ser `NULL`

### Cálculo de end_date Automático

```go
func CalculateEndDate(startDate time.Time, frequency string, interval int, totalOccurrences int) time.Time {
    if totalOccurrences == 0 {
        return time.Time{} // NULL
    }
    
    switch frequency {
    case "daily":
        return startDate.AddDate(0, 0, interval * (totalOccurrences - 1))
    
    case "weekly":
        return startDate.AddDate(0, 0, 7 * interval * (totalOccurrences - 1))
    
    case "monthly":
        return startDate.AddDate(0, interval * (totalOccurrences - 1), 0)
    
    case "yearly":
        return startDate.AddDate(interval * (totalOccurrences - 1), 0, 0)
    }
}
```

**Ejemplo:**
```
Fecha inicio: 2026-01-16
Frecuencia: monthly
Interval: 1
Total cuotas: 6

Cálculo: 2026-01-16 + 5 meses = 2026-06-16
```

### Cálculo de Próxima Ocurrencia

```go
func CalculateNextOccurrence(expense Expense) time.Time {
    switch expense.RecurrenceFrequency {
    case "daily":
        return expense.Date.AddDate(0, 0, expense.RecurrenceInterval)
    
    case "weekly":
        return expense.Date.AddDate(0, 0, 7 * expense.RecurrenceInterval)
    
    case "monthly":
        nextMonth := expense.Date.AddDate(0, expense.RecurrenceInterval, 0)
        return SetDayOfMonth(nextMonth, expense.RecurrenceDayOfMonth)
    
    case "yearly":
        nextYear := expense.Date.AddDate(expense.RecurrenceInterval, 0, 0)
        return SetDayOfMonth(nextYear, expense.RecurrenceDayOfMonth)
    }
}
```

### Edge Cases

#### Día 31 en meses de 30 días

**Problema:** Recurrencia mensual día 31, pero febrero tiene 28/29 días.

**Solución:**
```go
func SetDayOfMonth(date time.Time, day int) time.Time {
    lastDayOfMonth := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
    
    if day > lastDayOfMonth {
        day = lastDayOfMonth
    }
    
    return time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, time.UTC)
}
```

**Ejemplo:**
```
Recurrencia: día 31 de cada mes
Enero 31 → ✅ 31-ene
Febrero 31 → ⚠️ Ajusta a 28-feb (o 29 si bisiesto)
Marzo 31 → ✅ 31-mar
Abril 31 → ⚠️ Ajusta a 30-abr
```

---

## 📡 API

### POST /expenses (Actualizado)

**Request Body:**
```json
{
  "category_id": "uuid (opcional)",
  "amount": 8000,
  "currency": "ARS",
  "expense_type": "recurring",
  "description": "Zapatillas - Cuota 1/6",
  "date": "2026-01-16",
  
  // ⭐ NUEVOS CAMPOS DE RECURRENCIA
  "recurrence_frequency": "monthly",
  "recurrence_interval": 1,
  "recurrence_day_of_month": 16,
  "recurrence_day_of_week": null,
  "total_occurrences": 6,
  "current_occurrence": 1
}
```

**Response:**
```json
{
  "id": "uuid",
  "description": "Zapatillas - Cuota 1/6",
  "amount": 8000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-16",
  "end_date": "2026-06-16",
  
  "recurrence_frequency": "monthly",
  "recurrence_interval": 1,
  "recurrence_day_of_month": 16,
  "recurrence_day_of_week": null,
  "total_occurrences": 6,
  "current_occurrence": 1,
  "parent_expense_id": null,
  
  "created_at": "2026-01-16T10:00:00Z"
}
```

### GET /expenses?month=2026-02

**Lógica de Filtrado:**

Para gastos recurrentes, calcula si están activos en el mes solicitado:

```go
func IsActiveInMonth(expense Expense, month string) bool {
    // Parsear mes solicitado
    requestedMonth, _ := time.Parse("2006-01", month)
    
    // Verificar que el gasto haya empezado
    if expense.Date.After(requestedMonth) {
        return false
    }
    
    // Verificar end_date si existe
    if expense.EndDate != nil && expense.EndDate.Before(requestedMonth) {
        return false
    }
    
    // Verificar si la fecha calculada cae en el mes
    nextOccurrence := CalculateNextOccurrenceForMonth(expense, requestedMonth)
    return nextOccurrence.Month() == requestedMonth.Month()
}
```

---

## 📝 Ejemplos Completos

### Ejemplo 1: Alquiler Mensual

**POST /expenses:**
```json
{
  "description": "Alquiler Depto",
  "amount": 80000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-02-05",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 5,
  "total_occurrences": null
}
```

**Consultas:**
```
GET /expenses?month=2026-02 → Aparece (día 5)
GET /expenses?month=2026-03 → Aparece (día 5)
GET /expenses?month=2026-12 → Aparece (día 5)
```

---

### Ejemplo 2: Compra en 6 Cuotas

**POST /expenses:**
```json
{
  "description": "Notebook Dell",
  "amount": 8000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-10",
  "recurrence_frequency": "monthly",
  "recurrence_day_of_month": 10,
  "total_occurrences": 6,
  "current_occurrence": 1
}
```

**Backend calcula `end_date`:** `2026-06-10`

**Consultas:**
```
GET /expenses?month=2026-01 → "Cuota 1/6"
GET /expenses?month=2026-02 → "Cuota 2/6"
GET /expenses?month=2026-06 → "Cuota 6/6"
GET /expenses?month=2026-07 → NO aparece (ya terminó)
```

---

## 🚀 Implementación

### Roadmap

**Fase 1: Base de Datos**
- [ ] Crear migración `012_add_recurrence_fields.sql`
- [ ] Ejecutar en desarrollo y producción
- [ ] Testear constraints

**Fase 2: Backend**
- [ ] Actualizar structs en Go
- [ ] Agregar validaciones de recurrencia
- [ ] Implementar `CalculateEndDate()`
- [ ] Implementar `CalculateNextOccurrence()`
- [ ] Manejar edge cases (día 31, año bisiesto)
- [ ] Tests unitarios

**Fase 3: Frontend**
- [ ] Actualizar types TypeScript
- [ ] Mejorar `ExpenseForm` con campos condicionales
- [ ] Selector de frecuencia
- [ ] Selector de día del mes/semana
- [ ] Toggle "Sin fin" vs "Cantidad de cuotas"
- [ ] Mostrar "Cuota X/Y" en lista

**Fase 4: Testing**
- [ ] Tests end-to-end
- [ ] Casos edge (meses cortos, bisiestos)
- [ ] Performance con muchos gastos recurrentes

---

## ⚠️ Consideraciones

### Performance

Con 100 gastos recurrentes, consultar un mes requiere:
- Iterar 100 registros
- Calcular próxima ocurrencia para cada uno
- Filtrar por mes

**Optimización:**
- Índice en `recurrence_frequency`
- Cache de gastos recurrentes activos
- Considerar materialized view para gastos futuros

### Alternativa: Generación Física

En lugar de cálculo on-demand, CRON job que crea gastos físicos:

**Ventajas:**
- Queries simples (no cálculos)
- Performance predecible

**Desventajas:**
- Complejidad adicional (CRON)
- Posibles inconsistencias
- Más difícil de modificar gastos pasados

**Decisión (v1.1):** On-demand primero, CRON si hay problemas de performance.

---

**Última actualización:** 2026-01-16  
**Status:** En diseño para v1.1  
**Versión:** 1.0
