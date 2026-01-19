package savings_goals

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// DeleteSavingsGoal handles DELETE /api/savings-goals/:id
// Only allows deletion if current_amount = 0
func DeleteSavingsGoal(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context
		accountID, exists := middleware.GetAccountID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Get savings goal ID from URL
		goalID := c.Param("id")
		if goalID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "savings_goal_id is required"})
			return
		}

		ctx := c.Request.Context()

		// Check if goal exists and get current_amount
		var currentAmount float64
		var name string
		checkQuery := `SELECT name, current_amount FROM savings_goals WHERE id = $1 AND account_id = $2`
		err := db.QueryRow(ctx, checkQuery, goalID, accountID).Scan(&name, &currentAmount)

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "meta de ahorro no encontrada o no pertenece a esta cuenta"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check savings goal"})
			return
		}

		// Check if goal has funds
		if currentAmount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":          "No se puede eliminar una meta de ahorro con fondos asignados",
				"current_amount": currentAmount,
				"suggestion":     "Retire todos los fondos primero o archive la meta (is_active = false)",
			})
			return
		}

		// Delete the goal (CASCADE will delete transactions too)
		deleteQuery := `DELETE FROM savings_goals WHERE id = $1 AND account_id = $2`
		cmdTag, err := db.Exec(ctx, deleteQuery, goalID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete savings goal"})
			return
		}

		if cmdTag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "meta de ahorro no encontrada"})
			return
		}

		// Obtener user_id del contexto para logging
		userID, _ := middleware.GetUserID(c)

		// Log de eliminación exitosa
		logger.Info("savings_goal.deleted", "Meta de ahorro eliminada", map[string]interface{}{
			"goal_id":    goalID,
			"account_id": accountID,
			"user_id":    userID,
			"name":       name,
			"ip":         c.ClientIP(),
		})

		c.JSON(http.StatusOK, gin.H{
			"message":         "Meta de ahorro eliminada exitosamente",
			"savings_goal_id": goalID,
			"name":            name,
		})
	}
}
