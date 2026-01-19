# 📖 Guía Narrativa de Funcionalidades - Bolsillo Claro

**Versión:** 1.0  
**Fecha:** 2026-01-16  
**Autor:** Documentación consolidada del proyecto

Este documento explica de forma narrativa y entendible qué hace cada módulo del sistema, cómo funcionan las features, y qué flujos están disponibles.

---

## 📋 Índice

- [Autenticación](#-módulo-de-autenticación)
- [Cuentas](#-módulo-de-cuentas)
- [Gastos](#-módulo-de-gastos)
- [Ingresos](#-módulo-de-ingresos)
- [Metas de Ahorro](#-módulo-de-metas-de-ahorro)
- [Categorías](#️-módulo-de-categorías)
- [Dashboard](#-módulo-de-dashboard)
- [Sistema de Recurrencia Avanzado](#-sistema-de-recurrencia-avanzado)
- [Sistema de Multi-Currency](#-sistema-de-multi-currency-modo-3)
- [Cuentas Familiares](#-sistema-de-cuentas-familiares)
- [Preguntas Frecuentes](#-preguntas-frecuentes)

---

## 🔐 **MÓDULO DE AUTENTICACIÓN**

El sistema de autenticación permite registrar usuarios nuevos e iniciar sesión. Al registrarte, proporcionás email, contraseña (mínimo 8 caracteres) y nombre. El sistema hashea la contraseña con bcrypt y devuelve automáticamente tokens JWT (access + refresh), lo que significa que quedás logueado inmediatamente sin necesidad de hacer login después del registro.

Al hacer login, ingresás email y contraseña, y si son correctos recibís un access token (válido por 15 minutos) y un refresh token (válido por 7 días). El refresh token se almacena en una cookie httpOnly por seguridad. Cuando el access token expira, podés usar el refresh token para obtener uno nuevo sin tener que volver a ingresar credenciales.

Todos los endpoints excepto register y login requieren el header `Authorization: Bearer <token>`. Si el token es inválido o expiró, obtenés error 401.

**Flujo típico:**
1. Usuario se registra → Recibe tokens → Queda logueado
2. Después de 15 minutos → Access token expira
3. Frontend usa refresh token → Obtiene nuevo access token
4. Después de 7 días → Refresh token expira → Usuario debe volver a hacer login

---

## 💰 **MÓDULO DE CUENTAS**

Una cuenta es la unidad fundamental del sistema donde se organizan todos los datos financieros. Un usuario puede crear múltiples cuentas completamente aisladas entre sí (por ejemplo: "Finanzas Personales", "Gastos Familia", "Mi Negocio").

### Crear una cuenta

Al crear una cuenta debés especificar:
- **Nombre:** Cualquier nombre descriptivo
- **Tipo:** `personal` o `family` (OBLIGATORIO)
- **Moneda primaria:** `ARS` o `USD` (esta será la moneda base para visualizaciones consolidadas)
- **Initial balance:** Siempre es 0, este campo existe pero no se usa actualmente

Si elegís tipo `family`, debés agregar al menos un miembro familiar (Mamá, Papá, Hijo, etc.). Los miembros no son usuarios con login propio, son etiquetas para poder asignar gastos e ingresos a personas específicas y después analizar quién gastó cuánto. Los miembros solo necesitan nombre, el email es opcional.

Al crear una cuenta se genera automáticamente una meta de ahorro especial llamada "Ahorro General" sin deadline, que funciona como ahorro no destinado a ningún objetivo específico.

### Gestionar cuentas

Podés listar todas tus cuentas, ver el detalle de cada una (que incluye estadísticas del mes actual), actualizar nombre y moneda, o eliminar una cuenta completa (esto elimina en cascada todos los gastos, ingresos, metas y miembros asociados).

**Headers importantes:**
- Endpoints de cuentas: Solo `Authorization: Bearer <token>`
- Endpoints de gastos/ingresos/dashboard: `Authorization` + `X-Account-ID: <uuid>`

El header `X-Account-ID` indica en qué cuenta estás operando. Esto permite cambiar de cuenta fácilmente desde el frontend.

---

## 💸 **MÓDULO DE GASTOS**

El módulo de gastos permite registrar tanto compras puntuales como compromisos financieros recurrentes.

### Crear un gasto

Un gasto puede tener **categoría o no** (es completamente opcional). Puede ser de **gasto único (`one-time`)** o **recurrente (`recurring`)**.

#### Gastos Únicos (One-Time)

Son compras que ocurren una sola vez: supermercado, cena en restaurante, compra de ropa, taxi. Estos gastos se registran en el mes en que ocurrieron y no afectan proyecciones futuras.

**Campos obligatorios:**
- Descripción
- Monto (positivo)
- Moneda (ARS, USD, EUR)
- Fecha (YYYY-MM-DD)
- Tipo: `one-time`

**Campos opcionales:**
- Categoría
- Family member (obligatorio si la cuenta es tipo `family`)

#### Gastos Recurrentes (Recurring)

Son compromisos que se repiten automáticamente: Netflix, Spotify, gimnasio, alquiler, seguro. 

**Configuración básica:**
- Fecha de inicio (`date`)
- Fecha de fin opcional (`end_date`): Si es null, el gasto se repite indefinidamente

**Recurrencia avanzada (según docs):**

Si es **recurrente**, podés decidir la **frecuencia de recurrencia**: `daily` (diario), `weekly` (semanal), `monthly` (mensual), o `yearly` (anual). Además, ese gasto recurrente requiere que especifiques automáticamente:

- **Día de cobro/débito:** 
  - Si es **semanal**: día de la semana del 0 al 6 (0=Domingo, 6=Sábado)
  - Si es **mensual**: día del mes del 1 al 31
  - Si es **anual**: día y mes específico

También podés configurar el **intervalo de recurrencia** (por ejemplo, "cada 2 semanas" significa `recurrence_interval: 2`).

El gasto recurrente puede tener **límite de recurrencias**. Por ejemplo, si comprás algo en 6 cuotas, configurás `total_occurrences: 6` y el sistema trackea en qué cuota estás (`current_occurrence`), mostrando "Cuota 3/6". Si no ponés límite (`total_occurrences: null`), el gasto se repite indefinidamente hasta que lo elimines o modifiques.

### Multi-Currency (Modo 3)

En la **moneda**, si la cuenta está configurada en pesos argentinos (`currency: ARS`), el frontend debería poner por default los pesos argentinos (aunque el backend requiere que envíes el campo explícitamente).

Al hacer un gasto en dólares cuando tu cuenta está en pesos, podés usar el **sistema de multi-currency Modo 3**: ponés el monto que realmente se debitó en pesos y el sistema hace automáticamente la conversión. 

**Ejemplo:**
```json
{
  "description": "Claude Pro - Enero 2026",
  "amount": 20,
  "currency": "USD",
  "amount_in_primary_currency": 31500,
  "date": "2026-01-16"
}
```

El sistema automáticamente:
- Calcula el tipo de cambio efectivo: `31500 / 20 = 1575`
- Guarda ambos valores: `exchange_rate: 1575` y `amount_in_primary_currency: 31500`
- Asocia la fecha del gasto como fecha del tipo de cambio

Esto captura perfectamente el "dólar tarjeta" argentino con impuestos incluidos (oficial $900 + 30% imp. PAÍS + 45% percepción ganancias = $1575).

### Impacto en el Balance

Cada gasto **se descuenta de los ingresos** en el cálculo del balance disponible que muestra el dashboard:

```
available_balance = total_income - total_expenses - total_assigned_to_goals
```

**Cómo funciona:**
- El dashboard suma todos los gastos del mes (one-time + recurring activos)
- Los gastos recurring **NO se crean físicamente cada mes**, se calculan virtualmente
- Al consultar gastos de febrero, el sistema devuelve gastos recurring que estén activos en febrero

### Gestionar Gastos Recurrentes

#### ¿Cómo freno un gasto recurrente infinito?

**Opción 1: Ponerle fecha de fin**
```json
PUT /expenses/uuid-del-alquiler
{
  "end_date": "2026-03-31"
}
```
El gasto deja de aparecer después de marzo.

**Opción 2: Eliminarlo**
```json
DELETE /expenses/uuid-del-alquiler
```
El gasto desaparece completamente del sistema.

**⚠️ Importante:** Como los gastos recurring no se crean físicamente cada mes, al eliminar el gasto recurring, el historial completo desaparece. Si querés mantener el historial, primero deberías ponerle `end_date` al mes anterior y después eliminarlo.

### Listado y Filtros

Podés listar gastos con filtros:
- `month`: Mes específico (YYYY-MM)
- `type`: 'one-time', 'recurring', o 'all'
- `category_id`: Filtrar por categoría
- `family_member_id`: Filtrar por miembro (solo cuentas family)
- `currency`: Filtrar por moneda

También podés:
- Obtener detalle de un gasto específico
- Actualizar gastos (excepto el tipo, no se puede convertir one-time en recurring o viceversa)
- Eliminar gastos

**Endpoint especial:**
- `GET /expenses/commitments?month=YYYY-MM`: Calcula compromisos recurrentes activos ese mes

---

## 💰 **MÓDULO DE INGRESOS**

El módulo de ingresos funciona de manera muy similar al de gastos, con la misma distinción entre ingresos puntuales y recurrentes.

### Crear un ingreso

Un ingreso puede tener **categoría o no** (opcional). Puede ser **único (`one-time`)** o **recurrente (`recurring`)**.

#### Ingresos Únicos (One-Time)

Dinero que entra una sola vez: venta de un artículo, bono único, reembolso, regalo en efectivo. Estos ingresos se registran en el mes en que ocurrieron.

#### Ingresos Recurrentes (Recurring)

Flujos de dinero que se repiten mensualmente: sueldo fijo, proyecto freelance que paga mensualmente, alquiler de propiedad, pensión, etc.

**Configuración:**
- **Fecha de inicio:** Cuándo empieza el ingreso recurrente
- **Fecha de fin (`end_date`):** Opcional. Si es null, el ingreso se repite indefinidamente (como un sueldo permanente). Si tiene fecha, representa contratos temporales (ej: proyecto freelance por 6 meses).

**Ejemplo - Sueldo permanente:**
```json
{
  "description": "Sueldo mensual",
  "amount": 200000,
  "currency": "ARS",
  "income_type": "recurring",
  "date": "2026-01-01",
  "end_date": null
}
```

**Ejemplo - Proyecto temporal:**
```json
{
  "description": "Proyecto freelance React",
  "amount": 1500,
  "currency": "USD",
  "income_type": "recurring",
  "date": "2026-01-01",
  "end_date": "2026-06-30"
}
```

### Multi-Currency

El sistema de **multi-currency Modo 3** también funciona para ingresos: si recibís USD pero querés registrar el monto exacto en pesos que ingresó a tu cuenta (considerando comisiones, tipo de cambio MEP, etc.), el sistema calcula la tasa efectiva automáticamente.

Campo `amount_in_primary_currency` es **opcional**: si no lo proporcionás, el sistema calcula la conversión usando `exchange_rate` provisto o busca la tasa en la tabla `exchange_rates` para esa fecha. Si no encuentra tasa, te pide que la proporciones.

### Gestión

Si la cuenta es tipo `family`, debés especificar qué miembro generó ese ingreso.

Podés listar ingresos con filtros idénticos a gastos (mes, tipo, categoría, miembro, moneda), obtener detalle individual, actualizar, o eliminar.

---

## 🎯 **MÓDULO DE METAS DE AHORRO**

Las metas de ahorro permiten convertir objetivos abstractos en planes concretos con números específicos.

### Crear una meta

Al crear una meta especificás:
- **Nombre:** Descripción del objetivo (ej: "Vacaciones en Brasil", "Auto nuevo")
- **Monto objetivo (`target_amount`):** Cuánto dinero querés juntar
- **Moneda:** ARS o USD
- **Deadline (opcional):** Fecha límite para alcanzar la meta
- **Descripción adicional (opcional):** Notas sobre la meta

Cada cuenta tiene automáticamente una meta especial llamada **"Ahorro General"** (`is_general: true`) que no tiene deadline y funciona como ahorro no destinado a ningún objetivo específico.

### Progreso y Cálculos

Las metas tienen un **monto actual (`current_amount`)** que empieza en 0 y se actualiza automáticamente cuando agregás o retirás fondos. El sistema calcula el **progreso porcentual** (`progress_percentage`) basado en `current_amount / target_amount × 100`.

**Metas con deadline** calculan automáticamente `required_monthly_savings`:

```
Querés juntar: $300,000
Ya tenés: $50,000
Tiempo: 6 meses
Cálculo: (300000 - 50000) / 6 = $41,666.67 por mes
```

**Metas sin deadline** no tienen este cálculo porque no hay presión temporal.

### Agregar y Retirar Fondos

**Agregar fondos:**
```json
POST /savings-goals/:id/add-funds
{
  "amount": 30000,
  "description": "Ahorro enero"
}
```

**Retirar fondos:**
```json
POST /savings-goals/:id/withdraw-funds
{
  "amount": 5000,
  "description": "Adelanto para pasaje"
}
```

Cada operación:
- Crea una entrada en `savings_goal_transactions`
- Actualiza automáticamente `current_amount` de la meta
- No podés retirar más de lo que tenés (`current_amount`)

### Impacto en el Balance

Cuando agregás fondos a una meta, **esto se descuenta del balance disponible**:

```
available_balance = total_income - total_expenses - total_assigned_to_goals
```

**Cómo funciona:**
- Las metas NO crean gastos reales (expenses)
- El dashboard calcula `total_assigned_to_goals` sumando el `current_amount` de todas tus metas activas
- Este monto representa el "capital inmovilizado" total (dinero que tenés pero no está disponible)
- Se resta del balance disponible

**Ejemplo:**
```
Ingresos enero: $200,000
Gastos enero: $120,000
Metas activas: "Vacaciones" $30,000 + "Auto" $20,000 = $50,000
→ Balance disponible: 200,000 - 120,000 - 50,000 = $30,000
```

**⚠️ Punto no claro:** Si retirás fondos de una meta, ¿se suma al available_balance? La documentación no lo especifica claramente.

### Historial y Gestión

El detalle de una meta incluye el historial completo de transacciones (agregados y retiros) para tener trazabilidad total de los movimientos.

Podés:
- Listar metas (filtro por activas/inactivas)
- Obtener detalle con historial
- Actualizar datos de la meta
- Eliminar metas

---

## 🏷️ **MÓDULO DE CATEGORÍAS**

El sistema tiene dos tipos de categorías completamente independientes: categorías de gastos y categorías de ingresos.

### Categorías Predefinidas

Existen **categorías predefinidas** (seeds) que todos los usuarios ven automáticamente:

**15 categorías de gastos:**
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

**10 categorías de ingresos:**
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

Cada categoría tiene nombre, ícono (emoji), y color (hex).

### Categorías Custom

Los usuarios pueden crear **categorías custom** específicas para su cuenta. Estas categorías solo son visibles para esa cuenta y tienen `is_custom: true` mientras que las predefinidas tienen `is_custom: false`.

Al listar categorías, recibirás una mezcla de las predefinidas (globales) y tus categorías custom.

### Restricciones

- **No podés editar ni borrar categorías predefinidas**
- Solo podés gestionar tus categorías custom
- No podés borrar categorías custom que tengan gastos o ingresos asociados (mantiene integridad referencial)
- Los nombres de categorías deben ser únicos por scope: las predefinidas son únicas globalmente, las custom son únicas por cuenta

---

## 📊 **MÓDULO DE DASHBOARD**

El dashboard proporciona un resumen financiero completo del mes actual o de cualquier mes que especifiques.

### Endpoint Principal

```
GET /dashboard/summary?month=YYYY-MM
```

Si no especificás `month`, usa el mes actual.

### Información Devuelta

**Totales del período:**
- `total_income`: Suma de todos los ingresos del mes en la moneda primaria
- `total_expenses`: Suma de todos los gastos del mes en la moneda primaria
- `total_assigned_to_goals`: Total de fondos en metas de ahorro activas (capital inmovilizado)
- `available_balance`: Cálculo automático = `total_income - total_expenses - total_assigned_to_goals`
- `primary_currency`: La moneda primaria de la cuenta (ARS o USD)

**Análisis de gastos:**
- `expenses_by_category`: Desglose de gastos agrupados por categoría
  - Total por categoría
  - Porcentaje del total
  - Nombre, ícono y color de la categoría
- `top_expenses`: Los 5 gastos más grandes del mes con todos sus detalles

**Transacciones recientes:**
- `recent_transactions`: Últimas 10 transacciones (mezcla de gastos e ingresos ordenados por fecha)
  - Cada una con campo `type` que indica si es "expense" o "income"

### Conversión Automática

**Todos los montos se convierten automáticamente a la moneda primaria de la cuenta** usando los `exchange_rate` guardados en cada transacción, lo que significa que ves un resumen consolidado sin importar en qué monedas fueron los movimientos originales.

**Ejemplo:**
```
Cuenta en ARS:
- Ingreso: $200,000 ARS
- Gasto: $50 USD (guardado con exchange_rate: 1000) = $50,000 ARS
- Total expenses mostrado: $50,000 ARS
```

Si no hay datos para el mes solicitado, los totales son 0 y los arrays están vacíos.

---

## 🔄 **SISTEMA DE RECURRENCIA AVANZADO**

El sistema implementa un modelo avanzado de recurrencia para gastos con campos adicionales en la tabla `expenses`.

### Campos de Recurrencia

- `recurrence_frequency`: 'daily', 'weekly', 'monthly', 'yearly'
- `recurrence_interval`: Cada cuántos períodos (ej: 2 = cada 2 semanas)
- `recurrence_day_of_month`: Día del mes (1-31) para frecuencia mensual/anual
- `recurrence_day_of_week`: Día de semana (0-6, 0=Domingo) para frecuencia semanal
- `total_occurrences`: Cantidad total de repeticiones. NULL = infinito
- `current_occurrence`: Número de ocurrencia actual (para mostrar "Cuota 3/6")
- `parent_expense_id`: ID del gasto padre para gastos auto-generados

### Casos de Uso Soportados

#### 1. Alquiler mensual indefinido
```json
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
Se repite todos los días 5 de cada mes, indefinidamente.

#### 2. Compra en cuotas
```json
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
6 cuotas de $8,000, día 16 de cada mes. El sistema mostrará "Cuota 1/6", "Cuota 2/6", etc.

#### 3. Suscripción anual
```json
{
  "description": "Netflix Premium - Anual",
  "amount": 5000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-15",
  "recurrence_frequency": "yearly",
  "recurrence_day_of_month": 15,
  "total_occurrences": null
}
```
Se cobra una vez al año, cada 15 de enero.

#### 4. Gastos semanales
```json
{
  "description": "Gym - Todos los lunes",
  "amount": 8000,
  "currency": "ARS",
  "expense_type": "recurring",
  "date": "2026-01-06",
  "recurrence_frequency": "weekly",
  "recurrence_day_of_week": 1,
  "total_occurrences": null
}
```
Se repite todos los lunes (día 1 de la semana).

### Validaciones

- Si es recurring, `recurrence_frequency` es **REQUERIDO**
- Si frequency='monthly' o 'yearly' → `recurrence_day_of_month` es **REQUERIDO** (1-31)
- Si frequency='weekly' → `recurrence_day_of_week` es **REQUERIDO** (0=Domingo, 6=Sábado)
- `recurrence_interval` default = 1
- `current_occurrence` default = 1
- Si `total_occurrences` está definido → `end_date` se calcula automáticamente
- Si es one-time, todos los campos de recurrencia deben ser NULL

### Cálculo Automático

El sistema calcula automáticamente cuándo ocurre la próxima instancia del gasto basándose en la frecuencia e intervalo configurados.

**Ejemplo - Cada 2 semanas:**
```json
{
  "recurrence_frequency": "weekly",
  "recurrence_interval": 2,
  "recurrence_day_of_week": 1
}
```
Próximas fechas: 06-ene, 20-ene, 03-feb, 17-feb...

---

## 💱 **SISTEMA DE MULTI-CURRENCY (Modo 3)**

El sistema implementa tres modos para manejar conversiones de moneda, siendo el **Modo 3** la estrella del sistema.

### Modo 1: Moneda Local (Automático)

Si el gasto/ingreso es en la misma moneda que la cuenta:
- `exchange_rate` = 1.0
- `amount_in_primary_currency` = amount

**Ejemplo:**
```json
{
  "amount": 15000,
  "currency": "ARS"
}
```
En cuenta ARS → `exchange_rate: 1.0`, `amount_in_primary_currency: 15000`

---

### Modo 2: Con Exchange Rate Provisto

Usuario proporciona la tasa de cambio:

```json
{
  "amount": 10,
  "currency": "USD",
  "exchange_rate": 900
}
```

Backend calcula: `amount_in_primary_currency = 10 × 900 = 9000`

---

### Modo 3: Con Monto Real Pagado ⭐

**ESTA ES LA FUNCIONALIDAD ESTRELLA**

Usuario proporciona el monto exacto debitado en moneda primaria:

```json
{
  "amount": 20,
  "currency": "USD",
  "amount_in_primary_currency": 31500
}
```

Backend calcula la tasa efectiva: `exchange_rate = 31500 / 20 = 1575`

### ¿Por qué es tan importante el Modo 3?

**Problema real (Argentina 2026):**

Usuario compra Claude Pro por USD 20 con tarjeta.
- Dólar oficial: $900
- **Dólar tarjeta (con impuestos):** $1,575
  - Base: $900
  - + 30% imp. PAÍS: $270
  - + 45% percepción ganancias: $405
  - **Total: $1,575 por dólar**
- **Monto real debitado:** ARS $31,500

Si guardamos solo "USD 20 a tasa 900", perdemos $13,500 de diferencia con la realidad.

**Con Modo 3:**
```json
{
  "description": "Claude Pro - Enero 2026",
  "amount": 20,
  "currency": "USD",
  "amount_in_primary_currency": 31500,
  "date": "2026-01-16"
}
```

El sistema automáticamente:
- Calcula: `exchange_rate = 31500 / 20 = 1575`
- Guarda ambos valores
- Registra la fecha del tipo de cambio (fecha del gasto)

**✅ Captura perfecta del dólar tarjeta argentino!**

### Tabla exchange_rates

Almacena histórico de tipos de cambio por fecha. Si el usuario no proporciona `exchange_rate` ni `amount_in_primary_currency`, el sistema busca la tasa en esta tabla para la fecha del movimiento.

Si no encuentra tasa, devuelve error:
```json
{
  "error": "no exchange rate found for this date",
  "suggestion": "please provide either 'exchange_rate' or 'amount_in_primary_currency'",
  "details": {
    "from_currency": "USD",
    "to_currency": "ARS",
    "date": "2026-01-16"
  }
}
```

### Snapshot Histórico

Cada transacción guarda su propio `exchange_rate` y `amount_in_primary_currency`, lo que significa que si cambió el tipo de cambio después, las transacciones viejas mantienen su valor correcto histórico.

**Ejemplo:**
```
Enero: Gasto USD 10 a cambio 900 → guardado como ARS 9,000
Febrero: Dólar sube a 1,200
Marzo: Consultás gastos de enero → sigue mostrando ARS 9,000 (correcto!)
```

---

## 👨‍👩‍👧‍👦 **SISTEMA DE CUENTAS FAMILIARES**

Las cuentas tipo `family` tienen funcionalidad especial para trackear quién gasta/ingresa qué.

### Crear Cuenta Familiar

Al crear una cuenta familiar debés agregar al menos un miembro:

```json
{
  "name": "Gastos Familia",
  "type": "family",
  "currency": "ARS",
  "members": [
    {
      "name": "Mamá",
      "email": "mama@example.com"
    },
    {
      "name": "Papá",
      "email": "papa@example.com"
    },
    {
      "name": "Juan"
    }
  ]
}
```

### Miembros Familiares

Cada miembro tiene:
- **Nombre:** Identificador del miembro (ej: "Mamá", "Papá", "Juan")
- **Email (opcional):** No se usa para autenticación, es solo informativo
- **is_active:** Bandera para desactivar miembros sin borrarlos

**⚠️ Importante:** Los miembros NO son usuarios. Solo el owner de la cuenta (el usuario autenticado) puede ver y gestionar todos los datos. Los miembros son etiquetas para atribuir movimientos.

### Atribución de Movimientos

Al crear gastos o ingresos en cuentas family, el campo `family_member_id` es **OBLIGATORIO**:

```json
{
  "description": "Supermercado",
  "amount": 25000,
  "currency": "ARS",
  "family_member_id": "uuid-de-mama",
  "date": "2026-01-16"
}
```

Esto permite después analizar:
- Cuánto gastó cada miembro
- En qué categorías gasta más cada uno
- Qué proporción de gastos totales paga cada miembro
- Quién aporta qué ingresos

### Gestión de Miembros

Podés:
- Agregar miembros nuevos
- Actualizar sus datos (nombre, email)
- Desactivar miembros (`is_active: false`)

**⚠️ No se pueden eliminar completamente miembros que tienen movimientos financieros asociados** porque destruiría la integridad histórica. Solo los marcás como inactivos y dejan de aparecer en selectores pero sus datos históricos siguen visibles.

---

## ❓ **PREGUNTAS FRECUENTES**

### ¿Si agrego fondos a una meta de ahorro, esto se descuenta de mis ingresos?

**SÍ**, pero de forma virtual/calculada:

Cuando agregás $30,000 a una meta:
- NO se crea un gasto (expense)
- El dashboard calcula `total_assigned_to_goals` sumando el `current_amount` de todas tus metas activas
- Este monto representa el "capital inmovilizado" (dinero que tenés pero no está disponible para gastar)
- Se resta en el cálculo de `available_balance`

**Fórmula:**
```
available_balance = total_income - total_expenses - total_assigned_to_goals
```

**Ejemplo:**
```
Ingresos enero: $200,000
Gastos enero: $120,000
Metas activas: "Vacaciones" $30,000 + "Auto" $20,000 = $50,000
→ Balance disponible: $30,000
```

**Nota:** `total_assigned_to_goals` refleja el capital TOTAL inmovilizado en metas activas, no solo lo agregado este mes. Esto te muestra cuánto dinero real tenés "congelado" en objetivos de ahorro.

---

### ¿Si agrego un gasto recurrente infinito, cada mes me crea un gasto automático?

**NO** se crean gastos automáticos físicamente.

**Cómo funciona:**
- Los gastos recurring se guardan **UNA SOLA VEZ** en la base de datos
- Al consultar `GET /expenses?month=2026-02`, el backend calcula qué gastos recurring están activos ese mes
- Aparecen en la lista pero no hay múltiples registros físicos

**Ventaja:** No duplicás datos
**Desventaja:** Si eliminás el gasto recurring, perdés todo el historial

**Ejemplo:**
```
Creás: Netflix $5,000/mes desde enero (infinito)
Base de datos: 1 registro
Consultas: 
  - GET /expenses?month=2026-01 → Aparece Netflix
  - GET /expenses?month=2026-02 → Aparece Netflix
  - GET /expenses?month=2026-12 → Aparece Netflix
```

---

### ¿Cómo freno un gasto recurrente infinito si ya no lo tengo?

**Opción 1: Ponerle fecha de fin (Recomendado)**

```json
PUT /expenses/uuid-del-alquiler
{
  "end_date": "2026-03-31"
}
```

El gasto deja de aparecer después de marzo pero mantiene el historial de enero-marzo.

**Opción 2: Eliminarlo directamente**

```json
DELETE /expenses/uuid-del-alquiler
```

El gasto desaparece completamente del sistema.

⚠️ **CUIDADO:** Como los gastos recurring no se crean físicamente cada mes, al eliminar el registro único, perdés TODO el historial. Si el gasto ya lleva 6 meses registrado, perdés esos 6 meses de datos.

**Recomendación:**
1. Ponerle `end_date` al mes anterior
2. Esperar un tiempo por si necesitás consultarlo
3. Recién ahí eliminarlo

---

### ¿Qué pasa si retiro fondos de una meta? ¿Se suma al available_balance?

**SÍ**, automáticamente:

**Cómo funciona:**
- Al retirar fondos, se crea una transacción tipo "withdraw"
- La meta actualiza su `current_amount` (resta el monto retirado)
- El dashboard calcula `total_assigned_to_goals` sumando el `current_amount` de todas las metas activas
- Como el `current_amount` de la meta disminuyó, `total_assigned_to_goals` también disminuye
- Por lo tanto, el `available_balance` aumenta automáticamente

**Ejemplo:**
```
Antes del retiro:
- Ingresos: $200,000
- Gastos: $120,000  
- Metas activas: $50,000
→ Balance disponible: $30,000

Retirás $20,000 de "Vacaciones":
- Ingresos: $200,000
- Gastos: $120,000
- Metas activas: $30,000 (bajó de $50k a $30k)
→ Balance disponible: $50,000 (aumentó $20k)
```

---

### ¿Puedo cambiar el tipo de una cuenta de personal a family?

**NO**, según la documentación:

> No se puede cambiar el type de la cuenta

Esto es porque requeriría migración compleja de datos (agregar miembros, reasignar gastos/ingresos, etc.).

Si necesitás cambiar de personal a family, tendrías que crear una cuenta nueva.

---

### ¿Qué pasa con los gastos en cuotas si cambio la fecha de vencimiento de la tarjeta?

**Esta funcionalidad NO está implementada aún.**

El sistema de recurrencia avanzado permite especificar "día del mes" pero NO hay lógica para modificar en masa todas las cuotas futuras.

**Workaround actual:**
Tendrías que eliminar el gasto recurring y crear uno nuevo con la fecha correcta (pero perdés el tracking de cuotas ya pagadas).

---

### ¿Puedo tener gastos sin categoría?

**SÍ**, las categorías son completamente opcionales.

El sistema funciona perfectamente sin categorías. El análisis `expenses_by_category` del dashboard simplemente mostrará una categoría "Sin categoría" para gastos no categorizados.

---

### ¿Cuántas cuentas puedo tener?

**Ilimitadas.** 

No hay restricción en la cantidad de cuentas que un usuario puede crear. Cada cuenta está completamente aislada de las demás.

---

## 📌 **NOTAS FINALES**

Este documento describe las funcionalidades según la **documentación actual del proyecto**. Algunas features documentadas pueden estar en fase de diseño o implementación parcial.

Para información técnica detallada sobre endpoints, request/response formats, y códigos de error, consultá:
- `API.md` - Especificación completa de la API REST
- `docs/RECURRENCE-SYSTEM-DESIGN.md` - Diseño técnico del sistema de recurrencia

---

**Versión del documento:** 1.0  
**Última actualización:** 2026-01-16
