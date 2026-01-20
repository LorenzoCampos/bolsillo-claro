# 📚 Documentación - Bolsillo Claro

Esta carpeta contiene toda la documentación técnica y de diseño del proyecto.

---

## 📑 Índice de Documentos

### 🔧 Diseño Técnico

- **[RECURRENCE-SYSTEM-DESIGN.md](./RECURRENCE-SYSTEM-DESIGN.md)**  
  Diseño completo del sistema de recurrencia para gastos.  
  **Status:** 📝 En Diseño → 🚧 Listo para implementar  
  **Última actualización:** 2026-01-16

---

## 🗺️ Documentos del Proyecto

### Documentación Principal

- **[API.md](../API.md)** - Documentación completa de la API REST
- **[API-CHEATSHEET.md](../API-CHEATSHEET.md)** - Referencia rápida de endpoints
- **[CHANGELOG.md](../CHANGELOG.md)** - Historial de cambios y decisiones
- **[README.md](../README.md)** - Información general del proyecto

### Frontend

- **[frontend/CHANGELOG.md](../frontend/CHANGELOG.md)** - Cambios del frontend
- **[frontend/README.md](../frontend/README.md)** - Setup y desarrollo frontend

### Backend

- **[backend/README.md](../backend/README.md)** - Setup y desarrollo backend
- **[backend/migrations/](../backend/migrations/)** - Migraciones de base de datos

---

## 🎯 Flujo de Trabajo para Nuevas Features

1. **Diseño Técnico**
   - Crear documento en `/docs/NOMBRE-FEATURE-DESIGN.md`
   - Definir: objetivo, casos de uso, schema DB, API, frontend
   - Revisar y aprobar diseño

2. **Actualizar Documentación**
   - Actualizar `API.md` con nuevos endpoints
   - Actualizar `CHANGELOG.md` con la feature
   - Documentar decisiones técnicas

3. **Implementación**
   - Backend: Migración → Handlers → Tests
   - Frontend: Types → Services → Components → Pages
   - Integración y testing

4. **Marcar como completado**
   - Actualizar status en `CHANGELOG.md`
   - Actualizar README si es necesario

---

## 📝 Convenciones de Documentación

### Formato de Documentos de Diseño

```markdown
# 🎯 Nombre de la Feature - Diseño Técnico

**Versión:** 1.0
**Fecha:** YYYY-MM-DD
**Autor:** [Nombre]
**Status:** 📝 En Diseño | 🚧 En Implementación | ✅ Completado

## Objetivo
[Descripción breve]

## Casos de Uso
[Ejemplos reales]

## Diseño de Base de Datos
[Schema, migraciones, constraints]

## API Changes
[Nuevos endpoints o modificaciones]

## Frontend Changes
[Componentes, types, services]

## Consideraciones Técnicas
[Performance, edge cases, validaciones]
```

### Emojis Usados

- 📝 Documento / En diseño
- 🚧 En construcción / implementación
- ✅ Completado
- 🎯 Objetivo / Meta
- 🔧 Técnico / Implementación
- 💡 Idea / Caso de uso
- 🗄️ Base de datos
- 📡 API
- 🎨 Frontend / UI
- ⚠️ Advertencia / Importante
- 🐛 Bug
- ⚡ Performance
- 🔒 Seguridad

---

**Última actualización:** 2026-01-16
