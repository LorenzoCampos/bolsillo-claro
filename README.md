# Bolsillo Claro

Gestor financiero personal y familiar construido para trackear gastos, ingresos y metas de ahorro. Diseñado específicamente para la realidad argentina con soporte nativo para múltiples monedas y el "dólar tarjeta".

## ⚡ Quick Start

```bash
# Backend (Go)
cd backend
cp .env.example .env  # Configurar variables
go run cmd/server/main.go

# Frontend (React + Vite)
cd frontend
pnpm install
pnpm dev
```

El backend corre en `http://localhost:8080` y el frontend en `http://localhost:5173`.

---

## 🎯 ¿Qué hace este proyecto?

Bolsillo Claro responde tres preguntas fundamentales:
1. **¿Cuánto gasto?** - Trackea gastos puntuales y recurrentes (Netflix, alquiler, etc.)
2. **¿Cuánto ingresa?** - Maneja ingresos fijos, variables y temporales
3. **¿Cuánto estoy ahorrando?** - Metas con cálculo automático de ahorro mensual necesario

### Características Principales

- 💰 **Multi-moneda real**: Sistema "Modo 3" que captura el dólar tarjeta argentino (impuestos incluidos)
- 🔄 **Gastos recurrentes**: Suscripciones, servicios y cuotas se trackean automáticamente
- 👨‍👩‍👧‍👦 **Cuentas familiares**: Atribuí gastos e ingresos a miembros específicos
- 🎯 **Metas de ahorro**: Define objetivos y el sistema calcula cuánto ahorrar mensualmente
- 📊 **Dashboard inteligente**: Resumen consolidado con análisis por categoría
- 🏷️ **Categorías flexibles**: Predefinidas + custom por cuenta

---

## 📚 Documentación

### Para Usuarios/Product
- **[FEATURES.md](./FEATURES.md)** - Guía narrativa completa: qué hace cada módulo y cómo funciona
- **[CHANGELOG.md](./CHANGELOG.md)** - Historial de versiones y cambios

### Para Desarrolladores
- **[API.md](./API.md)** - Especificación completa de endpoints, request/response, validaciones
- **[STACK.md](./STACK.md)** - Stack tecnológico y decisiones arquitectónicas
- **[docs/DATABASE.md](./docs/DATABASE.md)** - Schema de base de datos, migraciones, constraints
- **[docs/MULTI-CURRENCY.md](./docs/MULTI-CURRENCY.md)** - Sistema de multi-moneda (Modo 3)
- **[docs/RECURRENCE.md](./docs/RECURRENCE.md)** - Sistema de recurrencia avanzado

---

## 🛠️ Stack Tecnológico

### Backend
- **Go 1.23** - Lenguaje principal
- **Gin** - Framework HTTP
- **PostgreSQL 15** - Base de datos
- **pgx/v5** - Driver de PostgreSQL
- **JWT** - Autenticación (access + refresh tokens)

### Frontend
- **React 18** - UI library
- **Vite 6** - Build tool
- **TypeScript** - Tipado estático
- **TailwindCSS v4** - Styling
- **TanStack Query v5** - Data fetching y cache
- **React Hook Form + Zod** - Formularios y validación
- **Axios** - Cliente HTTP

### Deployment
- **Docker** - Containerización
- **Apache 2.4** - Reverse proxy
- **VPS Debian 12** - Servidor
- **URL Producción**: https://api.fakerbostero.online/bolsillo

---

## 🚀 Setup Completo

### Prerrequisitos

- Go 1.23+
- Node.js 18+
- PostgreSQL 15+
- pnpm (package manager)

### 1. Clonar el repositorio

```bash
git clone https://github.com/LorenzoCampos/bolsillo-claro.git
cd bolsillo-claro
```

### 2. Setup Backend

```bash
cd backend

# Instalar dependencias (Go las descarga automáticamente)
go mod download

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus credenciales de PostgreSQL y JWT_SECRET
```

**Variables de entorno requeridas (`.env`):**
```bash
DATABASE_URL="postgresql://usuario:password@localhost:5432/bolsillo_claro"
JWT_SECRET="genera-un-string-random-seguro"  # openssl rand -base64 32
JWT_ACCESS_EXPIRY="15m"
JWT_REFRESH_EXPIRY="7d"
PORT="8080"
FRONTEND_URL="http://localhost:5173"
```

**Crear base de datos y ejecutar migraciones:**
```bash
# Crear database
psql -U postgres -c "CREATE DATABASE bolsillo_claro;"

# Ejecutar migraciones en orden
psql -U postgres -d bolsillo_claro -f migrations/001_create_users_table.sql
psql -U postgres -d bolsillo_claro -f migrations/002_create_accounts_table.sql
# ... (ejecutar todas las migraciones en orden numérico)
```

