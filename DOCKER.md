# 🐳 Guía Docker - Bolsillo Claro

Esta guía te enseña cómo correr el proyecto completo (backend + base de datos) usando Docker.

---

## 📋 Prerrequisitos

Necesitás tener instalado:
- **Docker** (versión 20.10 o superior)
- **Docker Compose** (versión 1.29 o superior)

### Verificar instalación:
```bash
docker --version
docker-compose --version
```

### Instalar Docker:
- **Linux**: `curl -fsSL https://get.docker.com | sh`
- **Mac**: Descargar Docker Desktop desde [docker.com](https://www.docker.com/products/docker-desktop)
- **Windows**: Descargar Docker Desktop desde [docker.com](https://www.docker.com/products/docker-desktop)

---

## 🚀 Inicio Rápido

### 1. Clonar el repositorio
```bash
git clone https://github.com/LorenzoCampos/bolsillo-claro.git
cd bolsillo-claro
```

### 2. Levantar los servicios
```bash
docker-compose up
```

**¡Y LISTO!** 🎉

- Backend corriendo en: `http://localhost:8080`
- Postgres corriendo en: `localhost:5432`
- Migraciones ejecutadas automáticamente

### 3. Detener los servicios
Presioná `Ctrl + C` en la terminal donde corriste `docker-compose up`

O desde otra terminal:
```bash
docker-compose down
```

---

## 📚 Comandos Esenciales

### Levantar servicios (modo detached - en background)
```bash
docker-compose up -d
```
- Los servicios corren en segundo plano
- Podés cerrar la terminal y siguen corriendo
- Para ver logs: `docker-compose logs -f`

### Ver logs en tiempo real
```bash
docker-compose logs -f
```
- `-f`: Follow (sigue mostrando logs nuevos)
- Para ver logs de un servicio específico: `docker-compose logs -f backend`

### Ver estado de los servicios
```bash
docker-compose ps
```
Muestra qué servicios están corriendo y su estado de salud (health)

### Detener servicios
```bash
docker-compose down
```
- Para y elimina los contenedores
- **NO elimina los datos** de Postgres (están en el volumen persistente)

### Detener Y eliminar volúmenes (¡CUIDADO! Esto BORRA todos los datos)
```bash
docker-compose down -v
```
- Usa esto solo si querés empezar desde cero con la base de datos vacía

### Reconstruir las imágenes (cuando cambiás código)
```bash
docker-compose up --build
```
- Fuerza a Docker a reconstruir la imagen del backend
- Útil cuando modificás el Dockerfile o el código Go

### Entrar a un contenedor (para debugging)
```bash
# Entrar al backend
docker exec -it bolsillo-claro-backend sh

# Entrar a Postgres
docker exec -it bolsillo-claro-db psql -U bolsillo_user -d bolsillo_claro
```

---

## 🗄️ Base de Datos

### Conectarse a Postgres desde tu máquina
Podés usar cualquier cliente de Postgres (pgAdmin, DBeaver, TablePlus, etc.):

```
Host: localhost
Port: 5432
User: bolsillo_user
Password: bolsillo_password_dev
Database: bolsillo_claro
```

### Ejecutar migraciones manualmente (si es necesario)
Las migraciones se ejecutan automáticamente la primera vez que se crea la base de datos.

Si necesitás ejecutarlas manualmente:
```bash
docker exec -it bolsillo-claro-db psql -U bolsillo_user -d bolsillo_claro -f /docker-entrypoint-initdb.d/001_create_users_table.sql
```

### Hacer backup de la base de datos
```bash
docker exec bolsillo-claro-db pg_dump -U bolsillo_user bolsillo_claro > backup.sql
```

### Restaurar backup
```bash
docker exec -i bolsillo-claro-db psql -U bolsillo_user -d bolsillo_claro < backup.sql
```

### Resetear la base de datos (empezar desde cero)
```bash
# Detener servicios y eliminar volúmenes
docker-compose down -v

# Levantar de nuevo (recreará todo)
docker-compose up
```

---

## 🔧 Configuración

### Variables de entorno
Las variables están definidas en `docker-compose.yml`.

Para producción, crear un archivo `.env` y usar:
```bash
docker-compose --env-file .env up
```

### Cambiar puertos
Si el puerto 8080 o 5432 ya están ocupados en tu máquina, podés cambiarlos en `docker-compose.yml`:

```yaml
ports:
  - "9090:8080"  # Ahora el backend estará en localhost:9090
```

---

## 🐛 Troubleshooting

### El backend no puede conectarse a Postgres
1. Verificá que Postgres esté "healthy": `docker-compose ps`
2. Mirá los logs: `docker-compose logs postgres`
3. Asegurate de que la `DATABASE_URL` use `postgres` como host (no `localhost`)

### Las migraciones no se ejecutan
Las migraciones se ejecutan **solo la primera vez** que se crea la base de datos.

Si ya tenés una base de datos creada:
1. Eliminá el volumen: `docker-compose down -v`
2. Volvé a levantar: `docker-compose up`

### Puerto ya en uso
Si ves un error como "port is already allocated":
1. Verificá qué está usando el puerto: `lsof -i :8080` (Mac/Linux) o `netstat -ano | findstr :8080` (Windows)
2. Detené el proceso que usa ese puerto
3. O cambiá el puerto en `docker-compose.yml`

### Cambios en el código no se reflejan
Docker construye la imagen una sola vez. Para que tome los cambios:
```bash
docker-compose up --build
```

---

## 🎓 Conceptos Clave

### ¿Qué es una imagen?
Es un "snapshot" de tu aplicación con todas sus dependencias. Es inmutable.

### ¿Qué es un contenedor?
Es una instancia corriendo de una imagen. Podés tener múltiples contenedores de la misma imagen.

### ¿Qué es un volumen?
Es almacenamiento persistente. Sin volúmenes, cuando un contenedor se destruye, se pierden los datos.

### ¿Qué hace `depends_on`?
Asegura que el backend NO arranque hasta que Postgres esté "healthy" (listo para recibir conexiones).

### ¿Por qué usar `postgres` como host en vez de `localhost`?
Dentro de Docker, cada contenedor tiene su propio `localhost`. Docker crea un DNS interno donde los servicios se encuentran por nombre.

---

## 🚀 Producción

Para producción, NO uses este `docker-compose.yml` directamente. Considerá:

1. **Secrets**: Usar Docker Secrets o variables de entorno seguras
2. **Volúmenes**: Backups automáticos de la base de datos
3. **Reverse Proxy**: Nginx o Traefik delante del backend
4. **SSL**: Certificados HTTPS
5. **Monitoring**: Logs centralizados, métricas

Ver `backend/Dockerfile` que ya tiene optimizaciones de producción (multi-stage build, usuario no-root, etc.)

---

## 📖 Recursos

- [Documentación oficial de Docker](https://docs.docker.com/)
- [Documentación de Docker Compose](https://docs.docker.com/compose/)
- [Best practices de Dockerfile](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)

---

**¿Problemas? Abrí un issue en GitHub.**
