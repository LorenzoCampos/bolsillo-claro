# Bolsillo Claro

Bolsillo Claro es un gestor financiero familiar construido para mi uso personal y el de mi familia. El sistema permite trackear gastos recurrentes y puntuales, registrar ingresos fijos y variables, establecer metas de ahorro con cálculo automático de progreso, y mantener una lista de compras futura. Todo diseñado específicamente para la realidad argentina, con soporte nativo para pesos y dólares.

El proyecto nació de una necesidad personal: quería claridad sobre cuánto gasto realmente, cuánto ingresa, y cuánto estoy ahorrando. Las aplicaciones existentes no modelaban bien la realidad de ingresos variables, gastos recurrentes que se pierden de vista, o la convivencia entre múltiples monedas. Bolsillo Claro resuelve estos problemas de forma directa.

## Características Principales

El sistema está organizado en torno a cuatro módulos que trabajan juntos para dar una visión completa de la situación financiera.

**Gestión de Gastos** permite registrar tanto compras puntuales como compromisos recurrentes que se repiten automáticamente cada mes. Los gastos recurrentes son especialmente importantes porque muchos pagos mensuales (servicios de streaming, gimnasio, seguros) pasan desapercibidos hasta que los sumás y descubrís que tenés miles de pesos comprometidos cada mes.

**Tracking de Ingresos** reconoce que no todos tienen un sueldo fijo. El sistema maneja ingresos recurrentes con duración configurable (un proyecto freelance que paga mensualmente por seis meses), ingresos fijos permanentes (un sueldo tradicional), e ingresos puntuales (trabajos cortos o bonos ocasionales).

**Metas de Ahorro** convierte el ahorro de un concepto abstracto en un plan concreto. Cuando querés juntar trescientos mil pesos para unas vacaciones en seis meses, el sistema calcula automáticamente que necesitás ahorrar cincuenta mil por mes. Las metas pueden tener deadline o ser indefinidas, y el progreso se actualiza automáticamente a medida que agregás entradas de ahorro.

**Lista de Compras** funciona como un registro de productos que querés comprar eventualmente, sin fecha definida. No es para gastos inmediatos sino para ese equipo que te gustó, esa herramienta que necesitás cuando tengas presupuesto, o ese electrodoméstico futuro. Lo útil es que podés crear metas de ahorro directamente desde productos o categorías de la lista.

**Sistema de Cuentas Múltiples** permite crear y gestionar diferentes contextos financieros completamente aislados. Podés tener una cuenta familiar para gastos compartidos del hogar y una cuenta personal para gastos individuales, cambiando entre ellas con un click. Los datos están completamente separados entre cuentas.

**Miembros Familiares** en cuentas tipo familia permiten atribuir cada gasto e ingreso a un miembro específico. Esto no requiere usuarios separados con credenciales propias, son simplemente etiquetas que permiten analizar después quién gastó qué, en qué categorías, y qué proporción de los gastos totales paga cada miembro de la familia.

## Stack Tecnológico

El proyecto está construido con un backend en Golang y un frontend en React, separados pero integrados para simplicidad de deployment.

### Backend

**Golang** es el lenguaje base del backend. La elección de Go fue principalmente motivada por el objetivo de aprender el lenguaje en el contexto de un proyecto real. Go ofrece sintaxis limpia, compilación a binario único que simplifica deployment, y un ecosistema maduro para APIs web.

**Gin** funciona como el framework web para manejar routing, middleware, y las respuestas HTTP de la API REST. Es el framework más popular de Go con excelente documentación y rendimiento sólido.

**PostgreSQL** es la base de datos elegida por su robustez, soporte completo para transacciones ACID críticas en aplicaciones financieras, y capacidades avanzadas de queries para análisis y reportes complejos.

**pgx** es el driver de PostgreSQL para Go, elegido por su rendimiento superior y soporte completo de características de PostgreSQL incluyendo arrays, JSON, y tipos custom.

**sqlc** genera código Go type-safe automáticamente desde queries SQL. Escribís SQL normal y sqlc produce funciones Go tipadas que eliminan errores en tiempo de compilación. Es el approach recomendado en la comunidad Go para interactuar con bases de datos relacionales.

