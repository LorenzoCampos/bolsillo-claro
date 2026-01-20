# ⚠️ ARCHIVOS HISTÓRICOS - NO USAR

Esta carpeta contiene documentación obsoleta guardada para referencia histórica únicamente.

## 🚫 **IMPORTANTE: NO USAR ESTOS ARCHIVOS COMO REFERENCIA**

Estos documentos fueron archivados el **16 de enero de 2026** durante una consolidación masiva de documentación que eliminó duplicaciones, conflictos y contenido obsoleto.

---

## 📚 Documentación Actual y Vigente

Si necesitás información sobre el proyecto, usá estos archivos (en la raíz del repo):

### Documentación Principal
- **`README.md`** - Overview del proyecto, setup rápido, getting started
- **`API.md`** - Especificación completa de todos los endpoints de la API
- **`STACK.md`** - Stack tecnológico y decisiones arquitectónicas
- **`FEATURES.md`** - Guía narrativa de funcionalidades (qué hace cada módulo)
- **`CHANGELOG.md`** - Historial de cambios y versiones

### Documentación Técnica (carpeta `docs/`)
- **`docs/DATABASE.md`** - Schema completo de base de datos, migraciones, constraints
- **`docs/MULTI-CURRENCY.md`** - Sistema de multi-moneda (Modo 3)
- **`docs/RECURRENCE.md`** - Sistema de recurrencia de gastos/ingresos

---

## 📦 Archivos en esta carpeta (solo para referencia histórica)

| Archivo Original | ¿Por qué se archivó? |
|------------------|----------------------|
| `ARCHITECTURE.md` (2871 líneas) | Monstruoso, duplicaba contenido de todos los demás docs. Migrado a DATABASE.md y API.md |
| `API.md` (1473 líneas) | Contenía duplicaciones y mezcla de diseño + implementación. Consolidado en nuevo API.md limpio |
| `API-CHEATSHEET.md` (271 líneas) | Resumen redundante de API.md. Fusionado en el nuevo API.md bien organizado |
| `DECISIONS.md` (608 líneas) | Decisiones técnicas aisladas del contexto. Migradas inline a STACK.md y docs específicas |
| `STACK.md` (492 líneas) | Contenía info obsoleta ("Frontend en desarrollo 🚧" cuando ya estaba implementado). Reescrito |
| `CHANGELOG.md` (199 líneas) | Mezclaba decisiones técnicas con cambios de código. Limpiado para solo versiones |
| `README.md` (313 líneas) | Demasiado largo para ser un README. Simplificado a ~150 líneas |
| `docs-README.md` (111 líneas) | Índice innecesario. README principal ahora apunta directamente a docs/ |
| `RECURRENCE-SYSTEM-DESIGN.md` (441 líneas) | Bien hecho pero duplicado en API.md. Renombrado a docs/RECURRENCE.md y actualizado |
| `frontend-README.md` (73 líneas) | Boilerplate de Vite sin info del proyecto. Reemplazado |

---

## 🎯 Resultados de la Consolidación

**Antes:**
- 10 archivos markdown
- ~6,852 líneas totales
- Duplicaciones masivas
- Información contradictoria
- Docs obsoletas mezcladas con actuales

**Después:**
- 7 archivos markdown (4 raíz + 3 en docs/)
- ~2,150 líneas totales
- **68% de reducción** manteniendo toda la info importante
- Sin duplicaciones
- Info clara, organizada, actualizada

---

## 📅 Información de Archivado

- **Fecha:** 2026-01-16
- **Razón:** Consolidación masiva para eliminar quilombo documental
- **Proceso:** Plan A - Consolidación agresiva con migración de contenido
- **Backup:** Todos los archivos originales preservados en esta carpeta

---

## ❓ Si necesitás algo de estos archivos viejos

1. **Revisá primero la nueva documentación** - Probablemente ya esté ahí, mejor explicado
2. Si NO está en la nueva doc y es importante → Avisá para agregarlo
3. NO copies contenido viejo sin verificar que sigue siendo válido

---

**Mantenido por:** Gentleman Programming & Lorenzo  
**Última actualización de este README:** 2026-01-16
