package incomes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func DeleteIncome(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context (set by AccountMiddleware)
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Get income ID from URL parameter
		incomeID := c.Param("id")
		if incomeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "income_id is required"})
			return
		}

		// Delete the income (only if it belongs to this account)
		deleteQuery := "DELETE FROM incomes WHERE id = $1 AND account_id = $2"
		commandTag, err := db.Exec(c.Request.Context(), deleteQuery, incomeID, accountID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete income: " + err.Error()})
			return
		}

		// Check if any row was actually deleted
		if commandTag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "income not found or does not belong to this account"})
			return
		}

		// Return success with no content
		c.JSON(http.StatusOK, gin.H{
			"message": "income deleted successfully",
			"id":      incomeID,
		})
	}
}
