# 🌐 Acceso desde Red Local - Bolsillo Claro

Esta guía te explica cómo acceder a tu backend de Bolsillo Claro desde **cualquier dispositivo en tu red local** (celular, tablet, otra computadora, etc.).

---

## 📍 TU IP LOCAL

Tu máquina tiene la siguiente IP en la red local:

```
192.168.0.46
```

**IMPORTANTE**: Esta IP puede cambiar si:
- Reiniciás el router
- Tu máquina obtiene una IP diferente del DHCP
- Cambiás de red WiFi

### Cómo verificar tu IP actual:

```bash
# Linux/Mac
hostname -I | awk '{print $1}'

# O también
ip addr show | grep "inet " | grep -v 127.0.0.1

# Windows (PowerShell)
ipconfig | findstr IPv4
```

---

## 🚀 URLs de Acceso

### Backend API:
```
http://192.168.0.46:9090
```

### Health Check:
```
http://192.168.0.46:9090/api/health
```

### Postgres (desde otro cliente):
```
Host: 192.168.0.46
Port: 5433
User: bolsillo_user
Password: bolsillo_password_dev
Database: bolsillo_claro
```

---

## 📱 CÓMO ACCEDER DESDE DIFERENTES DISPOSITIVOS

### 1. Desde tu Celular

#### A. Usando el navegador:
1. Conectá tu celular a la **misma red WiFi** que tu computadora
2. Abrí el navegador (Chrome, Safari, etc.)
3. Accedé a: `http://192.168.0.46:9090/api/health`
4. Deberías ver: `{"message":"Bolsillo Claro API está funcionando correctamente","status":"ok"}`

#### B. Usando una app (React Native, Flutter, etc.):
```javascript
// Configurá la URL del backend
const API_URL = "http://192.168.0.46:9090/api";

// Ejemplo de fetch
fetch(`${API_URL}/health`)
  .then(res => res.json())
  .then(data => console.log(data));
```

---

### 2. Desde otra Computadora

#### A. Usando curl:
```bash
curl http://192.168.0.46:9090/api/health
```

#### B. Usando Postman/Insomnia:
1. Crear nueva request
2. URL: `http://192.168.0.46:9090/api/register`
3. Método: POST
4. Body (JSON):
```json
{
  "name": "Test User",
  "email": "test@example.com",
  "password": "password123"
}
```

#### C. Frontend (React/Vite):
```javascript
// src/config.js o similar
export const API_URL = import.meta.env.PROD 
  ? "https://api.produccion.com"
  : "http://192.168.0.46:9090/api";

// Usar en servicios
import { API_URL } from './config';

axios.get(`${API_URL}/accounts`)
```

---

### 3. Desde una Tablet

Mismo procedimiento que el celular. Asegurate de estar en la misma red WiFi.

---

## 🔒 CORS (Cross-Origin Resource Sharing)

### ¿Qué es CORS?

Es un mecanismo de seguridad del navegador que **bloquea** peticiones desde un origen diferente al del servidor.

**Ejemplo de problema**:
- Tu frontend corre en `http://192.168.0.50:3000` (otra máquina)
- Intenta hacer fetch a `http://192.168.0.46:9090/api/users`
- El navegador lo bloquea por CORS ❌

### Orígenes permitidos actualmente:

```
http://localhost:5173          # Vite local
http://localhost:3000           # React local
http://192.168.0.46:5173       # Vite desde red local
http://192.168.0.46:3000       # React desde red local
http://192.168.0.46:9090       # API desde red local
```

### Agregar más orígenes:

Editá `docker-compose.yml`:

```yaml
environment:
  ALLOWED_ORIGINS: http://localhost:5173,http://192.168.0.46:5173,http://192.168.0.50:3000
```

Y reiniciá el backend:
```bash
docker-compose restart backend
```

### Para desarrollo rápido (permitir TODO):

**⚠️ SOLO EN DESARROLLO, NUNCA EN PRODUCCIÓN**

```yaml
environment:
  ALLOWED_ORIGINS: "*"
```

---

## 🔥 FIREWALL (si no funciona)

Si no podés acceder desde otros dispositivos, puede ser el **firewall** de tu sistema operativo.

### Linux (UFW):
```bash
# Ver estado del firewall
sudo ufw status

# Permitir puerto 9090
sudo ufw allow 9090/tcp

# Permitir puerto 5433 (Postgres)
sudo ufw allow 5433/tcp
```

### Linux (firewalld):
```bash
# Ver estado
sudo firewall-cmd --state

# Permitir puertos
sudo firewall-cmd --zone=public --add-port=9090/tcp --permanent
sudo firewall-cmd --zone=public --add-port=5433/tcp --permanent
sudo firewall-cmd --reload
```

### Mac:
Por defecto, macOS no bloquea puertos salientes. Si tenés problemas, revisá "Configuración → Seguridad y Privacidad → Firewall".

### Windows:
```powershell
# Permitir puerto 9090
netsh advfirewall firewall add rule name="Bolsillo Backend" dir=in action=allow protocol=TCP localport=9090

# Permitir puerto 5433
netsh advfirewall firewall add rule name="Bolsillo Postgres" dir=in action=allow protocol=TCP localport=5433
```

---

## 🧪 TESTING RÁPIDO

