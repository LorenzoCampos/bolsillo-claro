package savings_goals

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// ListSavingsGoals handles GET /api/savings-goals
func ListSavingsGoals(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context
		accountID, exists := middleware.GetAccountID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		ctx := c.Request.Context()

		// Query all active savings goals for this account
		query := `
			SELECT 
				id, account_id, name, description, target_amount, 
				current_amount, currency, saved_in, deadline, 
				is_active, created_at, updated_at
			FROM savings_goals
			WHERE account_id = $1 AND is_active = true
			ORDER BY created_at DESC
		`

		rows, err := db.Query(ctx, query, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch savings goals"})
			return
		}
		defer rows.Close()

		savingsGoals := []SavingsGoalResponse{}
		for rows.Next() {
			var goal SavingsGoalResponse
			var description, savedIn *string
			var deadline *time.Time
			var createdAt, updatedAt time.Time

			err := rows.Scan(
				&goal.ID, &goal.AccountID, &goal.Name, &description,
				&goal.TargetAmount, &goal.CurrentAmount, &goal.Currency,
				&savedIn, &deadline, &goal.IsActive, &createdAt, &updatedAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse savings goal"})
				return
			}

			// Set optional fields
			goal.Description = description
			goal.SavedIn = savedIn

			if deadline != nil {
				deadlineStr := deadline.Format("2006-01-02")
				goal.Deadline = &deadlineStr
			}

			// Calculate progress percentage
			if goal.TargetAmount > 0 {
				goal.ProgressPercentage = (goal.CurrentAmount / goal.TargetAmount) * 100
			} else {
				goal.ProgressPercentage = 0
			}

			goal.CreatedAt = createdAt.Format(time.RFC3339)
			goal.UpdatedAt = updatedAt.Format(time.RFC3339)

			savingsGoals = append(savingsGoals, goal)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading savings goals"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"savings_goals": savingsGoals,
			"total_count":   len(savingsGoals),
		})
	}
}
