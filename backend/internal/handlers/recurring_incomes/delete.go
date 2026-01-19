package recurring_incomes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// DeleteRecurringIncome maneja DELETE /api/recurring-expenses/:id
// SOFT DELETE: Solo marca is_active = false (no borra datos)
// Los gastos ya generados NO se eliminan (histórico preservado)
func DeleteRecurringIncome(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recurringID := c.Param("id")

		// Obtener account_id del contexto
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Account-ID header requerido",
			})
			return
		}

		// Obtener user_id del contexto
		userID, _ := middleware.GetUserID(c)

		ctx := c.Request.Context()

		// Verificar que existe y pertenece a esta cuenta
		var existsCheck bool
		checkQuery := "SELECT EXISTS(SELECT 1 FROM recurring_incomes WHERE id = $1 AND account_id = $2)"
		err := pool.QueryRow(ctx, checkQuery, recurringID, accountID).Scan(&existsCheck)
		if err != nil || !existsCheck {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Ingreso recurrente no encontrado",
			})
			return
		}

		// Contar cuántos gastos se generaron (para informar al usuario)
		var generatedCount int
		countQuery := "SELECT COUNT(*) FROM expenses WHERE recurring_income_id = $1"
		err = pool.QueryRow(ctx, countQuery, recurringID).Scan(&generatedCount)
		if err != nil {
			generatedCount = 0
		}

		// SOFT DELETE: marcar como inactivo
		// Esto detiene la generación de nuevos gastos sin borrar el histórico
		deleteQuery := "UPDATE recurring_incomes SET is_active = false WHERE id = $1 AND account_id = $2"
		_, err = pool.Exec(ctx, deleteQuery, recurringID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error eliminando ingreso recurrente",
				"details": err.Error(),
			})
			return
		}

		// Log de eliminación
		logger.Info("recurring_expense.deleted", "Ingreso recurrente eliminado (soft delete)", map[string]interface{}{
			"recurring_income_id": recurringID,
			"account_id":           accountID,
			"user_id":              userID,
			"generated_expenses":   generatedCount,
			"ip":                   c.ClientIP(),
		})

		c.JSON(http.StatusOK, gin.H{
			"message":            "Ingreso recurrente eliminado exitosamente",
			"generated_expenses": generatedCount,
			"note":               "Los gastos ya generados NO se eliminan. Solo se detiene la generación futura.",
		})
	}
}