### Test 1: Desde la misma máquina (localhost)
```bash
curl http://localhost:9090/api/health
# ✅ Debería funcionar
```

### Test 2: Desde la misma máquina (IP local)
```bash
curl http://192.168.0.46:9090/api/health
# ✅ Debería funcionar
```

### Test 3: Desde otro dispositivo
```bash
# Desde tu celular/tablet conectado a la misma WiFi
# Abrí el navegador → http://192.168.0.46:9090/api/health
# ✅ Debería mostrar: {"message":"Bolsillo Claro API está funcionando correctamente","status":"ok"}
```

### Test 4: CORS desde navegador
```javascript
// Abrí la consola del navegador en tu celular/tablet (F12)
fetch('http://192.168.0.46:9090/api/health')
  .then(r => r.json())
  .then(console.log)
// ✅ Si funciona, CORS está bien configurado
// ❌ Si sale "blocked by CORS", agregá el origen a ALLOWED_ORIGINS
```

---

## 🌟 IP ESTÁTICA (OPCIONAL)

Para que tu IP no cambie cada vez que reiniciás el router:

### En el Router (recomendado):
1. Entrá al panel de administración del router (generalmente `192.168.0.1` o `192.168.1.1`)
2. Buscá "DHCP Reservation" o "Reserva de IP"
3. Asigná la IP `192.168.0.46` a la MAC address de tu máquina

### En tu Sistema Operativo:
```bash
# Linux - Editá /etc/network/interfaces o usar NetworkManager

# Mac - Configuración → Red → Avanzado → TCP/IP → Configurar IPv4: Manual

# Windows - Configuración → Red e Internet → Cambiar opciones del adaptador
```

---

## 🚀 EJEMPLO COMPLETO: Frontend + Backend en Red Local

### Escenario:
- **Backend**: En tu máquina principal (192.168.0.46:9090)
- **Frontend Vite**: En otra máquina (192.168.0.50:5173)

### Configuración del Frontend:

```javascript
// vite.config.ts
export default defineConfig({
  server: {
    host: '0.0.0.0', // Exponé Vite en todas las interfaces
    port: 5173,
  },
  // NO necesitás proxy porque estás accediendo directamente al backend
})

// src/config.ts
export const API_URL = "http://192.168.0.46:9090/api";
```

### Actualizar ALLOWED_ORIGINS en docker-compose.yml:

```yaml
ALLOWED_ORIGINS: http://localhost:5173,http://192.168.0.46:5173,http://192.168.0.50:5173
```

### Reiniciar backend:
```bash
docker-compose restart backend
```

---

## 📝 RESUMEN

| Dispositivo | URL | Notas |
|------------|-----|-------|
| **Tu máquina** | `http://localhost:9090` | Acceso local |
| **Cualquier dispositivo en red** | `http://192.168.0.46:9090` | Mismo WiFi |
| **Postgres externo** | `192.168.0.46:5433` | pgAdmin, DBeaver, etc. |

---

## ❓ TROUBLESHOOTING

### No puedo acceder desde otro dispositivo:
1. ✅ Verificá que ambos dispositivos estén en la misma red WiFi
2. ✅ Verificá tu IP actual: `hostname -I`
3. ✅ Verificá que Docker esté corriendo: `docker-compose ps`
4. ✅ Verificá el firewall (ver sección arriba)
5. ✅ Intentá hacer ping: `ping 192.168.0.46`

### Error de CORS:
1. ✅ Agregá el origen del frontend a `ALLOWED_ORIGINS` en `docker-compose.yml`
2. ✅ Reiniciá el backend: `docker-compose restart backend`
3. ✅ Verificá los logs: `docker-compose logs backend | grep CORS`

### La IP cambió:
1. ✅ Verificá la nueva IP: `hostname -I`
2. ✅ Actualizá `ALLOWED_ORIGINS` en `docker-compose.yml`
3. ✅ Reiniciá: `docker-compose restart backend`
4. ✅ Considerá configurar IP estática (ver arriba)

---

**¿Problemas? Abrí un issue en GitHub con los logs de `docker-compose logs`**


{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZDEyYmE5MjMtNzU1MS00NzNiLTllZjktNjg5OWVhMWY3YjIzIiwiZW1haWwiOiJlc3RhLnR1cm5lcjFAeWFob28uY29tIiwiaXNzIjoiYm9sc2lsbG8tY2xhcm8iLCJleHAiOjE3Njg5NDY0MTksImlhdCI6MTc2ODk0NTUxOX0.qNxVUVgm2EK5nrhFyJTNyOJcI34VZ6leSaFvIaeOCB0",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZDEyYmE5MjMtNzU1MS00NzNiLTllZjktNjg5OWVhMWY3YjIzIiwiZW1haWwiOiIiLCJpc3MiOiJib2xzaWxsby1jbGFybyIsImV4cCI6MTc2OTU1MDMxOSwiaWF0IjoxNzY4OTQ1NTE5fQ.0CVBwLk7wrL2W4-zVrSMWrOoh01UpEpPg_nGjzMuKv4",
    "user": {
        "id": "d12ba923-7551-473b-9ef9-6899ea1f7b23",
        "email": "esta.turner1@yahoo.com",
        "name": "Marjorie Witting"
    }
}