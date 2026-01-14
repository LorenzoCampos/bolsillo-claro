# 📝 Changelog - Bolsillo Claro Frontend

Todos los cambios notables en el frontend serán documentados aquí.

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/).

---

## [Sin Release] - 2026-01-14

### ✅ Agregado (Setup Inicial)

#### Core Setup
- ✅ Proyecto Vite 6.x con template React + TypeScript
- ✅ Node.js v22.20.0 / npm v10.9.3
- ✅ Configuración de Vite para acceso remoto (`host: 0.0.0.0`)
- ✅ Estructura de carpetas (components, pages, hooks, services, types, context, utils)

#### Dependencias Instaladas

**Framework & Core** (175 paquetes base)
```json
{
  "react": "^18.3.x",
  "react-dom": "^18.3.x",
  "typescript": "~5.6.x",
  "vite": "^6.x.x"
}
```

**Stack Completo** (+36 paquetes)
```json
{
  "react-router-dom": "^6.x.x",           // Routing
  "@tanstack/react-query": "^5.x.x",      // Data fetching & caching
  "@tanstack/react-query-devtools": "^5.x.x", // DevTools para TanStack Query
  "axios": "^1.x.x",                      // HTTP client
  "react-hook-form": "^7.x.x",            // Form management
  "zod": "^3.x.x",                        // Validation
  "@hookform/resolvers": "^3.x.x",        // Integración React Hook Form + Zod
  "tailwindcss": "^4.0.0-beta.x"          // Styling (v4 beta - zero config)
}
```

**Total:** 212 paquetes, 0 vulnerabilidades ✅

#### Configuraciones
- ✅ Tailwind CSS v4 configurado (zero-config, solo `@import "tailwindcss"`)
- ✅ Vite configurado para desarrollo remoto
- ✅ TypeScript strict mode habilitado

---

### 🚧 Pendiente

#### Próximos Pasos Inmediatos
1. 🚧 Configurar Axios instance con base URL
2. 🚧 Configurar TanStack Query provider
3. 🚧 Crear AuthContext para manejo de sesión
4. 🚧 Crear página de Login
5. 🚧 Implementar protected routes
6. 🚧 Crear página de Dashboard

#### Componentes a Crear
- [ ] UI Components (Button, Input, Card, Modal)
- [ ] Layout Components (Header, Sidebar, Footer)
- [ ] Form Components (LoginForm, RegisterForm)
- [ ] Page Components (Login, Dashboard, Expenses, etc.)

#### Servicios a Implementar
- [ ] `services/api.ts` - Axios instance con interceptors
- [ ] `services/auth.ts` - Login, Register, Refresh, Logout
- [ ] `services/expenses.ts` - CRUD de gastos
- [ ] `services/incomes.ts` - CRUD de ingresos
- [ ] `services/savingsGoals.ts` - CRUD de metas

#### Types y Schemas
- [ ] `types/user.ts` - User, LoginRequest, RegisterRequest
- [ ] `types/expense.ts` - Expense, CreateExpenseRequest
- [ ] `types/income.ts` - Income, CreateIncomeRequest
- [ ] `types/savingsGoal.ts` - SavingsGoal, Transaction

#### Hooks Personalizados
- [ ] `useAuth()` - Hook para autenticación
- [ ] `useExpenses()` - Hook para gestión de gastos
- [ ] `useIncomes()` - Hook para gestión de ingresos
- [ ] `useSavingsGoals()` - Hook para metas de ahorro

---

### 📊 Estadísticas

- **Tamaño node_modules:** ~250MB
- **Dependencias totales:** 212 paquetes
- **Vulnerabilidades:** 0
- **Puerto desarrollo:** 5173
- **URL desarrollo:** http://200.58.105.147:5173

---

### 🔧 Comandos Disponibles

```bash
npm run dev      # Inicia dev server (puerto 5173)
npm run build    # Build para producción
npm run preview  # Preview del build
npm run lint     # Lint con ESLint
```

---

### 🎯 Decisiones Técnicas

#### ¿Por qué Tailwind v4 beta?
- Zero-config (no necesita `tailwind.config.js`)
- Más rápido (nueva engine en Rust)
- Ya estable para producción
- Menos boilerplate

#### ¿Por qué desarrollo en VPS?
- Accesible desde cualquier dispositivo
- Siempre activo para ver progreso
- Simula ambiente de producción real

#### ¿Por qué TypeScript desde el principio?
- Previene bugs antes de runtime
- Autocompletado increíble
- Refactoring seguro
- Es lo que se usa en empresas serias

---

**Última actualización:** 2026-01-14 21:52 UTC-3  
**Autor:** Gentleman AI + Lorenzo  
**Versión:** 0.1.0 (Setup)