**Correr el servidor:**
```bash
go run cmd/server/main.go
# Backend escuchando en http://localhost:8080
```

### 3. Setup Frontend

```bash
cd frontend

# Instalar dependencias
pnpm install

# Correr dev server
pnpm dev
# Frontend disponible en http://localhost:5173
```

Durante desarrollo, Vite hace proxy de `/api/*` al backend en puerto 8080 automáticamente (ver `vite.config.ts`).

---

## 📁 Estructura del Proyecto

```
bolsillo-claro/
├── backend/
│   ├── cmd/
│   │   └── server/main.go          # Entry point
│   ├── internal/
│   │   ├── handlers/               # HTTP handlers por módulo
│   │   ├── middleware/             # Auth, Account, CORS
│   │   ├── config/                 # Configuración
│   │   └── database/               # Conexión DB
│   ├── pkg/
│   │   └── auth/                   # JWT, bcrypt
│   ├── migrations/                 # Migraciones SQL
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── pages/                  # Páginas principales
│   │   ├── components/             # Componentes reutilizables
│   │   ├── services/               # API calls (Axios)
│   │   ├── hooks/                  # Custom hooks
│   │   └── types/                  # TypeScript types + Zod schemas
│   ├── package.json
│   └── vite.config.ts
│
├── docs/                           # Documentación técnica
├── FEATURES.md                     # Guía de funcionalidades
├── API.md                          # Especificación API
├── STACK.md                        # Stack y decisiones
└── README.md                       # Este archivo
```

---

## 🔐 Autenticación

El sistema usa JWT con dos tipos de tokens:
- **Access token**: Válido 15 minutos, va en header `Authorization: Bearer <token>`
- **Refresh token**: Válido 7 días, almacenado en cookie httpOnly

Al registrarte quedás logueado automáticamente (devuelve tokens).

**Headers requeridos:**
```bash
# Endpoints de cuentas
Authorization: Bearer <access_token>

# Endpoints de gastos/ingresos/dashboard
Authorization: Bearer <access_token>
X-Account-ID: <account_uuid>
```

---

## 🎨 Conceptos Clave

### Cuentas
La unidad fundamental. Podés tener múltiples cuentas completamente aisladas:
- **Personal**: Finanzas individuales
- **Familiar**: Con miembros (Mamá, Papá, Hijo) para atribuir gastos/ingresos

### Gastos Recurrentes
Compromisos que se repiten automáticamente: Netflix, gimnasio, alquiler. El sistema los detecta en cada mes sin crear registros duplicados.

### Multi-Currency Modo 3
Registrás un gasto en USD indicando cuántos pesos te debitaron realmente. El sistema calcula la tasa efectiva (captura dólar tarjeta con impuestos).

**Ejemplo:**
```json
{
  "amount": 20,
  "currency": "USD",
  "amount_in_primary_currency": 31500
}
// → Sistema calcula: exchange_rate = 1575 (dólar tarjeta)
```

### Metas de Ahorro
Con deadline → calcula ahorro mensual necesario
Sin deadline → objetivo de largo plazo

Agregar fondos a una meta descuenta del balance disponible (descuento virtual, no crea expense).

---

## 🧪 Testing

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
pnpm test
```

---

## 📝 Comandos Útiles

### Backend
```bash
go run cmd/server/main.go           # Dev server
go build -o bin/server cmd/server/main.go  # Build
go fmt ./...                        # Format
go test ./...                       # Tests
```

### Frontend
```bash
pnpm dev                            # Dev server
pnpm build                          # Build producción
pnpm preview                        # Preview build
pnpm lint                           # ESLint
```

### Database
```bash
# Conectar a DB
psql -U postgres -d bolsillo_claro

# Backup
pg_dump -U postgres bolsillo_claro > backup.sql

# Restore
psql -U postgres bolsillo_claro < backup.sql
```

---

## 🤝 Contribuir

Este proyecto es de uso personal pero abierto a contribuciones. Si encontrás bugs o querés proponer features:

1. Abrí un issue describiendo el problema/feature
2. Fork del repo
3. Creá tu branch: `git checkout -b feature/nueva-feature`
4. Commit: `git commit -m 'Add: nueva feature'`
5. Push: `git push origin feature/nueva-feature`
6. Abrí un Pull Request

---

## 📄 Licencia

Este proyecto es de código abierto bajo licencia MIT.

---

## 🙋 Contacto

- **Autor:** Gentleman Programming & Lorenzo
- **Email:** [tu-email]
- **GitHub:** [@LorenzoCampos](https://github.com/LorenzoCampos)

---

**Creado con ❤️ para gestionar finanzas sin vueltas**
