# 🚀 Quick Start - Bolsillo Claro Frontend

## ✅ Estado Actual

**Frontend:** ✅ Corriendo en http://localhost:5173  
**Backend:** ✅ Corriendo en http://localhost:9090/api  
**Red Local:** ✅ Accesible en http://192.168.0.46:5173

---

## 🌐 URLs de Acceso

### Desde esta máquina (localhost):
```
Frontend: http://localhost:5173
Backend:  http://localhost:9090/api
```

### Desde otros dispositivos en tu red local:
```
Frontend: http://192.168.0.46:5173
Backend:  http://192.168.0.46:9090/api
```

**Ejemplo:** Abrí el navegador de tu celular/tablet y entrá a `http://192.168.0.46:5173`

---

## 🔧 Comandos Útiles

### Desarrollo
```bash
# Levantar servidor de desarrollo
pnpm dev

# Type check (verificar errores de TypeScript)
pnpm exec tsc --noEmit

# Linting
pnpm lint

# Build para producción
pnpm build

# Preview del build
pnpm preview
```

### Debugging
```bash
# Ver logs del servidor Vite
pnpm dev

# Type check en modo watch
pnpm exec tsc --noEmit --watch

# Verificar que el backend responde
curl http://localhost:9090/api
```

---

## 🛠️ Herramientas de Desarrollo

### React Query DevTools
Cuando el frontend esté corriendo, vas a ver un ícono flotante en la esquina inferior izquierda (solo en desarrollo).

Click en el ícono para:
- Ver todas las queries en cache
- Ver estado de cada query (loading, success, error)
- Ver cuándo se refetchea data
- Invalidar queries manualmente
- Ver mutations

### Vite HMR (Hot Module Replacement)
Los cambios que hagas en el código se reflejan INSTANTÁNEAMENTE en el browser sin necesidad de recargar la página.

---

## 📂 Archivos de Configuración Importantes

### `.env.development` (actual)
```env
VITE_API_URL=http://localhost:9090/api
VITE_ENV=development
```

### `.env.production`
```env
VITE_API_URL=https://api.fakerbostero.online/bolsillo/api
VITE_ENV=production
```

### `vite.config.ts`
```typescript
server: {
  host: '0.0.0.0',  // Escucha en todas las interfaces
  port: 5173,
  strictPort: true,
}
```

---

## 🔐 Variables de Entorno

El frontend usa estas variables de entorno:

- `VITE_API_URL` - URL del backend API
- `VITE_ENV` - Ambiente (development/production)

**Importante:** Las variables DEBEN empezar con `VITE_` para ser expuestas al frontend.

---

## 🎯 Próximos Pasos de Desarrollo

### 1. Verificar Conexión con Backend
Abrí el navegador en http://localhost:5173 y abrí la consola (F12).

En la consola del navegador, ejecutá:
```javascript
// Test de conexión básico
fetch('http://localhost:9090/api')
  .then(r => r.text())
  .then(console.log)
```

### 2. Implementar Login/Register
Crear componentes en `src/features/auth/`:
- `Login.tsx`
- `Register.tsx`

### 3. Crear Custom Hooks
En `src/hooks/`:
- `useAuth.ts` - Login, logout, register
- `useExpenses.ts` - CRUD de expenses
- `useIncomes.ts` - CRUD de incomes

### 4. Setup Router
En `App.tsx`, configurar React Router con:
- Public routes (/, /login, /register)
- Protected routes (/dashboard, /expenses, etc.)

---

## 🐛 Troubleshooting

### Frontend no carga
```bash
# Verificar que Vite está corriendo
ps aux | grep vite

# Si no está corriendo, levantar
pnpm dev
```

### Backend no responde
```bash
# Verificar que el backend está corriendo en puerto 9090
curl http://localhost:9090/api

# Verificar procesos
ps aux | grep "go run\|backend"
```

### Cambios no se reflejan
```bash
# Hard refresh en el browser
Ctrl + Shift + R (Windows/Linux)
Cmd + Shift + R (Mac)

# O reiniciar Vite
# Ctrl+C para detener
pnpm dev
```

### CORS errors
Si ves errores de CORS en la consola del browser, verificá que el backend esté configurado para aceptar requests desde localhost:5173.

---

## 📚 Documentación

- [README.md](./README.md) - Documentación completa del proyecto
- [FIXES_APPLIED.md](./FIXES_APPLIED.md) - Fixes aplicados al setup
- [SETUP_SUMMARY.md](./SETUP_SUMMARY.md) - Resumen completo del setup
- [../API.md](../API.md) - Documentación completa de la API backend

---

## ✅ Checklist de Setup

- [x] pnpm instalado
- [x] Dependencias instaladas (`pnpm install`)
- [x] TypeScript compila sin errores
- [x] Variables de entorno configuradas
- [x] Servidor Vite corriendo en puerto 5173
- [x] Accesible desde red local (0.0.0.0)
- [x] Backend corriendo en localhost:9090
- [x] React Query DevTools instalado
- [x] Axios configurado con interceptors
- [x] Tipos TypeScript de toda la API creados

---

**Todo listo para empezar a desarrollar! 🎉**

**Sugerencia:** Empezá por crear un componente de Login simple para probar la conexión con el backend.
