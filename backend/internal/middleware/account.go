package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LorenzoCampos/bolsillo-claro/internal/database"
)

// AccountMiddleware valida que el header X-Account-ID existe y pertenece al usuario autenticado
// Este middleware debe aplicarse DESPUÉS del AuthMiddleware
func AccountMiddleware(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extraer user_id del contexto (viene del AuthMiddleware)
		userID, userExists := c.Get("user_id")
		if !userExists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Usuario no autenticado",
			})
			c.Abort()
			return
		}

		// Extraer X-Account-ID del header
		accountID := c.GetHeader("X-Account-ID")
		if accountID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Header X-Account-ID es requerido",
			})
			c.Abort()
			return
		}

		// Validar que la cuenta existe y pertenece al usuario autenticado
		ctx := c.Request.Context()
		var accountExists bool
		query := `
		SELECT EXISTS(
			SELECT 1 FROM accounts 
			WHERE id = $1 AND user_id = $2
		)
	`
		err := db.Pool.QueryRow(ctx, query, accountID, userID).Scan(&accountExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error verificando cuenta",
			})
			c.Abort()
			return
		}

		if !accountExists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "La cuenta no existe o no te pertenece",
			})
			c.Abort()
			return
		}

		// Todo OK - guardar account_id en el contexto
		c.Set("account_id", accountID)

		// Continuar con el siguiente handler
		c.Next()
	}
}

// GetAccountID es una función helper que extrae el account_id del contexto
// Debe ser llamada solo después del AccountMiddleware
func GetAccountID(c *gin.Context) (string, bool) {
	accountID, exists := c.Get("account_id")
	if !exists {
		return "", false
	}

	accountIDStr, ok := accountID.(string)
	return accountIDStr, ok
}