**JWT (JSON Web Tokens)** maneja la autenticación mediante tokens en las peticiones HTTP. Los access tokens tienen duración corta y se complementan con refresh tokens almacenados en cookies httpOnly para mayor seguridad.

**embed.FS** es una feature nativa de Go que permite embeber archivos estáticos directamente en el binario compilado. El proyecto usa esto para incluir el build de producción de React dentro del ejecutable de Go, resultando en un solo archivo que contiene tanto backend como frontend.

### Frontend

**React** proporciona la biblioteca base para construir la interfaz de usuario mediante componentes reutilizables. El proyecto usa React con JavaScript inicialmente, con migración gradual a TypeScript planeada a medida que la complejidad crece.

**Vite** es el build tool y dev server que reemplaza Create React App. Ofrece hot module replacement instantáneo durante desarrollo y genera builds de producción optimizados. Es significativamente más rápido que alternativas como Webpack.

**JavaScript** es el lenguaje inicial del frontend por simplicidad y velocidad de desarrollo. El proyecto está configurado para migrar gradualmente a TypeScript archivo por archivo, empezando por componentes críticos que manejen lógica financiera compleja.

**Tailwind CSS** maneja todo el styling mediante utility classes directamente en JSX. Esto elimina decisiones de naming de clases CSS y permite desarrollo rápido manteniendo consistencia visual.

**shadcn/ui** proporciona componentes de UI prehechos y accesibles construidos sobre Radix UI. Incluye formularios, modals, dropdowns, y otros componentes complejos que aceleran significativamente el desarrollo sin sacrificar customización.

**TanStack Query** (anteriormente React Query) maneja el data fetching, caching, y sincronización de estado del servidor. Elimina boilerplate de manejo de loading states, errors, y actualizaciones automáticas.

**Recharts** genera las visualizaciones y gráficos del dashboard, permitiendo crear charts de distribución de gastos por categoría, tendencias temporales, y progreso de metas de ahorro con código declarativo simple.

**pnpm** es el package manager elegido por su velocidad superior a npm y yarn, y mejor gestión de espacio en disco mediante un store compartido de dependencias.

## Setup y Desarrollo

Para levantar el proyecto en tu máquina local, seguí estos pasos en orden.

### Prerrequisitos

Necesitás tener instalado Go versión 1.21 o superior, Node.js versión 18 o superior (para el frontend), y PostgreSQL versión 14 o superior corriendo en tu máquina.

Si no tenés Go instalado, podés descargarlo desde go.dev/dl. Durante la instalación, asegurate de que la variable de entorno GOPATH esté configurada correctamente.

Si no tenés PostgreSQL instalado, podés descargarlo desde postgresql.org o usar Docker para correr una instancia en un contenedor. Docker es particularmente útil si no querés instalar PostgreSQL directamente en tu sistema operativo.

### Instalación

El proyecto tiene dos partes principales que necesitás configurar: el backend en Go y el frontend en React. Primero cloná el repositorio en tu máquina local y entrá al directorio del proyecto.

```bash
git clone https://github.com/LorenzoCampos/bolsillo-claro.git
cd bolsillo-claro
```

Para el backend, las dependencias de Go se manejan mediante go.mod. No necesitás instalar nada manualmente porque Go descarga automáticamente las dependencias cuando compilás o corrés el proyecto por primera vez.

Para el frontend, entrá al directorio frontend e instalá las dependencias usando pnpm:

```bash
cd frontend
pnpm install
```

Si no tenés pnpm instalado, podés instalarlo globalmente con npm:

```bash
npm install -g pnpm
```

### Configuración de Variables de Entorno

Creá un archivo .env en la raíz del directorio backend con las siguientes variables. Este archivo contiene información sensible como credenciales de base de datos y claves secretas, por lo que nunca debe ser commiteado a git (ya está incluido en .gitignore).

```
DATABASE_URL="postgresql://usuario:password@localhost:5432/bolsillo_claro"
JWT_SECRET="genera-un-string-random-muy-seguro"
JWT_ACCESS_EXPIRY="15m"
JWT_REFRESH_EXPIRY="7d"
PORT="8080"
FRONTEND_URL="http://localhost:5173"
```

La variable DATABASE_URL debe apuntar a tu instancia de PostgreSQL. Reemplazá usuario y password con tus credenciales reales de PostgreSQL. El nombre de la base de datos puede ser bolsillo_claro o el que prefieras usar.

