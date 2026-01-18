# ============================================================================
# STAGE 1: Build
# ============================================================================
FROM golang:1.23-alpine AS builder

# Instalar dependencias de compilación
RUN apk add --no-cache git

# Directorio de trabajo
WORKDIR /app

# Copiar go.mod y go.sum primero (cache de dependencias)
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Descargar dependencias adicionales (si go.mod cambió)
RUN go mod download && go mod tidy

# Compilar binario optimizado
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/server ./cmd/server/main.go

# ============================================================================
# STAGE 2: Production
# ============================================================================
FROM alpine:latest

# Instalar certificados SSL y timezone data
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root para seguridad
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

# Copiar binario desde builder
COPY --from=builder /app/bin/server .
COPY --from=builder /app/migrations ./migrations

# Cambiar permisos
RUN chown -R appuser:appuser /home/appuser

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Comando por defecto
CMD ["./server"]
