# Bolsillo Claro - Arquitectura y Especificaciones Técnicas Completas

Este documento contiene las especificaciones técnicas detalladas de cada módulo del sistema Bolsillo Claro. Está diseñado como una guía de implementación práctica que responde preguntas concretas sobre qué funcionalidad existe, cómo se comporta cada feature, qué validaciones se aplican, y cómo interactúan los diferentes módulos entre sí.

## Tabla de Contenidos

1. [Visión General del Sistema](#visión-general-del-sistema)
2. [Arquitectura de Datos](#arquitectura-de-datos)
3. [Módulo de Autenticación y Usuarios](#módulo-de-autenticación-y-usuarios)
4. [Sistema de Cuentas](#sistema-de-cuentas)
5. [Módulo de Gastos](#módulo-de-gastos)
6. [Módulo de Ingresos](#módulo-de-ingresos)
7. [Módulo de Ahorros](#módulo-de-ahorros)
8. [Módulo de Lista de Compras](#módulo-de-lista-de-compras)
9. [Sistema de Categorización](#sistema-de-categorización)
10. [Dashboard y Analytics](#dashboard-y-analytics)
11. [Sistema de Monedas](#sistema-de-monedas)
12. [Reglas de Negocio Transversales](#reglas-de-negocio-transversales)
13. [Flujos de Usuario Principales](#flujos-de-usuario-principales)
14. [Roadmap de Implementación Detallado](#roadmap-de-implementación-detallado)

---

## Visión General del Sistema

Bolsillo Claro es un gestor financiero personal y familiar que responde tres preguntas fundamentales: ¿cuánto gasto?, ¿cuánto ingresa?, y ¿cuánto estoy ahorrando? El sistema está diseñado específicamente para la realidad económica argentina donde conviven múltiples monedas, los ingresos pueden ser variables, y las familias necesitan trackear contribuciones individuales de cada miembro.

### Principios de Diseño Fundamentales

El sistema se construye sobre varios principios que guían todas las decisiones de implementación.

**Simplicidad sobre Completitud**: Cada feature debe resolver un problema real sin agregar complejidad innecesaria. Si una funcionalidad requiere explicación extensa para ser entendida, probablemente es muy compleja para incluir. La experiencia de usuario debe sentirse natural e intuitiva desde el primer uso.

**Datos Aislados por Cuenta**: La cuenta es la unidad fundamental de organización. Todos los datos financieros pertenecen a una cuenta específica. Un usuario puede administrar múltiples cuentas completamente separadas. Nunca debe ser posible ver, mezclar, o confundir datos entre cuentas diferentes. Este aislamiento debe ser absoluto tanto a nivel de interfaz como de base de datos.

**Categorización Opcional**: El sistema funciona completamente sin categorías. Un usuario puede registrar gastos e ingresos sin nunca asignarles categoría, y todas las funcionalidades principales siguen operando normalmente. Las categorías son un feature opt-in para usuarios que quieren ese nivel de análisis adicional, pero no son requisito obligatorio.

**Realidad de Múltiples Monedas**: Pesos y dólares son ciudadanos de primera clase en el sistema. No hay "moneda principal" con otras como secundarias. Cada transacción se registra en su moneda original y se mantiene así en la base de datos. Las conversiones solo ocurren para visualización consolidada, nunca alterando los datos originales.

**Claridad sobre Automatización**: El sistema calcula y muestra información útil automáticamente, pero siempre de forma transparente. Cuando el sistema hace un cálculo o proyección, el usuario debe poder entender de dónde vino ese número. No hay "cajas negras" o cálculos opacos que generen desconfianza.

**Mobile-First Pero Web-First**: La interfaz principal es web, pero debe funcionar perfectamente en móviles. La mayoría de las consultas rápidas ("¿cuánto gasté este mes?") suceden desde el celular, mientras que análisis más profundos y configuración ocurren desde desktop. El diseño debe adaptarse a ambos contextos fluidamente.

### Módulos Principales y Sus Interacciones

El sistema se organiza en cuatro módulos principales que interactúan entre sí de formas específicas.

El **Módulo de Gastos** permite registrar compras y compromisos financieros. Distingue entre gastos puntuales que ocurren una sola vez y gastos recurrentes que se repiten automáticamente cada mes. Los gastos recurrentes son particularmente importantes porque representan compromisos financieros continuos que muchas veces se olvidan al calcular disponibilidad de dinero mensual.

El **Módulo de Ingresos** maneja todo el dinero que entra a la cuenta. Reconoce que no todos tienen un sueldo fijo y permanente. Un freelancer puede tener contratos que pagan mensualmente por seis meses y luego terminan. Un comerciante tiene ventas variables. El sistema modelá estas realidades mediante diferentes tipos de ingresos con duraciones configurables.

El **Módulo de Ahorros** transforma objetivos abstractos en planes concretos con números específicos. Una meta de ahorro puede tener deadline, en cuyo caso el sistema calcula automáticamente cuánto necesitás ahorrar por mes para alcanzarla. O puede ser indefinida, funcionando como un objetivo de largo plazo sin presión temporal. Cada cuenta tiene automáticamente una meta especial de "Ahorro General" para dinero que no está destinado a ningún objetivo específico.

El **Módulo de Lista de Compras** es funcionalidad secundaria que permite registrar productos que querés comprar eventualmente sin fecha definida. No es para gastos inmediatos sino para deseos futuros. Lo útil es que podés convertir un producto de la lista en una meta de ahorro, creando un puente directo entre "quiero esto" y "voy a ahorrar para esto".

Estos módulos interactúan de formas específicas. Los gastos recurrentes proyectan compromisos mensuales que afectan cuánto dinero realmente tenés disponible. Los ingresos recurrentes proyectan entradas futuras que informan tus posibilidades de ahorro. Las metas de ahorro con deadline calculan cuánto necesitás guardar mensualmente basándose en la diferencia entre tus ingresos proyectados y gastos proyectados. La lista de compras puede generar metas de ahorro que se integran en el módulo de ahorros.

---

## Arquitectura de Datos

La base de datos PostgreSQL organiza toda la información del sistema en tablas relacionadas que reflejan los conceptos del negocio. Cada tabla tiene un propósito específico y relaciones claras con otras tablas. El aislamiento por cuenta se implementa a nivel de base de datos mediante la columna account_id presente en prácticamente todas las tablas.

### Tabla: users

Esta tabla almacena los usuarios del sistema. Cada usuario puede administrar múltiples cuentas.

Columnas:
- `id` (UUID, primary key): Identificador único del usuario generado automáticamente
- `email` (VARCHAR(255), unique, not null): Email del usuario usado para login, debe ser único en todo el sistema
- `password_hash` (VARCHAR(255), not null): Hash bcrypt de la contraseña del usuario, nunca se almacena la contraseña en texto plano
- `name` (VARCHAR(255), not null): Nombre completo del usuario para mostrar en la interfaz
- `created_at` (TIMESTAMP, not null): Fecha y hora de creación de la cuenta
- `updated_at` (TIMESTAMP, not null): Fecha y hora de última actualización de datos del usuario

Relaciones:
- Un usuario puede tener muchas cuentas (relación uno a muchos con accounts)

Validaciones a nivel de aplicación:
- Email debe ser formato válido de email
- Password debe tener mínimo 8 caracteres al crearse
- Name no puede estar vacío

### Tabla: accounts

Representa las cuentas financieras que son la unidad fundamental de organización del sistema. Cada cuenta pertenece a un usuario y contiene todos los datos financieros asociados completamente aislados de otras cuentas.

Columnas:
- `id` (UUID, primary key): Identificador único de la cuenta
- `user_id` (UUID, foreign key → users.id, not null): Usuario propietario de esta cuenta
- `name` (VARCHAR(255), not null): Nombre descriptivo de la cuenta como "Finanzas Personales" o "Gastos Familia"
- `type` (ENUM: 'personal', 'family', not null): Tipo de cuenta que determina si tiene miembros familiares o no
- `currency` (ENUM: 'ARS', 'USD', not null): Moneda base preferida para visualizaciones consolidadas
- `created_at` (TIMESTAMP, not null): Fecha de creación de la cuenta
- `updated_at` (TIMESTAMP, not null): Fecha de última actualización

Relaciones:
- Pertenece a un usuario (relación muchos a uno con users)
- Puede tener muchos miembros si es tipo family (relación uno a muchos con family_members)
- Tiene muchos gastos (relación uno a muchos con expenses)
- Tiene muchos ingresos (relación uno a muchos con incomes)
- Tiene muchas metas de ahorro (relación uno a muchos con savings_goals)
- Tiene muchos productos en lista de compras (relación uno a muchos con wishlist_products)
- Tiene muchas categorías (relación uno a muchos con categories)

Validaciones:
- Name no puede estar vacío
- Type debe ser exactamente 'personal' o 'family'
- Currency debe ser exactamente 'ARS' o 'USD'

Reglas especiales:
- Cuando se crea una cuenta nueva, automáticamente se debe crear una meta de ahorro especial llamada "Ahorro General" asociada a esa cuenta
- Si la cuenta es de tipo family, debe tener al menos un miembro familiar asociado

### Tabla: family_members

Almacena los miembros de cuentas familiares. Estos miembros no son usuarios independientes, son etiquetas para atribuir movimientos financieros dentro de una cuenta familiar.

Columnas:
- `id` (UUID, primary key): Identificador único del miembro
- `account_id` (UUID, foreign key → accounts.id, not null): Cuenta familiar a la que pertenece este miembro
- `name` (VARCHAR(255), not null): Nombre del miembro como "Mamá", "Papá", "Juan"
- `email` (VARCHAR(255), nullable): Email opcional para funcionalidades futuras, no se usa para autenticación
- `is_active` (BOOLEAN, not null, default true): Si el miembro está activo y debe aparecer en selectores
- `created_at` (TIMESTAMP, not null): Fecha de creación del miembro

Relaciones:
- Pertenece a una cuenta (relación muchos a uno con accounts)
- Puede tener muchos gastos atribuidos (relación uno a muchos con expenses)
- Puede tener muchos ingresos atribuidos (relación uno a muchos con incomes)
- Puede tener muchas entradas de ahorro atribuidas (relación uno a muchos con savings_entries)

Validaciones:
- Name no puede estar vacío
- Solo puede pertenecer a cuentas de tipo family
- No puede haber dos miembros con el mismo name en la misma account_id

Reglas especiales:
- Solo cuentas de tipo family pueden tener family_members
- Los miembros con is_active=false no aparecen en formularios pero sus movimientos históricos siguen visibles
- No se pueden eliminar miembros que tengan movimientos financieros asociados, solo marcarlos como inactivos

### Tabla: expenses

Almacena todos los gastos, tanto puntuales como recurrentes. Los gastos recurrentes existen una sola vez en esta tabla pero representan compromisos mensuales que se repiten.

Columnas:
- `id` (UUID, primary key): Identificador único del gasto
- `account_id` (UUID, foreign key → accounts.id, not null): Cuenta a la que pertenece este gasto
- `family_member_id` (UUID, foreign key → family_members.id, nullable): Si la cuenta es familiar, qué miembro realizó este gasto
- `category_id` (UUID, foreign key → categories.id, nullable): Categoría opcional del gasto
- `description` (VARCHAR(500), not null): Descripción del gasto como "Cena en restaurante" o "Netflix"
- `amount` (DECIMAL(15,2), not null): Monto del gasto, debe ser positivo
- `currency` (ENUM: 'ARS', 'USD', not null): Moneda en la que se realizó el gasto
- `expense_type` (ENUM: 'one-time', 'recurring', not null): Si es gasto puntual o recurrente
- `date` (DATE, not null): Fecha del gasto para one-time, o fecha de inicio para recurring
- `end_date` (DATE, nullable): Solo para recurring indefinidos puede ser null, indica cuándo termina el compromiso
- `created_at` (TIMESTAMP, not null): Cuándo se registró este gasto en el sistema
- `updated_at` (TIMESTAMP, not null): Última actualización del registro

Relaciones:
- Pertenece a una cuenta (relación muchos a uno con accounts)
- Opcionalmente pertenece a un miembro familiar (relación muchos a uno con family_members)
- Opcionalmente pertenece a una categoría (relación muchos a uno con categories)

Validaciones:
- Amount debe ser mayor que cero
- Currency debe ser 'ARS' o 'USD'
- expense_type debe ser 'one-time' o 'recurring'
- Si expense_type es 'one-time', end_date debe ser null
- Si expense_type es 'recurring', end_date puede ser null (indefinido) o una fecha futura mayor a date
- Si account.type es 'family', family_member_id es obligatorio
- Si account.type es 'personal', family_member_id debe ser null
- Description no puede estar vacío

Reglas especiales:
- Los gastos recurrentes representan compromisos mensuales que se repiten hasta end_date
- Un gasto recurrente sin end_date se repite indefinidamente hasta que el usuario lo modifique o elimine
- Los gastos recurrentes cuentan una sola vez en el total de gastos del mes en que ocurren, no se multiplican por cada repetición futura
- Para calcular compromisos mensuales futuros, se suman todos los gastos recurrentes activos en ese mes específico

Casos de uso concretos:

**Gasto Puntual**: Compraste comida en el supermercado por $15,000 pesos el 5 de Enero.
- expense_type: 'one-time'
- amount: 15000.00
- currency: 'ARS'
- date: 2025-01-05
- end_date: null
- Este gasto cuenta solo en Enero 2025

**Gasto Recurrente Indefinido**: Pagás Netflix $5,000 pesos mensuales desde el 10 de Enero sin fecha de finalización.
- expense_type: 'recurring'
- amount: 5000.00
- currency: 'ARS'
- date: 2025-01-10
- end_date: null
- Este gasto cuenta en Enero, Febrero, Marzo, y todos los meses siguientes hasta que lo canceles

**Gasto Recurrente con Duración**: Pagás un gimnasio $8,000 mensuales desde el 1 de Enero hasta el 30 de Junio.
- expense_type: 'recurring'
- amount: 8000.00
- currency: 'ARS'
- date: 2025-01-01
- end_date: 2025-06-30
- Este gasto cuenta en Enero, Febrero, Marzo, Abril, Mayo, Junio pero no en Julio en adelante

### Tabla: incomes

Almacena todos los ingresos, con soporte para ingresos puntuales, recurrentes temporales, y recurrentes indefinidos.

Columnas:
- `id` (UUID, primary key): Identificador único del ingreso
- `account_id` (UUID, foreign key → accounts.id, not null): Cuenta a la que pertenece
- `family_member_id` (UUID, foreign key → family_members.id, nullable): Si la cuenta es familiar, qué miembro generó este ingreso
- `category_id` (UUID, foreign key → categories.id, nullable): Categoría opcional del ingreso
- `description` (VARCHAR(500), not null): Descripción como "Sueldo", "Freelance proyecto X", "Venta de artículo"
- `amount` (DECIMAL(15,2), not null): Monto del ingreso, debe ser positivo
- `currency` (ENUM: 'ARS', 'USD', not null): Moneda del ingreso
- `income_type` (ENUM: 'one-time', 'recurring', not null): Si es ingreso único o recurrente
- `date` (DATE, not null): Fecha del ingreso para one-time, o fecha de inicio para recurring
- `end_date` (DATE, nullable): Para recurring puede ser null (indefinido) o fecha de finalización
- `created_at` (TIMESTAMP, not null): Cuándo se registró en el sistema
- `updated_at` (TIMESTAMP, not null): Última actualización

Relaciones:
- Pertenece a una cuenta (relación muchos a uno con accounts)
- Opcionalmente pertenece a un miembro familiar (relación muchos a uno con family_members)
- Opcionalmente pertenece a una categoría (relación muchos a uno con categories)

Validaciones:
- Amount debe ser mayor que cero
- Currency debe ser 'ARS' o 'USD'
- income_type debe ser 'one-time' o 'recurring'
- Si income_type es 'one-time', end_date debe ser null
- Si income_type es 'recurring', end_date puede ser null o fecha futura mayor a date
- Si account.type es 'family', family_member_id es obligatorio
- Si account.type es 'personal', family_member_id debe ser null
- Description no puede estar vacío

Reglas especiales:
- Los ingresos recurrentes representan flujos de dinero que se repiten mensualmente hasta end_date
- Un ingreso recurrente sin end_date representa un flujo permanente como un sueldo fijo
- Para proyecciones futuras, se consideran todos los ingresos recurrentes activos en el período proyectado

Casos de uso concretos:

**Ingreso Puntual**: Vendiste un artículo por $50,000 pesos el 15 de Enero.
- income_type: 'one-time'
- amount: 50000.00
- currency: 'ARS'
- date: 2025-01-15
- end_date: null
- Este ingreso cuenta solo en Enero 2025

**Ingreso Recurrente Indefinido**: Tenés un sueldo fijo de $200,000 mensuales desde el 1 de Enero sin fecha de finalización.
- income_type: 'recurring'
- amount: 200000.00
- currency: 'ARS'
- date: 2025-01-01
- end_date: null
- Este ingreso cuenta en Enero y todos los meses siguientes indefinidamente

**Ingreso Recurrente Temporal**: Tenés un proyecto freelance que paga $1,500 USD mensuales desde Enero hasta Junio.
- income_type: 'recurring'
- amount: 1500.00
- currency: 'USD'
- date: 2025-01-01
- end_date: 2025-06-30
- Este ingreso cuenta en Enero, Febrero, Marzo, Abril, Mayo, Junio pero no después

### Tabla: savings_goals

Almacena las metas de ahorro. Cada cuenta tiene una meta especial de "Ahorro General" más cualquier meta específica que el usuario cree.

Columnas:
- `id` (UUID, primary key): Identificador único de la meta
- `account_id` (UUID, foreign key → accounts.id, not null): Cuenta a la que pertenece
- `name` (VARCHAR(255), not null): Nombre de la meta como "Vacaciones", "Auto nuevo", "Ahorro General"
- `target_amount` (DECIMAL(15,2), not null): Monto objetivo a alcanzar, debe ser positivo
- `current_amount` (DECIMAL(15,2), not null, default 0): Monto actualmente ahorrado, se actualiza automáticamente
- `currency` (ENUM: 'ARS', 'USD', not null): Moneda de la meta
- `deadline` (DATE, nullable): Fecha objetivo para alcanzar la meta, null significa sin deadline
- `is_general` (BOOLEAN, not null, default false): True solo para la meta de Ahorro General de cada cuenta
- `created_at` (TIMESTAMP, not null): Cuándo se creó la meta
- `updated_at` (TIMESTAMP, not null): Última actualización

Relaciones:
- Pertenece a una cuenta (relación muchos a uno con accounts)
- Tiene muchas entradas de ahorro (relación uno a muchos con savings_entries)

Validaciones:
- target_amount debe ser mayor que cero
- current_amount debe ser mayor o igual a cero y menor o igual a target_amount
- Currency debe ser 'ARS' o 'USD'
- Si deadline existe, debe ser una fecha futura
- Name no puede estar vacío
- Solo puede existir una meta con is_general=true por account_id

Reglas especiales:
- Cada cuenta tiene exactamente una meta con is_general=true que se crea automáticamente al crear la cuenta
- La meta general no tiene deadline (deadline debe ser null)
- current_amount se actualiza automáticamente sumando todas las savings_entries asociadas a esta meta
- Las metas con deadline calculan required_monthly_savings = (target_amount - current_amount) / meses_restantes
- Las metas sin deadline no tienen required_monthly_savings calculado

Casos de uso concretos:

**Meta General**: La meta de ahorro general que toda cuenta tiene automáticamente.
- name: "Ahorro General"
- target_amount: 1000000.00 (puede ser un número muy alto porque es indefinida)
- currency: 'ARS'
- deadline: null
- is_general: true
- Esta meta captura ahorros que no están destinados a objetivos específicos

**Meta con Deadline**: Querés juntar $300,000 para vacaciones en 6 meses (de Enero a Junio).
- name: "Vacaciones en Brasil"
- target_amount: 300000.00
- current_amount: 50000.00 (ya ahorraste algo)
- currency: 'ARS'
- deadline: 2025-06-30
- is_general: false
- Required monthly: (300000 - 50000) / 6 = 41,666.67 pesos por mes

**Meta Sin Deadline**: Querés ahorrar $10,000 USD para un fondo de emergencia pero sin presión temporal.
- name: "Fondo de Emergencia"
- target_amount: 10000.00
- current_amount: 2000.00
- currency: 'USD'
- deadline: null
- is_general: false
- No hay required monthly porque no tiene fecha límite

### Tabla: savings_entries

Registra cada entrada individual de ahorro hacia una meta específica. Cada vez que el usuario ahorra dinero para una meta, se crea una entrada acá.

Columnas:
- `id` (UUID, primary key): Identificador único de la entrada
- `savings_goal_id` (UUID, foreign key → savings_goals.id, not null): Meta a la que pertenece este ahorro
- `family_member_id` (UUID, foreign key → family_members.id, nullable): Si la cuenta es familiar, quién ahorró
- `amount` (DECIMAL(15,2), not null): Monto ahorrado en esta entrada, debe ser positivo
- `currency` (ENUM: 'ARS', 'USD', not null): Moneda del ahorro
- `date` (DATE, not null): Fecha en que se realizó el ahorro
- `notes` (TEXT, nullable): Notas opcionales sobre este ahorro
- `created_at` (TIMESTAMP, not null): Cuándo se registró en el sistema

Relaciones:
- Pertenece a una meta de ahorro (relación muchos a uno con savings_goals)
- Opcionalmente pertenece a un miembro familiar (relación muchos a uno con family_members)

Validaciones:
- Amount debe ser mayor que cero
- Currency debe coincidir con savings_goal.currency
- Date no puede ser futura
- Si account.type es 'family', family_member_id es obligatorio
- Si account.type es 'personal', family_member_id debe ser null

Reglas especiales:
- Al crear una entrada, automáticamente se actualiza savings_goals.current_amount sumando este amount
- Al eliminar una entrada, automáticamente se actualiza savings_goals.current_amount restando este amount
- No se permite modificar el amount de una entrada existente, solo eliminar y crear nueva
- La suma total de todas las entries de una meta debe ser igual a savings_goal.current_amount

### Tabla: wishlist_products

Almacena productos que el usuario quiere comprar eventualmente. No son gastos inmediatos sino deseos futuros.

Columnas:
- `id` (UUID, primary key): Identificador único del producto
- `account_id` (UUID, foreign key → accounts.id, not null): Cuenta a la que pertenece
- `wishlist_category_id` (UUID, foreign key → wishlist_categories.id, nullable): Categoría opcional del producto
- `name` (VARCHAR(255), not null): Nombre del producto como "Mouse Logitech G502"
- `price` (DECIMAL(15,2), nullable): Precio estimado del producto
- `currency` (ENUM: 'ARS', 'USD', nullable): Moneda del precio
- `url` (TEXT, nullable): URL donde viste el producto o donde se puede comprar
- `notes` (TEXT, nullable): Notas adicionales sobre el producto
- `status` (ENUM: 'pending', 'purchased', not null, default 'pending'): Si está pendiente o ya comprado
- `purchased_date` (DATE, nullable): Fecha de compra si status es 'purchased'
- `created_at` (TIMESTAMP, not null): Cuándo se agregó a la lista
- `updated_at` (TIMESTAMP, not null): Última actualización

Relaciones:
- Pertenece a una cuenta (relación muchos a uno con accounts)
- Opcionalmente pertenece a una categoría de wishlist (relación muchos a uno con wishlist_categories)

Validaciones:
- Name no puede estar vacío
- Si price existe, debe ser mayor que cero
- Si price existe, currency debe existir y ser 'ARS' o 'USD'
- Si currency existe, price debe existir
- status debe ser 'pending' o 'purchased'
- Si status es 'purchased', purchased_date debe existir
- Si status es 'pending', purchased_date debe ser null

Reglas especiales:
- Un producto puede no tener precio si aún no sabés cuánto cuesta
- Desde un producto con precio definido, el usuario puede crear una meta de ahorro directamente
- Al marcar un producto como purchased, se debe obligatoriamente ingresar la purchased_date
- Los productos purchased siguen en la lista pero aparecen visualmente diferentes y pueden ser filtrados

### Tabla: wishlist_categories

Categorías específicas para organizar la lista de compras. Son independientes de las categorías de gastos/ingresos.

Columnas:
- `id` (UUID, primary key): Identificador único de la categoría
- `account_id` (UUID, foreign key → accounts.id, not null): Cuenta a la que pertenece
- `name` (VARCHAR(255), not null): Nombre de la categoría como "Tecnología", "Hogar", "Ropa"
- `color` (VARCHAR(7), nullable): Color hex para visualización como "#FF5733"
- `created_at` (TIMESTAMP, not null): Cuándo se creó la categoría

Relaciones:
- Pertenece a una cuenta (relación muchos a uno con accounts)
- Puede tener muchos productos (relación uno a muchos con wishlist_products)

Validaciones:
- Name no puede estar vacío
- No puede haber dos categorías con el mismo name en la misma account_id
- Si color existe, debe ser formato hex válido (#RRGGBB)

Reglas especiales:
- El usuario puede crear metas de ahorro desde una categoría completa sumando todos los productos de esa categoría
- No se pueden eliminar categorías que tienen productos asociados
- Las categorías de wishlist son completamente independientes de las categorías de gastos/ingresos

### Tabla: categories

Categorías para gastos e ingresos. Tienen estructura jerárquica de dos niveles (categoría padre y subcategorías).

Columnas:
- `id` (UUID, primary key): Identificador único de la categoría
- `account_id` (UUID, foreign key → accounts.id, not null): Cuenta a la que pertenece
- `parent_category_id` (UUID, foreign key → categories.id, nullable): Si es subcategoría, ID de su categoría padre
- `name` (VARCHAR(255), not null): Nombre de la categoría
- `type` (ENUM: 'expense', 'income', not null): Si es categoría de gastos o ingresos
- `color` (VARCHAR(7), nullable): Color hex para visualización
- `icon` (VARCHAR(50), nullable): Nombre del ícono para mostrar en UI
- `created_at` (TIMESTAMP, not null): Cuándo se creó la categoría

Relaciones:
- Pertenece a una cuenta (relación muchos a uno con accounts)
- Opcionalmente tiene una categoría padre (relación muchos a uno consigo misma)
- Puede tener muchas subcategorías (relación uno a muchos consigo misma)
- Puede tener muchos gastos (si type es 'expense', relación uno a muchos con expenses)
- Puede tener muchos ingresos (si type es 'income', relación uno a muchos con incomes)

Validaciones:
- Name no puede estar vacío
- Type debe ser 'expense' o 'income'
- Si parent_category_id existe, la categoría padre debe ser del mismo type
- Una categoría no puede ser su propio padre (directa o indirectamente)
- La jerarquía solo permite dos niveles (padres y una generación de hijos)
- Si color existe, debe ser formato hex válido

Reglas especiales:
- Las categorías padre no pueden tener parent_category_id
- Las subcategorías deben tener parent_category_id
- Solo se permiten dos niveles de jerarquía: nivel 1 (padres) y nivel 2 (hijos de esos padres)
- No se pueden eliminar categorías que tienen gastos o ingresos asociados
- Las categorías expense solo pueden asignarse a gastos, las income solo a ingresos

Ejemplo de jerarquía:
```
Alimentación (padre, type=expense)
  ├── Supermercado (hijo)
  ├── Restaurantes (hijo)
  └── Delivery (hijo)

Vivienda (padre, type=expense)
  ├── Alquiler (hijo)
  ├── Servicios (hijo)
  └── Mantenimiento (hijo)

Trabajo (padre, type=income)
  ├── Sueldo (hijo)
  ├── Freelance (hijo)
  └── Bonos (hijo)
```

### Tabla: exchange_rates

Almacena histórico de tipos de cambio para conversiones entre monedas.

Columnas:
- `id` (UUID, primary key): Identificador único del registro
- `from_currency` (ENUM: 'ARS', 'USD', not null): Moneda origen
- `to_currency` (ENUM: 'ARS', 'USD', not null): Moneda destino
- `rate` (DECIMAL(15,6), not null): Tasa de conversión
- `date` (DATE, not null): Fecha a la que corresponde esta tasa
- `source` (VARCHAR(100), nullable): Fuente de donde se obtuvo la tasa como "ExchangeRate-API"
- `created_at` (TIMESTAMP, not null): Cuándo se registró en el sistema

Validaciones:
- from_currency y to_currency deben ser diferentes
- rate debe ser mayor que cero
- No puede haber más de un registro para el mismo from_currency, to_currency, y date

Reglas especiales:
- El sistema debe almacenar al menos una tasa por día para poder hacer conversiones
- Para conversiones históricas, se usa la tasa más cercana a la fecha de la transacción
- Si no existe tasa para una fecha específica, se usa la más reciente anterior
- El sistema puede cachear tasas en memoria para operaciones frecuentes

---

## Módulo de Autenticación y Usuarios

Este módulo maneja registro, login, y gestión de sesiones de usuarios. La autenticación usa JWT con access y refresh tokens.

### Registro de Usuario

Endpoint: `POST /api/auth/register`

Request body:
```json
{
  "email": "usuario@example.com",
  "password": "contraseña-segura",
  "name": "Juan Pérez"
}
```

Validaciones:
- Email debe ser formato válido y no estar ya registrado
- Password debe tener mínimo 8 caracteres
- Name no puede estar vacío

Proceso:
1. Validar que el email no exista en la base de datos
2. Hashear el password usando bcrypt con cost factor 12
3. Crear el usuario en la tabla users
4. No crear ninguna cuenta automáticamente - el usuario debe crear su primera cuenta manualmente después de registrarse
5. Retornar token JWT de acceso y refresh token

Response exitoso (201 Created):
```json
{
  "user": {
    "id": "uuid-del-usuario",
    "email": "usuario@example.com",
    "name": "Juan Pérez"
  },
  "accessToken": "jwt-access-token",
  "refreshToken": "jwt-refresh-token"
}
```

### Login

Endpoint: `POST /api/auth/login`

Request body:
```json
{
  "email": "usuario@example.com",
  "password": "contraseña"
}
```

Proceso:
1. Buscar usuario por email
2. Comparar password hasheado usando bcrypt
3. Si coincide, generar access token (duración 15 minutos) y refresh token (duración 7 días)
4. Retornar tokens

Response exitoso (200 OK):
```json
{
  "user": {
    "id": "uuid-del-usuario",
    "email": "usuario@example.com",
    "name": "Juan Pérez"
  },
  "accessToken": "jwt-access-token",
  "refreshToken": "jwt-refresh-token"
}
```

El refresh token debe almacenarse en una cookie httpOnly para mayor seguridad.

### Refresh Token

Endpoint: `POST /api/auth/refresh`

Request: Cookie con refresh token

Proceso:
1. Extraer refresh token de la cookie
2. Verificar que es válido y no ha expirado
3. Generar nuevo access token
4. Retornar nuevo access token

Response exitoso (200 OK):
```json
{
  "accessToken": "nuevo-jwt-access-token"
}
```

### Logout

Endpoint: `POST /api/auth/logout`

Headers: Authorization con Bearer token

Proceso:
1. Invalidar refresh token del usuario (agregar a blacklist o remover de tabla de tokens activos)
2. Limpiar cookie de refresh token
3. Retornar confirmación

Response exitoso (200 OK):
```json
{
  "message": "Logout exitoso"
}
```

### Middleware de Autenticación

Todos los endpoints excepto register y login requieren autenticación. El middleware debe:

1. Extraer token del header Authorization: Bearer <token>
2. Verificar que el token es válido y no ha expirado
3. Extraer user_id del token
4. Agregar user_id al contexto de la request para uso en handlers
5. Si el token es inválido o expiró, retornar 401 Unauthorized

---

## Sistema de Cuentas

Las cuentas son la unidad fundamental de organización. Un usuario puede tener múltiples cuentas completamente aisladas.

### Crear Cuenta

Endpoint: `POST /api/accounts`

Headers: Authorization requerido

Request body:
```json
{
  "name": "Finanzas Personales",
  "type": "personal",
  "currency": "ARS"
}
```

Para cuenta familiar:
```json
{
  "name": "Gastos Familia",
  "type": "family",
  "currency": "ARS",
  "members": [
    { "name": "Mamá", "email": "mama@example.com" },
    { "name": "Papá", "email": "papa@example.com" },
    { "name": "Juan" }
  ]
}
```

Validaciones:
- name no puede estar vacío
- type debe ser 'personal' o 'family'
- currency debe ser 'ARS' o 'USD'
- Si type es 'family', members debe tener al menos un miembro
- Si type es 'personal', members debe estar vacío o no existir

Proceso:
1. Validar datos de entrada
2. Crear registro en tabla accounts
3. Si type es 'family', crear registros en family_members para cada miembro
4. Crear meta de ahorro general automáticamente asociada a esta cuenta
5. Retornar cuenta creada con sus miembros si corresponde

Response exitoso (201 Created):
```json
{
  "id": "uuid-de-cuenta",
  "name": "Gastos Familia",
  "type": "family",
  "currency": "ARS",
  "members": [
    {
      "id": "uuid-miembro-1",
      "name": "Mamá",
      "email": "mama@example.com",
      "isActive": true
    },
    {
      "id": "uuid-miembro-2",
      "name": "Papá",
      "email": "papa@example.com",
      "isActive": true
    }
  ],
  "createdAt": "2025-01-12T10:00:00Z"
}
```

### Listar Cuentas del Usuario

Endpoint: `GET /api/accounts`

Headers: Authorization requerido

Response exitoso (200 OK):
```json
{
  "accounts": [
    {
      "id": "uuid-1",
      "name": "Finanzas Personales",
      "type": "personal",
      "currency": "ARS",
      "createdAt": "2025-01-01T10:00:00Z"
    },
    {
      "id": "uuid-2",
      "name": "Gastos Familia",
      "type": "family",
      "currency": "USD",
      "memberCount": 3,
      "createdAt": "2025-01-05T15:00:00Z"
    }
  ]
}
```

### Obtener Detalle de Cuenta

Endpoint: `GET /api/accounts/:accountId`

Headers: Authorization requerido

Validaciones:
- La cuenta debe pertenecer al usuario autenticado

Response exitoso (200 OK):
```json
{
  "id": "uuid-de-cuenta",
  "name": "Gastos Familia",
  "type": "family",
  "currency": "ARS",
  "members": [
    {
      "id": "uuid-miembro-1",
      "name": "Mamá",
      "email": "mama@example.com",
      "isActive": true
    }
  ],
  "stats": {
    "totalExpensesCurrentMonth": 150000.00,
    "totalIncomesCurrentMonth": 300000.00,
    "savingsProgress": 45.5
  },
  "createdAt": "2025-01-12T10:00:00Z"
}
```

### Actualizar Cuenta

Endpoint: `PUT /api/accounts/:accountId`

Headers: Authorization requerido

Request body:
```json
{
  "name": "Nuevo Nombre",
  "currency": "USD"
}
```

Validaciones:
- La cuenta debe pertenecer al usuario autenticado
- No se puede cambiar el type de la cuenta
- name no puede estar vacío
- currency debe ser 'ARS' o 'USD'

Proceso:
1. Validar datos
2. Actualizar campos permitidos
3. No permitir cambiar type porque requeriría migración compleja de datos
4. Retornar cuenta actualizada

### Eliminar Cuenta

Endpoint: `DELETE /api/accounts/:accountId`

Headers: Authorization requerido

Validaciones:
- La cuenta debe pertenecer al usuario autenticado
- El usuario debe confirmar explícitamente que quiere eliminar todos los datos

Proceso:
1. Eliminar en cascada todos los datos asociados: expenses, incomes, savings_goals, savings_entries, wishlist_products, categories, family_members
2. Eliminar el registro de account
3. Retornar confirmación

Response exitoso (200 OK):
```json
{
  "message": "Cuenta eliminada exitosamente",
  "deletedAccountId": "uuid-de-cuenta"
}
```

### Gestionar Miembros Familiares

#### Agregar Miembro

Endpoint: `POST /api/accounts/:accountId/members`

Headers: Authorization requerido

Request body:
```json
{
  "name": "Hijo Mayor",
  "email": "hijo@example.com"
}
```

Validaciones:
- La cuenta debe ser tipo family
- La cuenta debe pertenecer al usuario autenticado
- name no puede estar vacío
- No puede existir otro miembro con el mismo name en esta cuenta

Response exitoso (201 Created):
```json
{
  "id": "uuid-nuevo-miembro",
  "name": "Hijo Mayor",
  "email": "hijo@example.com",
  "isActive": true,
  "createdAt": "2025-01-12T10:00:00Z"
}
```

#### Actualizar Miembro

Endpoint: `PUT /api/accounts/:accountId/members/:memberId`

Headers: Authorization requerido

Request body:
```json
{
  "name": "Juan Pablo",
  "email": "juanpablo@example.com"
}
```

Validaciones:
- El miembro debe pertenecer a la cuenta especificada
- La cuenta debe pertenecer al usuario autenticado
- name no puede estar vacío

#### Desactivar Miembro

Endpoint: `PUT /api/accounts/:accountId/members/:memberId/deactivate`

Headers: Authorization requerido

Proceso:
1. Marcar is_active = false
2. El miembro ya no aparece en selectores de formularios
3. Los movimientos históricos del miembro siguen siendo visibles

No se permite eliminar completamente miembros que tienen movimientos financieros asociados porque destruiría la integridad histórica de los datos.

---

## Módulo de Gastos

El módulo de gastos maneja tanto compras puntuales como compromisos recurrentes mensuales. La distinción entre estos dos tipos es fundamental para el funcionamiento del sistema.

### Conceptos Fundamentales

**Gasto Puntual (one-time)**: Una compra que ocurre una sola vez en una fecha específica. Ejemplos: compra de supermercado, cena en restaurante, compra de ropa, taxi. Estos gastos se registran en el mes en que ocurrieron y no afectan proyecciones futuras.

**Gasto Recurrente (recurring)**: Un compromiso financiero que se repite automáticamente cada mes. Ejemplos: Netflix, Spotify, gimnasio, seguro, alquiler. Estos gastos tienen fecha de inicio y opcionalmente fecha de finalización. Si no tienen fecha de finalización, se consideran indefinidos y se repiten hasta que el usuario los modifique o elimine.

La importancia de los gastos recurrentes es que representan compromisos mensuales que muchas veces se olvidan. Cuando alguien pregunta "¿cuánto gasto por mes?", los gastos recurrentes son críticos porque se repiten independientemente de si los recordás o no.

### Crear Gasto Puntual

Endpoint: `POST /api/expenses`

Headers: Authorization requerido, X-Account-ID requerido

Request body para cuenta personal:
```json
{
  "description": "Compra supermercado",
  "amount": 25000.50,
  "currency": "ARS",
  "expenseType": "one-time",
  "date": "2025-01-12",
  "categoryId": "uuid-categoria-opcional"
}
```

Request body para cuenta familiar:
```json
{
  "description": "Cena restaurante",
  "amount": 35000.00,
  "currency": "ARS",
  "expenseType": "one-time",
  "date": "2025-01-12",
  "familyMemberId": "uuid-miembro-obligatorio",
  "categoryId": "uuid-categoria-opcional"
}
```

Validaciones:
- description no puede estar vacío
- amount debe ser mayor que cero
- currency debe ser 'ARS' o 'USD'
- expenseType debe ser 'one-time'
- date debe ser fecha válida, puede ser pasada, presente, o futura
- Si account.type es 'family', familyMemberId es obligatorio y debe ser un miembro activo de esa cuenta
- Si account.type es 'personal', familyMemberId debe ser null o no estar presente
- Si categoryId existe, debe ser una categoría type='expense' de la misma cuenta

Proceso:
1. Validar todos los datos
2. Crear registro en expenses con expense_type='one-time' y end_date=null
3. Retornar gasto creado

Response exitoso (201 Created):
```json
{
  "id": "uuid-del-gasto",
  "description": "Compra supermercado",
  "amount": 25000.50,
  "currency": "ARS",
  "expenseType": "one-time",
  "date": "2025-01-12",
  "category": {
    "id": "uuid-categoria",
    "name": "Alimentación"
  },
  "createdAt": "2025-01-12T15:30:00Z"
}
```

### Crear Gasto Recurrente

Endpoint: `POST /api/expenses`

Headers: Authorization requerido, X-Account-ID requerido

Request body para recurrente indefinido:
```json
{
  "description": "Netflix Premium",
  "amount": 5000.00,
  "currency": "ARS",
  "expenseType": "recurring",
  "date": "2025-01-15",
  "endDate": null,
  "familyMemberId": "uuid-miembro-si-familia",
  "categoryId": "uuid-categoria-opcional"
}
```

Request body para recurrente con duración:
```json
{
  "description": "Gimnasio (plan 6 meses)",
  "amount": 8000.00,
  "currency": "ARS",
  "expenseType": "recurring",
  "date": "2025-01-01",
  "endDate": "2025-06-30",
  "familyMemberId": "uuid-miembro-si-familia"
}
```

Validaciones:
- Todas las validaciones de gasto puntual aplican
- expenseType debe ser 'recurring'
- date es la fecha de inicio del compromiso recurrente
- endDate puede ser null (indefinido) o una fecha futura mayor a date
- Si endDate existe, debe ser al menos 1 mes después de date

Proceso:
1. Validar datos
2. Crear registro en expenses con expense_type='recurring'
3. El sistema automáticamente considera este gasto en todos los meses entre date y endDate (o infinito si endDate es null)
4. Retornar gasto creado

Response exitoso (201 Created):
```json
{
  "id": "uuid-del-gasto",
  "description": "Netflix Premium",
  "amount": 5000.00,
  "currency": "ARS",
  "expenseType": "recurring",
  "date": "2025-01-15",
  "endDate": null,
  "monthlyImpact": 5000.00,
  "activeMonths": "indefinido",
  "createdAt": "2025-01-12T15:30:00Z"
}
```

### Listar Gastos

Endpoint: `GET /api/expenses?month=2025-01&type=all`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `month` (opcional): Formato YYYY-MM para filtrar gastos de un mes específico. Si no se provee, retorna gastos del mes actual
- `type` (opcional): 'one-time', 'recurring', o 'all' (default: 'all')
- `categoryId` (opcional): UUID de categoría para filtrar
- `familyMemberId` (opcional): UUID de miembro para filtrar (solo en cuentas family)
- `currency` (opcional): 'ARS', 'USD', o 'all' (default: 'all')

Lógica de filtrado por mes:
- Gastos one-time: incluir solo si date está en el mes solicitado
- Gastos recurring: incluir si el mes solicitado está entre date y endDate (o si endDate es null y el mes es >= date)

Response exitoso (200 OK):
```json
{
  "month": "2025-01",
  "expenses": [
    {
      "id": "uuid-1",
      "description": "Compra supermercado",
      "amount": 25000.50,
      "currency": "ARS",
      "expenseType": "one-time",
      "date": "2025-01-12",
      "category": {
        "id": "uuid-cat",
        "name": "Alimentación"
      },
      "familyMember": {
        "id": "uuid-miembro",
        "name": "Mamá"
      }
    },
    {
      "id": "uuid-2",
      "description": "Netflix Premium",
      "amount": 5000.00,
      "currency": "ARS",
      "expenseType": "recurring",
      "date": "2025-01-15",
      "endDate": null,
      "category": {
        "id": "uuid-cat",
        "name": "Entretenimiento"
      }
    }
  ],
  "summary": {
    "totalOneTime": 25000.50,
    "totalRecurring": 5000.00,
    "total": 30000.50,
    "count": 2,
    "byType": {
      "one-time": { "count": 1, "total": 25000.50 },
      "recurring": { "count": 1, "total": 5000.00 }
    }
  }
}
```

### Obtener Detalle de Gasto

Endpoint: `GET /api/expenses/:expenseId`

Headers: Authorization requerido, X-Account-ID requerido

Validaciones:
- El gasto debe pertenecer a la cuenta especificada en X-Account-ID

Response exitoso (200 OK):
```json
{
  "id": "uuid-del-gasto",
  "description": "Netflix Premium",
  "amount": 5000.00,
  "currency": "ARS",
  "expenseType": "recurring",
  "date": "2025-01-15",
  "endDate": null,
  "category": {
    "id": "uuid-cat",
    "name": "Entretenimiento",
    "color": "#FF5733"
  },
  "familyMember": {
    "id": "uuid-miembro",
    "name": "Papá"
  },
  "recurringInfo": {
    "monthlyAmount": 5000.00,
    "activeMonths": "indefinido",
    "projectedTotal12Months": 60000.00
  },
  "createdAt": "2025-01-12T15:30:00Z",
  "updatedAt": "2025-01-12T15:30:00Z"
}
```

### Actualizar Gasto

Endpoint: `PUT /api/expenses/:expenseId`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "description": "Netflix Premium + Extra",
  "amount": 6000.00,
  "currency": "ARS",
  "date": "2025-01-15",
  "endDate": "2025-12-31",
  "categoryId": "uuid-nueva-categoria"
}
```

Validaciones:
- El gasto debe pertenecer a la cuenta especificada
- No se puede cambiar expenseType (no se puede convertir one-time en recurring o viceversa)
- Todas las demás validaciones de creación aplican

Proceso:
1. Validar datos
2. Actualizar campos permitidos
3. Actualizar updated_at
4. Retornar gasto actualizado

### Eliminar Gasto

Endpoint: `DELETE /api/expenses/:expenseId`

Headers: Authorization requerido, X-Account-ID requerido

Validaciones:
- El gasto debe pertenecer a la cuenta especificada

Proceso:
1. Eliminar registro de expenses
2. Retornar confirmación

Response exitoso (200 OK):
```json
{
  "message": "Gasto eliminado exitosamente",
  "deletedExpenseId": "uuid-del-gasto"
}
```

### Calcular Compromisos Mensuales

Endpoint: `GET /api/expenses/commitments?month=2025-01`

Headers: Authorization requerido, X-Account-ID requerido

Este endpoint es crítico porque calcula el total de compromisos recurrentes activos en un mes específico.

Query params:
- `month` (opcional): Formato YYYY-MM. Default: mes actual

Proceso:
1. Buscar todos los expenses con expense_type='recurring' de la cuenta
2. Filtrar solo los que están activos en el mes solicitado (date <= mes AND (endDate >= mes OR endDate IS NULL))
3. Sumar los amounts
4. Agrupar por categoría si las tienen
5. Retornar breakdown detallado

Response exitoso (200 OK):
```json
{
  "month": "2025-01",
  "commitments": [
    {
      "id": "uuid-1",
      "description": "Netflix Premium",
      "amount": 5000.00,
      "currency": "ARS",
      "date": "2025-01-15",
      "endDate": null,
      "category": "Entretenimiento"
    },
    {
      "id": "uuid-2",
      "description": "Gimnasio",
      "amount": 8000.00,
      "currency": "ARS",
      "date": "2025-01-01",
      "endDate": "2025-06-30",
      "category": "Salud"
    }
  ],
  "summary": {
    "totalMonthly": 13000.00,
    "count": 2,
    "byCategory": {
      "Entretenimiento": 5000.00,
      "Salud": 8000.00
    },
    "projectedAnnual": 156000.00
  }
}
```

---

## Módulo de Ingresos

El módulo de ingresos funciona de manera muy similar al de gastos, con la misma distinción entre ingresos puntuales y recurrentes. La diferencia principal es el concepto: mientras los gastos son salidas de dinero, los ingresos son entradas.

### Conceptos Fundamentales

**Ingreso Puntual (one-time)**: Dinero que entra una sola vez. Ejemplos: venta de un artículo, bono único, reembolso, regalo en efectivo. Estos ingresos se registran en el mes en que ocurrieron.

**Ingreso Recurrente (recurring)**: Flujo de dinero que se repite mensualmente. Ejemplos: sueldo fijo, proyecto freelance que paga mensualmente, alquiler de propiedad, pensión. Los ingresos recurrentes pueden tener duración definida (un contrato freelance por 6 meses) o ser indefinidos (sueldo permanente).

La importancia de distinguir ingresos recurrentes es poder proyectar entradas futuras y planificar ahorro. Si sabés que tenés $200,000 de ingreso mensual garantizado más $50,000 de proyecto freelance hasta Junio, podés planificar mejor tus metas de ahorro y gastos.

### Crear Ingreso Puntual

Endpoint: `POST /api/incomes`

Headers: Authorization requerido, X-Account-ID requerido

Request body para cuenta personal:
```json
{
  "description": "Venta de notebook usada",
  "amount": 150000.00,
  "currency": "ARS",
  "incomeType": "one-time",
  "date": "2025-01-10",
  "categoryId": "uuid-categoria-opcional"
}
```

Request body para cuenta familiar:
```json
{
  "description": "Bono por desempeño",
  "amount": 50000.00,
  "currency": "ARS",
  "incomeType": "one-time",
  "date": "2025-01-15",
  "familyMemberId": "uuid-miembro-obligatorio",
  "categoryId": "uuid-categoria-opcional"
}
```

Validaciones:
- description no puede estar vacío
- amount debe ser mayor que cero
- currency debe ser 'ARS' o 'USD'
- incomeType debe ser 'one-time'
- date debe ser fecha válida
- Si account.type es 'family', familyMemberId es obligatorio
- Si account.type es 'personal', familyMemberId debe ser null
- Si categoryId existe, debe ser una categoría type='income' de la misma cuenta

Proceso:
1. Validar datos
2. Crear registro en incomes con income_type='one-time' y end_date=null
3. Retornar ingreso creado

Response exitoso (201 Created):
```json
{
  "id": "uuid-del-ingreso",
  "description": "Venta de notebook usada",
  "amount": 150000.00,
  "currency": "ARS",
  "incomeType": "one-time",
  "date": "2025-01-10",
  "category": {
    "id": "uuid-categoria",
    "name": "Ventas"
  },
  "createdAt": "2025-01-12T10:00:00Z"
}
```

### Crear Ingreso Recurrente

Endpoint: `POST /api/incomes`

Headers: Authorization requerido, X-Account-ID requerido

Request body para recurrente indefinido (sueldo permanente):
```json
{
  "description": "Sueldo mensual",
  "amount": 200000.00,
  "currency": "ARS",
  "incomeType": "recurring",
  "date": "2025-01-01",
  "endDate": null,
  "familyMemberId": "uuid-miembro-si-familia",
  "categoryId": "uuid-categoria-opcional"
}
```

Request body para recurrente con duración (proyecto temporal):
```json
{
  "description": "Proyecto freelance React",
  "amount": 1500.00,
  "currency": "USD",
  "incomeType": "recurring",
  "date": "2025-01-01",
  "endDate": "2025-06-30",
  "familyMemberId": "uuid-miembro-si-familia"
}
```

Validaciones:
- Todas las validaciones de ingreso puntual aplican
- incomeType debe ser 'recurring'
- date es la fecha de inicio del ingreso recurrente
- endDate puede ser null (indefinido) o fecha futura mayor a date
- Si endDate existe, debe ser al menos 1 mes después de date

Proceso:
1. Validar datos
2. Crear registro en incomes con income_type='recurring'
3. Este ingreso se considera activo en todos los meses entre date y endDate
4. Retornar ingreso creado

Response exitoso (201 Created):
```json
{
  "id": "uuid-del-ingreso",
  "description": "Proyecto freelance React",
  "amount": 1500.00,
  "currency": "USD",
  "incomeType": "recurring",
  "date": "2025-01-01",
  "endDate": "2025-06-30",
  "monthlyImpact": 1500.00,
  "activeMonths": 6,
  "totalProjected": 9000.00,
  "createdAt": "2025-01-12T10:00:00Z"
}
```

### Listar Ingresos

Endpoint: `GET /api/incomes?month=2025-01&type=all`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `month` (opcional): Formato YYYY-MM. Default: mes actual
- `type` (opcional): 'one-time', 'recurring', o 'all'. Default: 'all'
- `categoryId` (opcional): UUID de categoría
- `familyMemberId` (opcional): UUID de miembro (solo family)
- `currency` (opcional): 'ARS', 'USD', o 'all'. Default: 'all'

Lógica de filtrado por mes:
- Ingresos one-time: incluir si date está en el mes
- Ingresos recurring: incluir si el mes está entre date y endDate (o endDate es null y mes >= date)

Response exitoso (200 OK):
```json
{
  "month": "2025-01",
  "incomes": [
    {
      "id": "uuid-1",
      "description": "Sueldo mensual",
      "amount": 200000.00,
      "currency": "ARS",
      "incomeType": "recurring",
      "date": "2025-01-01",
      "endDate": null,
      "category": {
        "id": "uuid-cat",
        "name": "Trabajo"
      },
      "familyMember": {
        "id": "uuid-miembro",
        "name": "Papá"
      }
    },
    {
      "id": "uuid-2",
      "description": "Venta notebook",
      "amount": 150000.00,
      "currency": "ARS",
      "incomeType": "one-time",
      "date": "2025-01-10"
    }
  ],
  "summary": {
    "totalOneTime": 150000.00,
    "totalRecurring": 200000.00,
    "total": 350000.00,
    "count": 2,
    "byType": {
      "one-time": { "count": 1, "total": 150000.00 },
      "recurring": { "count": 1, "total": 200000.00 }
    }
  }
}
```

### Obtener Detalle de Ingreso

Endpoint: `GET /api/incomes/:incomeId`

Headers: Authorization requerido, X-Account-ID requerido

Similar al detalle de gasto pero con datos de ingreso.

### Actualizar Ingreso

Endpoint: `PUT /api/incomes/:incomeId`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "description": "Sueldo mensual + adicional",
  "amount": 220000.00,
  "date": "2025-01-01",
  "endDate": null
}
```

Validaciones:
- El ingreso debe pertenecer a la cuenta
- No se puede cambiar incomeType
- Demás validaciones de creación aplican

### Eliminar Ingreso

Endpoint: `DELETE /api/incomes/:incomeId`

Headers: Authorization requerido, X-Account-ID requerido

Proceso similar a eliminar gasto.

### Proyectar Ingresos Futuros

Endpoint: `GET /api/incomes/projections?months=6`

Headers: Authorization requerido, X-Account-ID requerido

Este endpoint es crítico para planificación. Proyecta ingresos esperados en los próximos N meses basándose en ingresos recurrentes activos.

Query params:
- `months` (opcional): Cuántos meses hacia adelante proyectar. Default: 6, max: 24

Proceso:
1. Obtener todos los ingresos recurrentes activos
2. Para cada mes futuro hasta months, calcular qué ingresos estarán activos
3. Sumar proyecciones mensuales
4. Retornar breakdown mes a mes

Response exitoso (200 OK):
```json
{
  "projections": [
    {
      "month": "2025-02",
      "incomes": [
        {
          "description": "Sueldo mensual",
          "amount": 200000.00,
          "currency": "ARS"
        },
        {
          "description": "Proyecto freelance",
          "amount": 1500.00,
          "currency": "USD"
        }
      ],
      "totalARS": 200000.00,
      "totalUSD": 1500.00,
      "totalConvertedARS": 350000.00
    },
    {
      "month": "2025-03",
      "incomes": [
        {
          "description": "Sueldo mensual",
          "amount": 200000.00,
          "currency": "ARS"
        }
      ],
      "totalARS": 200000.00,
      "totalUSD": 0,
      "totalConvertedARS": 200000.00
    }
  ],
  "summary": {
    "averageMonthly": 275000.00,
    "total6Months": 1650000.00
  }
}
```

---

## Módulo de Ahorros

El módulo de ahorros convierte objetivos abstractos en planes concretos con números específicos y fechas. Cada cuenta tiene automáticamente una meta de "Ahorro General" más cualquier meta específica que el usuario cree.

### Conceptos Fundamentales

**Meta de Ahorro**: Un objetivo financiero con monto target. Puede tener deadline (fecha límite) o ser indefinido. Las metas con deadline calculan automáticamente cuánto necesitás ahorrar mensualmente. Las metas sin deadline son objetivos de largo plazo sin presión temporal.

**Entrada de Ahorro**: Cada vez que ahorrás dinero para una meta, se registra como una entrada. Las entradas suman al progreso de la meta. El current_amount de una meta es la suma de todas sus entradas.

**Meta de Ahorro General**: Cada cuenta tiene una meta especial llamada "Ahorro General" que se crea automáticamente. Esta meta no tiene deadline y captura ahorros que no están destinados a objetivos específicos. Es el default cuando alguien quiere "guardar plata" sin un propósito definido.

### Crear Meta de Ahorro

Endpoint: `POST /api/savings-goals`

Headers: Authorization requerido, X-Account-ID requerido

Request body para meta con deadline:
```json
{
  "name": "Vacaciones en Brasil",
  "targetAmount": 300000.00,
  "currency": "ARS",
  "deadline": "2025-07-15"
}
```

Request body para meta sin deadline:
```json
{
  "name": "Fondo de Emergencia",
  "targetAmount": 10000.00,
  "currency": "USD",
  "deadline": null
}
```

Validaciones:
- name no puede estar vacío
- targetAmount debe ser mayor que cero
- currency debe ser 'ARS' o 'USD'
- deadline puede ser null o una fecha futura
- name no puede ser "Ahorro General" (reservado para la meta automática)

Proceso:
1. Validar datos
2. Crear registro en savings_goals con current_amount=0 e is_general=false
3. Si tiene deadline, calcular required_monthly_savings
4. Retornar meta creada

Cálculo de required_monthly_savings:
```
meses_restantes = diferencia en meses entre hoy y deadline
required_monthly_savings = (targetAmount - currentAmount) / meses_restantes
```

Si meses_restantes es 0 o negativo (deadline pasó), required_monthly_savings se marca como "deadline vencido".

Response exitoso (201 Created):
```json
{
  "id": "uuid-de-meta",
  "name": "Vacaciones en Brasil",
  "targetAmount": 300000.00,
  "currentAmount": 0,
  "currency": "ARS",
  "deadline": "2025-07-15",
  "isGeneral": false,
  "progress": 0,
  "requiredMonthlySavings": 50000.00,
  "monthsRemaining": 6,
  "createdAt": "2025-01-12T10:00:00Z"
}
```

### Listar Metas de Ahorro

Endpoint: `GET /api/savings-goals`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `includeGeneral` (opcional): true o false. Default: true
- `status` (opcional): 'active', 'completed', 'all'. Default: 'active'
  - 'active': metas con currentAmount < targetAmount
  - 'completed': metas con currentAmount >= targetAmount
  - 'all': todas las metas

Response exitoso (200 OK):
```json
{
  "goals": [
    {
      "id": "uuid-1",
      "name": "Ahorro General",
      "targetAmount": 1000000.00,
      "currentAmount": 150000.00,
      "currency": "ARS",
      "deadline": null,
      "isGeneral": true,
      "progress": 15.0,
      "requiredMonthlySavings": null
    },
    {
      "id": "uuid-2",
      "name": "Vacaciones en Brasil",
      "targetAmount": 300000.00,
      "currentAmount": 50000.00,
      "currency": "ARS",
      "deadline": "2025-07-15",
      "isGeneral": false,
      "progress": 16.67,
      "requiredMonthlySavings": 41666.67,
      "monthsRemaining": 6
    },
    {
      "id": "uuid-3",
      "name": "Fondo de Emergencia",
      "targetAmount": 10000.00,
      "currentAmount": 2500.00,
      "currency": "USD",
      "deadline": null,
      "isGeneral": false,
      "progress": 25.0,
      "requiredMonthlySavings": null
    }
  ],
  "summary": {
    "totalGoals": 3,
    "totalSavedARS": 200000.00,
    "totalSavedUSD": 2500.00,
    "averageProgress": 18.89
  }
}
```

### Obtener Detalle de Meta

Endpoint: `GET /api/savings-goals/:goalId`

Headers: Authorization requerido, X-Account-ID requerido

Response exitoso (200 OK):
```json
{
  "id": "uuid-de-meta",
  "name": "Vacaciones en Brasil",
  "targetAmount": 300000.00,
  "currentAmount": 50000.00,
  "currency": "ARS",
  "deadline": "2025-07-15",
  "isGeneral": false,
  "progress": 16.67,
  "requiredMonthlySavings": 41666.67,
  "monthsRemaining": 6,
  "entries": [
    {
      "id": "uuid-entry-1",
      "amount": 30000.00,
      "date": "2025-01-05",
      "notes": "Primer ahorro del año",
      "familyMember": {
        "id": "uuid-miembro",
        "name": "Papá"
      }
    },
    {
      "id": "uuid-entry-2",
      "amount": 20000.00,
      "date": "2025-01-10",
      "notes": null,
      "familyMember": {
        "id": "uuid-miembro",
        "name": "Mamá"
      }
    }
  ],
  "stats": {
    "totalEntries": 2,
    "averageEntry": 25000.00,
    "remainingToGoal": 250000.00,
    "projectedCompletionDate": "2025-07-15"
  },
  "createdAt": "2025-01-01T10:00:00Z"
}
```

### Actualizar Meta

Endpoint: `PUT /api/savings-goals/:goalId`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "name": "Vacaciones en Brasil - Verano 2025",
  "targetAmount": 350000.00,
  "deadline": "2025-07-30"
}
```

Validaciones:
- La meta debe pertenecer a la cuenta
- No se puede actualizar la meta general (is_general=true)
- name no puede estar vacío
- targetAmount debe ser mayor que cero y mayor o igual a currentAmount
- deadline puede ser null o fecha futura

Proceso:
1. Validar datos
2. Actualizar campos
3. Recalcular requiredMonthlySavings si tiene deadline
4. Retornar meta actualizada

### Eliminar Meta

Endpoint: `DELETE /api/savings-goals/:goalId`

Headers: Authorization requerido, X-Account-ID requerido

Validaciones:
- La meta debe pertenecer a la cuenta
- No se puede eliminar la meta general (is_general=true)
- El usuario debe confirmar si la meta tiene entradas de ahorro

Proceso:
1. Si tiene entradas, el usuario debe confirmar explícitamente
2. Eliminar todas las savings_entries asociadas
3. Eliminar la meta
4. Retornar confirmación

Response exitoso (200 OK):
```json
{
  "message": "Meta eliminada exitosamente",
  "deletedGoalId": "uuid-de-meta",
  "deletedEntriesCount": 5
}
```

### Agregar Entrada de Ahorro

Endpoint: `POST /api/savings-goals/:goalId/entries`

Headers: Authorization requerido, X-Account-ID requerido

Request body para cuenta personal:
```json
{
  "amount": 30000.00,
  "date": "2025-01-12",
  "notes": "Ahorro de este mes"
}
```

Request body para cuenta familiar:
```json
{
  "amount": 20000.00,
  "date": "2025-01-12",
  "familyMemberId": "uuid-miembro-obligatorio",
  "notes": "Mi aporte del mes"
}
```

Validaciones:
- amount debe ser mayor que cero
- La moneda de la entrada se toma de savings_goal.currency automáticamente
- date no puede ser futura
- Si account.type es 'family', familyMemberId es obligatorio
- Si account.type es 'personal', familyMemberId debe ser null
- notes es opcional

Proceso:
1. Validar datos
2. Crear registro en savings_entries
3. Actualizar savings_goals.current_amount sumando este amount
4. Recalcular requiredMonthlySavings si la meta tiene deadline
5. Retornar entrada creada

Response exitoso (201 Created):
```json
{
  "id": "uuid-de-entrada",
  "amount": 30000.00,
  "currency": "ARS",
  "date": "2025-01-12",
  "notes": "Ahorro de este mes",
  "familyMember": null,
  "goal": {
    "id": "uuid-de-meta",
    "name": "Vacaciones en Brasil",
    "currentAmount": 80000.00,
    "progress": 26.67
  },
  "createdAt": "2025-01-12T15:00:00Z"
}
```

### Listar Entradas de Ahorro

Endpoint: `GET /api/savings-goals/:goalId/entries`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `startDate` (opcional): Formato YYYY-MM-DD
- `endDate` (opcional): Formato YYYY-MM-DD
- `familyMemberId` (opcional): UUID de miembro

Response exitoso (200 OK):
```json
{
  "goalId": "uuid-de-meta",
  "goalName": "Vacaciones en Brasil",
  "entries": [
    {
      "id": "uuid-entry-1",
      "amount": 30000.00,
      "date": "2025-01-12",
      "notes": "Ahorro de este mes",
      "familyMember": {
        "id": "uuid-miembro",
        "name": "Papá"
      },
      "createdAt": "2025-01-12T15:00:00Z"
    }
  ],
  "summary": {
    "totalAmount": 30000.00,
    "count": 1,
    "averageAmount": 30000.00
  }
}
```

### Eliminar Entrada de Ahorro

Endpoint: `DELETE /api/savings-goals/:goalId/entries/:entryId`

Headers: Authorization requerido, X-Account-ID requerido

Proceso:
1. Validar que la entrada pertenece a la meta y cuenta correctas
2. Eliminar la entrada
3. Actualizar savings_goals.current_amount restando el amount de esta entrada
4. Recalcular requiredMonthlySavings si corresponde
5. Retornar confirmación

Response exitoso (200 OK):
```json
{
  "message": "Entrada eliminada exitosamente",
  "deletedEntryId": "uuid-de-entrada",
  "goal": {
    "id": "uuid-de-meta",
    "currentAmount": 50000.00,
    "progress": 16.67
  }
}
```

### Crear Meta desde Producto de Lista

Endpoint: `POST /api/savings-goals/from-product`

Headers: Authorization requerido, X-Account-ID requerido

Este endpoint es especial porque crea una meta de ahorro directamente desde un producto de la wishlist.

Request body:
```json
{
  "productId": "uuid-del-producto",
  "deadline": "2025-06-30"
}
```

Validaciones:
- El producto debe pertenecer a la cuenta
- El producto debe tener price definido
- deadline puede ser null o fecha futura

Proceso:
1. Buscar el producto
2. Crear meta de ahorro con:
   - name = product.name
   - targetAmount = product.price
   - currency = product.currency
   - deadline = el proporcionado
3. Opcionalmente marcar el producto como "vinculado a meta"
4. Retornar meta creada

### Crear Meta desde Categoría de Wishlist

Endpoint: `POST /api/savings-goals/from-category`

Headers: Authorization requerido, X-Account-ID requerido

Crea una meta de ahorro sumando todos los productos de una categoría de wishlist.

Request body:
```json
{
  "categoryId": "uuid-de-categoria-wishlist",
  "deadline": "2025-12-31"
}
```

Proceso:
1. Buscar todos los productos pending de esa categoría con price definido
2. Sumar todos los prices (convertir a moneda común si hay mezcla)
3. Crear meta con:
   - name = "Categoría: " + category.name
   - targetAmount = suma total
   - currency = moneda más común de los productos
   - deadline = el proporcionado
4. Retornar meta creada con lista de productos incluidos

---

## Módulo de Lista de Compras

La lista de compras es funcionalidad secundaria que permite registrar productos que querés comprar eventualmente. No son gastos inmediatos sino deseos futuros. La característica especial es poder crear metas de ahorro desde productos.

### Crear Producto

Endpoint: `POST /api/wishlist/products`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "name": "Mouse Logitech G502",
  "price": 45000.00,
  "currency": "ARS",
  "url": "https://example.com/mouse",
  "notes": "Vi este mouse en la tienda, tiene RGB",
  "categoryId": "uuid-categoria-opcional"
}
```

Request body sin precio:
```json
{
  "name": "Auriculares buenos",
  "notes": "Todavía no sé qué modelo ni cuánto cuestan"
}
```

Validaciones:
- name no puede estar vacío
- Si price existe, debe ser mayor que cero y currency debe existir
- Si currency existe, price debe existir
- url debe ser URL válida si existe
- categoryId debe ser una categoría de wishlist de la misma cuenta si existe

Proceso:
1. Validar datos
2. Crear registro en wishlist_products con status='pending'
3. Retornar producto creado

Response exitoso (201 Created):
```json
{
  "id": "uuid-del-producto",
  "name": "Mouse Logitech G502",
  "price": 45000.00,
  "currency": "ARS",
  "url": "https://example.com/mouse",
  "notes": "Vi este mouse en la tienda, tiene RGB",
  "category": {
    "id": "uuid-categoria",
    "name": "Tecnología"
  },
  "status": "pending",
  "createdAt": "2025-01-12T10:00:00Z"
}
```

### Listar Productos

Endpoint: `GET /api/wishlist/products?status=pending`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `status` (opcional): 'pending', 'purchased', o 'all'. Default: 'pending'
- `categoryId` (opcional): UUID de categoría wishlist
- `hasPrice` (opcional): true o false para filtrar productos con/sin precio

Response exitoso (200 OK):
```json
{
  "products": [
    {
      "id": "uuid-1",
      "name": "Mouse Logitech G502",
      "price": 45000.00,
      "currency": "ARS",
      "url": "https://example.com/mouse",
      "category": {
        "id": "uuid-cat",
        "name": "Tecnología"
      },
      "status": "pending",
      "createdAt": "2025-01-12T10:00:00Z"
    },
    {
      "id": "uuid-2",
      "name": "Auriculares buenos",
      "price": null,
      "currency": null,
      "category": {
        "id": "uuid-cat",
        "name": "Tecnología"
      },
      "status": "pending",
      "createdAt": "2025-01-11T15:00:00Z"
    }
  ],
  "summary": {
    "totalProducts": 2,
    "totalPriceARS": 45000.00,
    "totalPriceUSD": 0,
    "productsWithoutPrice": 1
  }
}
```

### Actualizar Producto

Endpoint: `PUT /api/wishlist/products/:productId`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "name": "Mouse Logitech G502 HERO",
  "price": 50000.00,
  "currency": "ARS",
  "url": "https://example.com/mouse-hero",
  "notes": "Encontré versión actualizada"
}
```

Validaciones similares a creación.

### Marcar como Comprado

Endpoint: `PUT /api/wishlist/products/:productId/purchase`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "purchasedDate": "2025-01-15",
  "finalPrice": 48000.00,
  "notes": "Comprado en oferta!"
}
```

Validaciones:
- purchasedDate no puede ser futura
- finalPrice es opcional pero si existe debe ser mayor que cero

Proceso:
1. Actualizar status='purchased'
2. Guardar purchased_date
3. Opcionalmente actualizar price con finalPrice si se proporciona
4. Retornar producto actualizado

Response exitoso (200 OK):
```json
{
  "id": "uuid-del-producto",
  "name": "Mouse Logitech G502 HERO",
  "price": 48000.00,
  "currency": "ARS",
  "status": "purchased",
  "purchasedDate": "2025-01-15",
  "notes": "Comprado en oferta!",
  "updatedAt": "2025-01-15T14:00:00Z"
}
```

### Eliminar Producto

Endpoint: `DELETE /api/wishlist/products/:productId`

Headers: Authorization requerido, X-Account-ID requerido

Proceso directo: eliminar el producto y retornar confirmación.

### Crear Categoría de Wishlist

Endpoint: `POST /api/wishlist/categories`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "name": "Tecnología",
  "color": "#3B82F6"
}
```

Validaciones:
- name no puede estar vacío
- No puede existir otra categoría con el mismo name en la cuenta
- color debe ser hex válido si existe

### Listar Categorías de Wishlist

Endpoint: `GET /api/wishlist/categories`

Headers: Authorization requerido, X-Account-ID requerido

Response exitoso (200 OK):
```json
{
  "categories": [
    {
      "id": "uuid-1",
      "name": "Tecnología",
      "color": "#3B82F6",
      "productCount": 5,
      "totalPriceARS": 200000.00,
      "createdAt": "2025-01-01T10:00:00Z"
    }
  ]
}
```

---

## Sistema de Categorización

Las categorías son opcionales pero útiles para análisis detallado. Hay dos sistemas de categorías independientes: uno para gastos/ingresos y otro para la wishlist.

### Categorías de Gastos/Ingresos

Estas categorías tienen jerarquía de dos niveles: categorías padre y subcategorías hijas.

#### Crear Categoría Padre

Endpoint: `POST /api/categories`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "name": "Alimentación",
  "type": "expense",
  "color": "#10B981",
  "icon": "utensils"
}
```

Validaciones:
- name no puede estar vacío
- type debe ser 'expense' o 'income'
- parent_category_id debe ser null (es categoría padre)
- color debe ser hex válido si existe

#### Crear Subcategoría

Endpoint: `POST /api/categories`

Headers: Authorization requerido, X-Account-ID requerido

Request body:
```json
{
  "name": "Supermercado",
  "type": "expense",
  "parentCategoryId": "uuid-de-alimentacion",
  "color": "#10B981",
  "icon": "shopping-cart"
}
```

Validaciones:
- parentCategoryId debe existir y ser del mismo type
- La categoría padre no puede ser a su vez una subcategoría (solo 2 niveles)

#### Listar Categorías

Endpoint: `GET /api/categories?type=expense`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `type` (opcional): 'expense', 'income', o 'all'. Default: 'all'

Response con jerarquía (200 OK):
```json
{
  "categories": [
    {
      "id": "uuid-1",
      "name": "Alimentación",
      "type": "expense",
      "color": "#10B981",
      "icon": "utensils",
      "subcategories": [
        {
          "id": "uuid-sub-1",
          "name": "Supermercado",
          "type": "expense",
          "color": "#10B981",
          "icon": "shopping-cart"
        },
        {
          "id": "uuid-sub-2",
          "name": "Restaurantes",
          "type": "expense",
          "color": "#10B981",
          "icon": "restaurant"
        }
      ]
    },
    {
      "id": "uuid-2",
      "name": "Transporte",
      "type": "expense",
      "color": "#F59E0B",
      "icon": "car",
      "subcategories": []
    }
  ]
}
```

#### Actualizar Categoría

No se puede cambiar el type ni mover una categoría de padre a hijo o viceversa. Solo se pueden actualizar name, color, e icon.

#### Eliminar Categoría

No se puede eliminar una categoría que tiene gastos/ingresos asignados. Tampoco se puede eliminar una categoría padre que tiene subcategorías.

---

## Dashboard y Analytics

El dashboard es donde toda la data cruda se transforma en información útil y accionable. Muestra números clave, tendencias, y análisis que responden preguntas concretas sobre la situación financiera.

### Vista Principal del Dashboard

Endpoint: `GET /api/dashboard?month=2025-01`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `month` (opcional): Formato YYYY-MM. Default: mes actual

Este endpoint retorna un objeto masivo con todas las métricas principales del dashboard.

Response (200 OK):
```json
{
  "month": "2025-01",
  "summary": {
    "totalIncome": 350000.00,
    "totalExpenses": 180000.00,
    "balance": 170000.00,
    "totalSavings": 50000.00,
    "availableToSpend": 120000.00
  },
  "recurringCommitments": {
    "monthlyTotal": 45000.00,
    "count": 5,
    "items": [
      {
        "description": "Netflix",
        "amount": 5000.00,
        "currency": "ARS",
        "category": "Entretenimiento"
      }
    ]
  },
  "savingsGoals": {
    "goals": [
      {
        "id": "uuid-1",
        "name": "Vacaciones",
        "progress": 33.33,
        "currentAmount": 100000.00,
        "targetAmount": 300000.00,
        "requiredMonthlySavings": 40000.00
      }
    ],
    "totalSaved": 150000.00,
    "averageProgress": 35.5
  },
  "expensesByCategory": [
    {
      "category": "Alimentación",
      "amount": 80000.00,
      "percentage": 44.4
    },
    {
      "category": "Transporte",
      "amount": 35000.00,
      "percentage": 19.4
    }
  ],
  "incomesByCategory": [
    {
      "category": "Trabajo",
      "amount": 300000.00,
      "percentage": 85.7
    }
  ],
  "trends": {
    "expensesLast6Months": [
      { "month": "2024-08", "amount": 150000.00 },
      { "month": "2024-09", "amount": 160000.00 }
    ],
    "incomesLast6Months": [
      { "month": "2024-08", "amount": 300000.00 },
      { "month": "2024-09", "amount": 320000.00 }
    ]
  },
  "familyBreakdown": {
    "expenses": [
      { "member": "Papá", "amount": 100000.00, "percentage": 55.6 },
      { "member": "Mamá", "amount": 80000.00, "percentage": 44.4 }
    ],
    "incomes": [
      { "member": "Papá", "amount": 200000.00, "percentage": 57.1 },
      { "member": "Mamá", "amount": 150000.00, "percentage": 42.9 }
    ]
  }
}
```

### Análisis por Categoría

Endpoint: `GET /api/analytics/by-category?type=expense&startDate=2025-01-01&endDate=2025-01-31`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `type`: 'expense' o 'income'
- `startDate`: Formato YYYY-MM-DD
- `endDate`: Formato YYYY-MM-DD

Response con breakdown detallado por categoría incluyendo subcategorías.

### Análisis por Miembro Familiar

Endpoint: `GET /api/analytics/by-member?type=expense&startDate=2025-01-01&endDate=2025-01-31`

Headers: Authorization requerido, X-Account-ID requerido

Solo funciona en cuentas tipo family.

Response con breakdown de gastos/ingresos por miembro, mostrando quién contribuye qué proporción.

### Tendencias Temporales

Endpoint: `GET /api/analytics/trends?months=12`

Headers: Authorization requerido, X-Account-ID requerido

Query params:
- `months` (opcional): Cuántos meses hacia atrás analizar. Default: 6, max: 24

Response con datos mes a mes para graficar tendencias de gastos, ingresos, balance, y ahorro.

---

## Sistema de Monedas

El sistema maneja nativamente ARS y USD sin concepto de "moneda principal". Cada transacción se guarda en su moneda original.

### Obtener Tipo de Cambio Actual

Endpoint: `GET /api/exchange-rates/current`

Response:
```json
{
  "date": "2025-01-12",
  "rates": [
    {
      "from": "USD",
      "to": "ARS",
      "rate": 1050.00,
      "source": "ExchangeRate-API"
    },
    {
      "from": "ARS",
      "to": "USD",
      "rate": 0.000952,
      "source": "ExchangeRate-API"
    }
  ]
}
```

### Convertir Monto

Endpoint: `POST /api/exchange-rates/convert`

Request body:
```json
{
  "amount": 100.00,
  "fromCurrency": "USD",
  "toCurrency": "ARS",
  "date": "2025-01-12"
}
```

Response:
```json
{
  "originalAmount": 100.00,
  "originalCurrency": "USD",
  "convertedAmount": 105000.00,
  "convertedCurrency": "ARS",
  "rate": 1050.00,
  "date": "2025-01-12"
}
```

El parámetro date es opcional y se usa para conversiones históricas. Si no se provee, usa la tasa más reciente.

---

## Reglas de Negocio Transversales

Estas reglas aplican a todo el sistema y deben respetarse en todos los módulos.

### Aislamiento por Cuenta

- Todos los endpoints que manejan datos financieros requieren header X-Account-ID
- El middleware debe validar que la cuenta pertenece al usuario autenticado
- Todas las queries a la base de datos deben filtrar automáticamente por account_id
- Nunca debe ser posible ver o modificar datos de una cuenta diferente a la activa

### Atribución por Miembro Familiar

- En cuentas tipo family, todos los movimientos financieros (gastos, ingresos, entradas de ahorro) requieren family_member_id
- En cuentas tipo personal, family_member_id siempre debe ser null
- Los miembros con is_active=false no aparecen en selectores pero sus datos históricos permanecen visibles

### Categorización Opcional

- Categorizar gastos e ingresos es completamente opcional
- Todas las estadísticas principales funcionan sin categorías
- Solo los análisis específicos por categoría requieren que existan categorías asignadas
- Los movimientos sin categoría no se excluyen de ningún cálculo excepto análisis por categoría

### Manejo de Monedas

- Cada transacción se guarda en su moneda original
- Las conversiones solo ocurren para visualización, nunca alteran datos almacenados
- Para totales consolidados, se debe especificar en qué moneda mostrar el resultado
- Las tasas de cambio se deben actualizar diariamente y almacenar históricamente

### Cálculos de Gastos e Ingresos Recurrentes

- Un gasto/ingreso recurrente se cuenta una sola vez en el mes en que está activo
- Para calcular totales de un mes específico, incluir todos los recurrentes activos en ese mes
- Para proyecciones futuras, considerar todos los recurrentes que estarán activos en cada mes proyectado
- Los recurrentes sin end_date se consideran activos indefinidamente hasta que se modifiquen o eliminen

### Integridad de Datos de Ahorro

- El current_amount de una meta debe ser siempre igual a la suma de sus savings_entries
- Al crear una entrada, incrementar current_amount automáticamente
- Al eliminar una entrada, decrementar current_amount automáticamente
- No permitir que current_amount exceda target_amount al crear entradas

### Validaciones de Fechas

- Las fechas de gastos e ingresos one-time pueden ser pasadas, presentes, o futuras
- Las fechas de entradas de ahorro no pueden ser futuras (no podés ahorrar en el futuro)
- Los deadlines de metas de ahorro deben ser fechas futuras
- Los end_date de gastos/ingresos recurrentes deben ser posteriores a date

---

## Flujos de Usuario Principales

Estos son los flujos completos que un usuario típico ejecuta al usar el sistema.

### Flujo: Registrarse y Configurar Primera Cuenta

1. Usuario visita la app por primera vez
2. Completa formulario de registro con email, password, nombre
3. Sistema crea usuario y lo autentica automáticamente
4. Usuario es redirigido a página de creación de cuenta
5. Usuario selecciona tipo (personal o familiar) y moneda base
6. Si eligió familiar, agrega nombres de miembros
7. Sistema crea cuenta y meta de Ahorro General automáticamente
8. Usuario es redirigido al dashboard (vacío inicialmente)

### Flujo: Registrar Gasto Puntual

1. Usuario clickea botón "Nuevo Gasto"
2. Selecciona tipo "Gasto Único"
3. Completa formulario: descripción, monto, moneda, fecha
4. Si la cuenta es familiar, selecciona qué miembro gastó
5. Opcionalmente selecciona categoría si la tiene configurada
6. Clickea "Guardar"
7. Sistema valida y crea el gasto
8. Usuario ve confirmación y el gasto aparece en lista de gastos del mes

### Flujo: Configurar Gasto Recurrente

1. Usuario clickea "Nuevo Gasto"
2. Selecciona tipo "Gasto Recurrente"
3. Completa descripción (ej: "Netflix Premium"), monto, moneda
4. Selecciona fecha de inicio
5. Decide si tiene fecha de fin o es indefinido:
   - Si indefinido: deja end_date vacío
   - Si temporal: selecciona end_date (ej: en 6 meses)
6. Si la cuenta es familiar, selecciona miembro
7. Clickea "Guardar"
8. Sistema valida y crea el gasto recurrente
9. Este gasto ahora aparece en la sección de "Compromisos Mensuales" del dashboard
10. El sistema automáticamente lo considera en el cálculo de gastos de todos los meses futuros relevantes

### Flujo: Crear Meta de Ahorro con Deadline

1. Usuario clickea "Nueva Meta de Ahorro"
2. Ingresa nombre (ej: "Vacaciones en Brasil")
3. Ingresa monto objetivo (ej: $300,000)
4. Selecciona moneda
5. Selecciona fecha límite (ej: 6 meses en el futuro)
6. Clickea "Crear Meta"
7. Sistema calcula automáticamente ahorro mensual requerido: (300,000 - 0) / 6 = $50,000 por mes
8. Meta aparece en dashboard mostrando progreso 0% y "Necesitás ahorrar $50,000 por mes"

### Flujo: Ahorrar para una Meta

1. Usuario entra a detalle de una meta (ej: "Vacaciones")
2. Clickea "Agregar Ahorro"
3. Ingresa monto que está ahorrando ahora (ej: $50,000)
4. Selecciona fecha de hoy
5. Opcionalmente agrega nota (ej: "Ahorro de enero")
6. Si cuenta familiar, selecciona qué miembro ahorró
7. Clickea "Guardar"
8. Sistema crea entrada y actualiza current_amount de la meta a $50,000
9. Progreso se actualiza a 16.67% ($50,000 de $300,000)
10. Ahorro mensual requerido se recalcula: (300,000 - 50,000) / 5 meses restantes = $50,000 por mes

### Flujo: Crear Meta desde Producto de Wishlist

1. Usuario agrega producto a wishlist: "Mouse Logitech G502", $45,000 ARS
2. Desde detalle del producto, clickea "Crear Meta de Ahorro"
3. Sistema pre-completa formulario:
   - Nombre: "Mouse Logitech G502"
   - Monto: $45,000
   - Moneda: ARS
4. Usuario solo necesita elegir deadline (ej: 3 meses)
5. Clickea "Crear"
6. Meta se crea automáticamente calculando ahorro mensual: $45,000 / 3 = $15,000 por mes

### Flujo: Analizar Gastos del Mes

1. Usuario entra al dashboard
2. Ve card principal: "Este mes gastaste $180,000"
3. Scroll hacia abajo ve breakdown por categoría:
   - Alimentación: $80,000 (44%)
   - Transporte: $35,000 (19%)
   - Entretenimiento: $25,000 (14%)
4. Ve sección "Compromisos Recurrentes": $45,000 mensuales comprometidos en 5 servicios
5. Si cuenta familiar, ve análisis por miembro: Papá $100,000 (55%), Mamá $80,000 (45%)
6. Clickea gráfico de tendencias y ve cómo los gastos evolucionaron últimos 6 meses

### Flujo: Cambiar de Cuenta

1. Usuario está viendo cuenta "Finanzas Personales"
2. Clickea selector de cuenta en header
3. Ve lista de sus cuentas con íconos indicando tipo
4. Selecciona "Gastos Familia"
5. Sistema inmediatamente recarga toda la interfaz con datos de cuenta familiar
6. Ahora todos los formularios piden seleccionar miembro familiar
7. Dashboard muestra datos completamente diferentes

---

## Roadmap de Implementación Detallado

Esta sección desglosa el desarrollo en fases manejables con estimaciones realistas de tiempo.

### Fase 1: Setup Inicial y Autenticación (3-4 días)

**Backend:**
- Configurar proyecto Go con estructura de carpetas
- Setup de PostgreSQL local con base de datos de desarrollo
- Implementar tabla users
- Implementar endpoints de register, login, refresh, logout
- Implementar JWT con access y refresh tokens
- Crear middleware de autenticación
- Testing manual con Postman/Insomnia

**Frontend:**
- Configurar proyecto React con Vite
- Setup de Tailwind CSS
- Instalar y configurar shadcn/ui
- Implementar páginas de login y registro
- Implementar almacenamiento de tokens
- Implementar auto-refresh de access token
- Crear hook useAuth para gestionar autenticación
- Configurar proxy de Vite hacia backend en desarrollo

Al final de esta fase tenés autenticación funcional y podés registrarte, hacer login, y el sistema mantiene tu sesión.

### Fase 2: Sistema de Cuentas (4-5 días)

**Backend:**
- Implementar tabla accounts
- Implementar tabla family_members
- Implementar endpoints de crear, listar, actualizar, eliminar cuentas
- Implementar endpoints de gestionar miembros familiares
- Implementar middleware que valida account_id en header y verifica ownership
- Al crear cuenta, auto-crear meta de Ahorro General

**Frontend:**
- Implementar página de creación de cuenta con wizard para configurar miembros
- Implementar selector de cuenta en header que siempre muestra cuenta activa
- Implementar context o store (TanStack Query context) para cuenta activa
- Implementar cambio de cuenta con recarga de datos
- Crear componentes de formulario reutilizables para account y members

Al final de esta fase podés crear múltiples cuentas, cambiar entre ellas, y el sistema aísla datos correctamente.

### Fase 3: Módulo de Gastos (5-6 días)

**Backend:**
- Implementar tabla expenses
- Implementar endpoints de crear gasto puntual
- Implementar endpoints de crear gasto recurrente
- Implementar endpoint de listar gastos con filtros por mes, tipo, categoría, miembro
- Implementar endpoint de calcular compromisos recurrentes mensuales
- Implementar endpoints de actualizar y eliminar gastos

**Frontend:**
- Implementar página de lista de gastos con filtros
- Implementar formulario de nuevo gasto con toggle entre puntual y recurrente
- Implementar lógica condicional: si cuenta es familiar, mostrar selector de miembro
- Implementar componente de card para gasto individual mostrando tipo, monto, fecha
- Implementar vista de compromisos recurrentes mostrando impacto mensual
- Usar TanStack Query para fetching, caching, y mutaciones

Al final de esta fase el módulo de gastos está completo y funcional.

### Fase 4: Módulo de Ingresos (4-5 días)

**Backend:**
- Implementar tabla incomes (muy similar a expenses)
- Implementar endpoints de crear ingreso puntual
- Implementar endpoints de crear ingreso recurrente
- Implementar endpoint de listar ingresos con filtros
- Implementar endpoint de proyectar ingresos futuros
- Implementar endpoints de actualizar y eliminar ingresos

**Frontend:**
- Implementar página de lista de ingresos (reutilizar componentes de gastos donde sea posible)
- Implementar formulario de nuevo ingreso
- Implementar vista de proyección de ingresos futuros con breakdown mes a mes
- Usar componentes compartidos para reducir duplicación de código

Al final de esta fase tenés gestión completa de ingresos funcionando igual que gastos.

### Fase 5: Módulo de Ahorros (5-6 días)

**Backend:**
- Implementar tabla savings_goals
- Implementar tabla savings_entries
- Implementar endpoint de crear meta (validar que no se puede llamar "Ahorro General")
- Implementar lógica de auto-crear meta general al crear cuenta
- Implementar endpoint de listar metas con cálculo de required_monthly_savings
- Implementar endpoint de agregar entrada de ahorro
- Implementar actualización automática de current_amount al crear/eliminar entradas
- Implementar endpoints de actualizar y eliminar metas
- Implementar endpoint de listar entradas de una meta específica

**Frontend:**
- Implementar página de metas de ahorro con cards mostrando progreso visual
- Implementar formulario de crear meta con toggle de deadline
- Para metas con deadline, mostrar cálculo de ahorro mensual requerido
- Implementar vista de detalle de meta mostrando todas las entradas
- Implementar formulario de agregar entrada de ahorro
- Implementar barras de progreso visuales para cada meta
- Highlight especial para meta general

Al final de esta fase el sistema tiene los tres módulos core funcionando juntos.

### Fase 6: Lista de Compras (3-4 días)

**Backend:**
- Implementar tabla wishlist_products
- Implementar tabla wishlist_categories
- Implementar endpoints CRUD de productos
- Implementar endpoints CRUD de categorías wishlist
- Implementar endpoint de marcar producto como comprado
- Implementar endpoint de crear meta desde producto
- Implementar endpoint de crear meta desde categoría

**Frontend:**
- Implementar página de wishlist con lista de productos
- Implementar filtros por status (pending/purchased) y categoría
- Implementar formulario de agregar producto (precio opcional)
- Implementar botón "Crear Meta" en producto que pre-completa formulario de meta
- Implementar vista de productos purchased que se ven diferentes visualmente

Esta fase es relativamente rápida porque extiende funcionalidad existente.

### Fase 7: Sistema de Categorización (3-4 días)

**Backend:**
- Implementar tabla categories con parent_category_id
- Implementar endpoints de crear categoría padre
- Implementar endpoints de crear subcategoría
- Implementar validación de solo 2 niveles de jerarquía
- Implementar endpoint de listar categorías con estructura jerárquica
- Implementar endpoints de actualizar y eliminar (con validaciones de no eliminar si tiene movimientos)

**Frontend:**
- Implementar página de gestión de categorías mostrando jerarquía visualmente
- Implementar formulario de crear categoría con opción de hacerla padre o hijo
- En formularios de gastos/ingresos, agregar selector opcional de categoría
- Implementar selector jerárquico (padre > hijo) para mejor UX
- Asegurar que todo funciona igual sin categorías (son opcionales)

### Fase 8: Dashboard y Analytics (6-7 días)

**Backend:**
- Implementar endpoint masivo de dashboard que retorna todas las métricas principales
- Implementar cálculos de balance (ingresos - gastos - ahorros)
- Implementar análisis por categoría con porcentajes
- Implementar análisis por miembro familiar
- Implementar endpoint de tendencias temporales
- Optimizar queries para performance (usar JOINs eficientes, indexes, caching si es necesario)

**Frontend:**
- Implementar dashboard principal con layout de cards
- Card principal: balance del mes
- Card de compromisos recurrentes con lista expandible
- Card de metas de ahorro mostrando progreso visual
- Gráfico de pie para distribución de gastos por categoría usando Recharts
- Gráfico de línea para tendencias temporales de gastos e ingresos
- Si cuenta familiar, sección de breakdown por miembro con gráficos
- Hacer todo responsive para mobile

Esta fase requiere trabajo considerable de diseño visual y optimización de performance.

### Fase 9: Sistema de Monedas (2-3 días)

**Backend:**
- Implementar tabla exchange_rates
- Implementar job/cron que actualiza tasas diariamente desde API externa (ExchangeRate-API es gratuita hasta 1000 requests/mes)
- Implementar endpoint de obtener tasa actual
- Implementar endpoint de convertir montos
- Implementar lógica de usar tasa más cercana para conversiones históricas

**Frontend:**
- En dashboard y analytics, agregar toggle para ver totales en ARS o USD
- Cuando se muestra total consolidado, aplicar conversión transparente
- Mostrar tasa de cambio usada en tooltips
- Para transacciones individuales, siempre mostrar en moneda original

### Fase 10: Polish y Refinamiento (4-5 días)

**General:**
- Optimización de performance: revisar queries lentas, agregar indexes necesarios
- Implementar loading states elegantes en toda la app
- Implementar manejo de errores robusto con mensajes claros al usuario
- Implementar validaciones en frontend que repliquen validaciones de backend
- Hacer todo responsive: tablet y mobile
- Testing extensivo de edge cases
- Pulido visual: consistencia de colores, espaciados, tipografías
- Implementar confirmaciones antes de acciones destructivas (eliminar cuenta, eliminar meta con entradas)
- Documentar código complejo
- Limpiar código: remover console.logs, comentarios innecesarios, código muerto

Al final de esta fase el sistema está listo para uso real.

### Fases Futuras Opcionales

Estas son mejoras post-MVP que pueden agregarse después:

**Presupuestos (1 semana):**
- Tabla budgets con límites por categoría
- Alertas cuando te acercás o excedés límite
- Vista de presupuesto vs real

**Notificaciones (3-4 días):**
- Sistema de notificaciones in-app
- Opcionalmente push notifications si se hace mobile app

**Exportación de Datos (2-3 días):**
- Exportar gastos/ingresos a CSV
- Exportar reporte completo a PDF
- Backup completo de cuenta en JSON

**Multi-moneda Extendida (2-3 días):**
- Agregar EUR, GBP, y otras monedas
- Mantener mismo patrón de ARS/USD

**Modo Offline (1 semana):**
- Service workers para PWA
- IndexedDB para almacenamiento local
- Sincronización cuando vuelve conexión

### Estimación Total

**Tiempo mínimo (desarrollador experimentado, full-time, sin interrupciones):** 6-7 semanas

**Tiempo realista (desarrollador con otras responsabilidades, part-time 10-15 horas semanales):** 3-4 meses

**Tiempo conservador (considerando aprendizaje de Go, problemas inesperados, iteraciones de diseño):** 4-5 meses

Estas estimaciones asumen desarrollo sin diseño gráfico profesional, usando componentes funcionales de shadcn/ui directamente.

---

## Conclusión

Este documento contiene las especificaciones técnicas completas para implementar Bolsillo Claro. Cada módulo está detallado con:
- Estructura de datos precisa
- Endpoints de API con request/response examples
- Validaciones específicas que deben aplicarse
- Reglas de negocio que deben respetarse
- Casos de uso concretos con ejemplos reales
- Flujos de usuario completos
- Roadmap de implementación realista

El sistema es ambicioso pero completamente factible. La arquitectura está diseñada para ser simple donde sea posible, compleja solo donde sea necesario. El aislamiento por cuenta, la categorización opcional, y el soporte nativo de múltiples monedas son los diferenciadores clave que hacen que este sistema sea útil para la realidad argentina.

Usá este documento como referencia durante la implementación. Cuando tengas dudas sobre cómo debe comportarse una feature, volvé a este documento. Si encontrás casos edge que no están cubiertos, documentalos acá para referencia futura.