La variable JWT_SECRET es crítica para la seguridad. Esta clave se usa para firmar los tokens JWT de autenticación. Debe ser un string aleatorio y complejo. Nunca uses valores simples como "secret" o "123456". Podés generar una clave segura corriendo este comando en la terminal:

```bash
openssl rand -base64 32
```

Las variables JWT_ACCESS_EXPIRY y JWT_REFRESH_EXPIRY controlan cuánto tiempo duran los tokens. El access token debe tener vida corta (15 minutos es razonable) para limitar el impacto si es comprometido. El refresh token puede durar más (7 días) porque se almacena en una cookie httpOnly más segura.

La variable PORT define en qué puerto escucha el servidor de Go. El valor 8080 es estándar pero podés cambiarlo si ese puerto ya está ocupado en tu máquina.

La variable FRONTEND_URL indica la URL del frontend durante desarrollo, usada para configurar CORS correctamente. En desarrollo es http://localhost:5173 (el puerto default de Vite), pero en producción sería tu dominio real.

### Setup de Base de Datos

Una vez configuradas las variables de entorno, necesitás crear la base de datos y correr las migraciones SQL para establecer todas las tablas y relaciones. El proyecto usa archivos SQL planos para las migraciones en lugar de un ORM, lo cual te da control total sobre el schema.

Primero creá la base de datos manualmente en PostgreSQL si no existe. Podés hacer esto conectándote a PostgreSQL con psql o usando una herramienta gráfica como pgAdmin:

```bash
psql -U tu_usuario -c "CREATE DATABASE bolsillo_claro;"
```

Después corré las migraciones SQL que están en la carpeta migrations/ del backend. Estas migraciones crean todas las tablas necesarias con sus índices y constraints. El proyecto incluirá un script helper para ejecutar todas las migraciones en orden, pero también podés correrlas manualmente:

```bash
cd backend
psql -U tu_usuario -d bolsillo_claro -f migrations/001_initial_schema.sql
```

Finalmente, generá el código Go desde las queries SQL usando sqlc. Esta herramienta lee tus queries SQL en la carpeta sqlc/queries/ y genera funciones Go type-safe automáticamente:

```bash
cd backend
sqlc generate
```

Este comando crea archivos en db/sqlc/ con funciones Go que representan tus queries SQL. Cada vez que agregues o modifiques queries, necesitás correr sqlc generate de nuevo.

Si querés poblar la base de datos con datos de ejemplo para testing, el proyecto incluirá un seeder que podés correr desde Go:

```bash
cd backend
go run cmd/seed/main.go
```

### Correr el Proyecto

Durante desarrollo, necesitás correr el backend y frontend en terminales separadas. El backend de Go sirve la API REST en el puerto 8080, y el frontend de Vite corre su dev server en el puerto 5173 con hot reload automático.

En una terminal, iniciá el servidor backend de Go:

```bash
cd backend
go run cmd/server/main.go
```

El backend va a estar escuchando en http://localhost:8080. Go compila y ejecuta el código cada vez que salvás cambios, aunque no es tan instantáneo como el hot reload de Vite.

En otra terminal separada, iniciá el servidor de desarrollo de Vite para el frontend:

```bash
cd frontend
pnpm dev
```

El frontend va a estar disponible en http://localhost:5173. Vite proporciona hot module replacement extremadamente rápido, lo que significa que tus cambios en componentes React aparecen en el browser casi instantáneamente sin perder el estado de la aplicación.

Durante desarrollo, Vite está configurado para hacer proxy de las peticiones que empiezan con /api/ hacia el backend en el puerto 8080. Esto significa que cuando tu frontend hace fetch a /api/transactions, la petición se reenvía automáticamente a http://localhost:8080/api/transactions. Esta configuración elimina problemas de CORS durante desarrollo.

Una vez que ambos servidores están corriendo, abrí tu browser en http://localhost:5173 y deberías ver la aplicación funcionando con el frontend comunicándose con el backend.

### Comandos Útiles

Durante el desarrollo, estos son los comandos que vas a usar más frecuentemente organizados por backend y frontend.

**Backend (Go):**

Para correr el servidor en modo desarrollo: `go run cmd/server/main.go`

Para compilar el proyecto a un binario ejecutable: `go build -o bin/bolsillo-claro cmd/server/main.go`

