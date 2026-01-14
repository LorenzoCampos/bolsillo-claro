package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/auth"
)

// AuthMiddleware es un middleware que valida JWT tokens
// Solo permite el acceso si el token es válido
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extraer el header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header requerido",
			})
			c.Abort() // Detener la ejecución de los siguientes handlers
			return
		}

		// El header debe tener formato: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Formato de Authorization inválido. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Extraer el token
		tokenString := parts[1]

		// Validar el token
		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido o expirado",
			})
			c.Abort()
			return
		}

		// Token válido - guardar el user_id en el contexto
		// Esto permite que los handlers accedan al user_id sin volver a parsear el token
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		// Continuar con el siguiente handler
		c.Next()
	}
}

// GetUserID es una función helper que extrae el user_id del contexto
// Debe ser llamada solo después del AuthMiddleware
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}
