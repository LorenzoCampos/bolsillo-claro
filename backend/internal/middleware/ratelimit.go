package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RateLimitEntry representa un registro de intentos de una IP
type RateLimitEntry struct {
	Count      int       // Número de intentos
	FirstTry   time.Time // Timestamp del primer intento
	ResetAfter time.Time // Cuándo se resetea el contador
}

// RateLimiter maneja el rate limiting en memoria
// NOTA: En producción con múltiples instancias, considerar Redis para almacenamiento compartido
type RateLimiter struct {
	mu      sync.RWMutex               // Protege el map de accesos concurrentes
	entries map[string]*RateLimitEntry // IP -> Entry
	limit   int                        // Máximo de intentos permitidos
	window  time.Duration              // Ventana de tiempo (ej: 15 minutos)
}

// NewRateLimiter crea una nueva instancia del rate limiter
// limit: número máximo de requests permitidos
// window: ventana de tiempo (ej: 15 * time.Minute)
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*RateLimitEntry),
		limit:   limit,
		window:  window,
	}

	// Goroutine para limpiar entradas expiradas cada 5 minutos
	// Esto evita que el map crezca indefinidamente
	go rl.cleanupExpired()

	return rl
}

// cleanupExpired limpia entradas expiradas del map cada 5 minutos
func (rl *RateLimiter) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			if now.After(entry.ResetAfter) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow verifica si una IP puede hacer una request
// Retorna true si está permitido, false si excedió el límite
func (rl *RateLimiter) Allow(ip string) (allowed bool, resetAfter time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Si no existe entrada para esta IP, crear una nueva
	entry, exists := rl.entries[ip]
	if !exists {
		rl.entries[ip] = &RateLimitEntry{
			Count:      1,
			FirstTry:   now,
			ResetAfter: now.Add(rl.window),
		}
		return true, now.Add(rl.window)
	}

	// Si la ventana expiró, resetear el contador
	if now.After(entry.ResetAfter) {
		entry.Count = 1
		entry.FirstTry = now
		entry.ResetAfter = now.Add(rl.window)
		return true, entry.ResetAfter
	}

	// Incrementar contador
	entry.Count++

	// Verificar si excedió el límite
	if entry.Count > rl.limit {
		return false, entry.ResetAfter
	}

	return true, entry.ResetAfter
}

// Middleware retorna un middleware de Gin que aplica rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener la IP real del cliente
		// Gin's ClientIP() maneja correctamente X-Forwarded-For, X-Real-IP, etc.
		ip := c.ClientIP()

		// Verificar si está permitido
		allowed, resetAfter := rl.Allow(ip)

		if !allowed {
			// Log de rate limit excedido
			logger.LogRateLimitExceeded(ip, c.Request.URL.Path)

			// Calcular cuántos segundos faltan para reset
			retryAfter := int(time.Until(resetAfter).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}

			// Agregar headers informativos
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			c.Header("Retry-After", strconv.Itoa(retryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Demasiados intentos. Por favor, intentá de nuevo más tarde.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		// Request permitida, continuar
		c.Next()
	}
}

// RateLimitAuth crea un rate limiter específico para endpoints de autenticación
// Límite: 5 intentos cada 15 minutos (agresivo para prevenir brute-force)
func RateLimitAuth() gin.HandlerFunc {
	limiter := NewRateLimiter(5, 15*time.Minute)
	return limiter.Middleware()
}

// RateLimitGeneral crea un rate limiter más permisivo para endpoints generales
// Límite: 100 requests cada minuto
func RateLimitGeneral() gin.HandlerFunc {
	limiter := NewRateLimiter(100, 1*time.Minute)
	return limiter.Middleware()
}