Para ejecutar el binario compilado: `./bin/bolsillo-claro`

Para correr tests: `go test ./...`

Para formatear todo el código Go según el estilo estándar: `go fmt ./...`

Para regenerar el código de sqlc después de modificar queries: `sqlc generate`

Para agregar una nueva dependencia Go: `go get nombre-del-paquete`

**Frontend (React + Vite):**

Para correr el dev server con hot reload: `pnpm dev`

Para construir la aplicación para producción: `pnpm build`

Para previsualizar el build de producción localmente: `pnpm preview`

Para agregar una nueva dependencia: `pnpm add nombre-del-paquete`

Para agregar una dependencia de desarrollo: `pnpm add -D nombre-del-paquete`

Para actualizar todas las dependencias: `pnpm update`

**Base de Datos:**

Para conectarte directamente a la base de datos vía psql: `psql -U tu_usuario -d bolsillo_claro`

Para crear una nueva migración, creá manualmente un archivo SQL en migrations/ con el siguiente formato de nombre: `002_descripcion_de_cambio.sql`

Para hacer backup de la base de datos: `pg_dump -U tu_usuario bolsillo_claro > backup.sql`

Para restaurar desde un backup: `psql -U tu_usuario bolsillo_claro < backup.sql`

## Estructura del Proyecto

El proyecto está organizado en dos carpetas principales que separan claramente el backend en Go del frontend en React, facilitando el desarrollo independiente de cada parte.

### Backend (carpeta backend/)

La carpeta `cmd/` contiene los puntos de entrada de la aplicación. El archivo principal `cmd/server/main.go` inicializa el servidor HTTP, configura los routers de Gin, y conecta con la base de datos. Si el proyecto tiene otros comandos ejecutables como seeders o workers, también van acá como subcarpetas separadas.

La carpeta `api/` agrupa todo el código relacionado con la API REST. Dentro encontrás subcarpetas por recurso o dominio como `api/transactions/`, `api/accounts/`, `api/auth/`, cada una conteniendo los handlers (funciones que procesan las peticiones HTTP) y middleware relevante.

La carpeta `db/` maneja todo lo relacionado con la base de datos. `db/sqlc/` contiene el código auto-generado por sqlc que proporciona funciones type-safe para ejecutar queries. `db/migrations/` tiene los archivos SQL de migración numerados secuencialmente.

La carpeta `sqlc/` en la raíz del backend tiene dos subcarpetas importantes. `sqlc/queries/` contiene tus archivos .sql con las queries que escribís manualmente. `sqlc/schemas/` puede tener el schema de la base de datos si usás sqlc para generación de tipos. El archivo `sqlc.yaml` configura cómo sqlc genera el código.

La carpeta `internal/` contiene la lógica de negocio y código interno que no debería ser importado por otros proyectos. Acá van servicios, validadores, helpers, y cualquier código que no sea directamente HTTP handlers o acceso a datos.

La carpeta `pkg/` tiene código reutilizable que teóricamente podría ser usado por otros proyectos. Utilities generales, helpers de conversión de moneda, manejo de JWT, etc.

El archivo `go.mod` define el módulo de Go y lista todas las dependencias. `go.sum` contiene checksums de las dependencias para verificar integridad.

### Frontend (carpeta frontend/)

La carpeta `src/` es donde vive todo el código fuente de React. Se organiza en subcarpetas funcionales.

`src/components/` tiene todos los componentes React reutilizables. Los componentes base de shadcn/ui van en `components/ui/`, mientras que componentes específicos de tu aplicación van en subcarpetas como `components/transactions/`, `components/dashboard/`, etc.

`src/pages/` o `src/views/` (depende de tu preferencia) contiene los componentes que representan páginas completas de la aplicación. Por ejemplo `pages/Dashboard.jsx`, `pages/TransactionsList.jsx`. Estos componentes orquestan múltiples componentes más pequeños.

`src/api/` agrupa todas las funciones que hacen peticiones HTTP al backend. Cada archivo corresponde a un recurso, como `api/transactions.js` con funciones como `getTransactions()`, `createTransaction()`, etc.

`src/hooks/` tiene custom hooks de React que encapsulan lógica reutilizable. Por ejemplo `useAuth.js` para manejar autenticación, o `useTransactions.js` para obtener y mutar transacciones usando TanStack Query.

`src/lib/` contiene utilidades y helpers del frontend como funciones de formateo de fechas, conversión de moneda, validaciones, constantes, etc.

`src/styles/` puede tener archivos CSS globales si necesitás algo más allá de Tailwind, aunque con Tailwind esto generalmente es mínimo.

El archivo `vite.config.js` configura Vite, incluyendo el proxy hacia el backend durante desarrollo, plugins, y optimizaciones de build.

El archivo `package.json` define las dependencias de npm/pnpm y scripts disponibles.

### Raíz del Proyecto

El archivo `.gitignore` especifica qué archivos y carpetas no deberían ser trackeados por git, como `node_modules/`, archivos `.env`, y binarios compilados.

El archivo `README.md` (este archivo) proporciona documentación sobre cómo configurar y usar el proyecto.

El archivo `ARCHITECTURE.md` contiene la documentación completa sobre decisiones de diseño, especificaciones detalladas de cada módulo, y el plan de implementación.

## Conceptos Clave del Sistema

Para entender bien cómo funciona Bolsillo Claro, hay algunos conceptos fundamentales que vale la pena explicar.

### Cuentas Personales vs Familiares

El sistema maneja dos tipos de cuenta que funcionan de forma idéntica en cuanto a features disponibles, pero difieren en cómo atribuyen los movimientos financieros.

Una cuenta personal es administrada por un solo usuario y todos los gastos e ingresos pertenecen a esa persona. Es la opción más simple cuando gestionás solo tus finanzas individuales.

Una cuenta familiar permite definir múltiples miembros (Mamá, Papá, Hijo, etc.) y cada movimiento financiero se atribuye a un miembro específico. Esto permite analizar después quién gastó cuánto, en qué categorías, y qué proporción de gastos paga cada miembro. Importante: los miembros no son usuarios separados con login propio, son solo etiquetas dentro de esa cuenta familiar.

### Gastos Recurrentes vs Puntuales

Los gastos puntuales son compras one-time que hacés una vez y listo. Compraste algo, registrás el gasto, fin.

Los gastos recurrentes son compromisos que se repiten automáticamente cada mes. Cuando agregás un gasto recurrente de cinco mil pesos por Netflix, el sistema automáticamente crea una entrada de ese gasto cada mes sin que tengas que registrarlo manualmente. Esto es fundamental para trackear los gastos fijos mensuales que muchas veces se olvidan pero suman mucho.

### Metas de Ahorro

Las metas de ahorro pueden tener deadline o ser indefinidas. Una meta con deadline (por ejemplo, juntar trescientos mil pesos en seis meses) tiene cálculo automático de cuánto necesitás ahorrar mensualmente para alcanzarla. Una meta indefinida es simplemente un objetivo de largo plazo sin fecha específica.

Existe una meta especial llamada "Ahorro General" que se crea automáticamente en cada cuenta. Esta meta no tiene deadline y funciona como tu ahorro general que no está destinado a ningún objetivo específico.

### Múltiples Monedas

El sistema soporta nativamente pesos argentinos (ARS) y dólares estadounidenses (USD). Cuando registrás cualquier movimiento financiero, especificás en qué moneda es. Para visualizaciones consolidadas, el sistema convierte todo a la moneda base que configuraste usando un tipo de cambio que se actualiza.

## Documentación Adicional

Este README cubre lo esencial para empezar a trabajar en el proyecto. Para información más detallada sobre la arquitectura completa del sistema, decisiones de diseño, especificaciones de cada módulo, y el roadmap de implementación, revisá el documento `ARCHITECTURE.md` en la raíz del repositorio.

Ese documento contiene explicaciones profundas sobre cómo funciona cada módulo, el diseño de la base de datos con todas las relaciones entre tablas, el sistema de categorización opcional, las visualizaciones del dashboard, y estimaciones realistas de tiempo de desarrollo para cada fase.

## Notas de Desarrollo

A medida que trabajés en el proyecto, esta sección va a ir creciendo con observaciones y decisiones importantes que descubras durante el desarrollo. Es útil documentar acá cualquier gotcha, configuración no obvia, o solución a problemas específicos que encontraste.

(Esta sección se irá llenando durante el desarrollo del proyecto)
