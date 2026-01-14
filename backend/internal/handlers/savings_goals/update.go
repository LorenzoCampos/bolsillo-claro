package savings_goals

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// UpdateSavingsGoalRequest represents the request to update a savings goal
type UpdateSavingsGoalRequest struct {
	Name         *string  `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description  *string  `json:"description,omitempty"`
	TargetAmount *float64 `json:"target_amount,omitempty" binding:"omitempty,gt=0"`
	SavedIn      *string  `json:"saved_in,omitempty" binding:"omitempty,max=255"`
	Deadline     *string  `json:"deadline,omitempty"` // Format: YYYY-MM-DD or empty string to clear
	IsActive     *bool    `json:"is_active,omitempty"`
}

// UpdateSavingsGoal handles PUT /api/savings-goals/:id
func UpdateSavingsGoal(db *pgxpool.Pool) gin.HandlerFunc {
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

		// Parse request
		var req UpdateSavingsGoalRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Check if goal exists and belongs to this account
		var existingGoal struct {
			Name string
		}
		checkQuery := `SELECT name FROM savings_goals WHERE id = $1 AND account_id = $2`
		err := db.QueryRow(ctx, checkQuery, goalID, accountID).Scan(&existingGoal.Name)

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "meta de ahorro no encontrada o no pertenece a esta cuenta"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check savings goal"})
			return
		}

		// If name is being updated, check for duplicates
		if req.Name != nil && *req.Name != existingGoal.Name {
			var exists bool
			err = db.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM savings_goals WHERE account_id = $1 AND name = $2 AND id != $3 AND is_active = true)`,
				accountID, *req.Name, goalID,
			).Scan(&exists)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check duplicate name"})
				return
			}
			if exists {
				c.JSON(http.StatusConflict, gin.H{"error": "ya existe una meta de ahorro con ese nombre"})
				return
			}
		}

		// Validate deadline (if provided, must be future date)
		var deadlineDate *time.Time
		var clearDeadline bool
		if req.Deadline != nil {
			if *req.Deadline == "" {
				// Empty string means clear the deadline
				clearDeadline = true
			} else {
				parsedDate, err := time.Parse("2006-01-02", *req.Deadline)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline format, use YYYY-MM-DD"})
					return
				}

				// Check if deadline is in the future
				if parsedDate.Before(time.Now().Truncate(24 * time.Hour)) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "deadline must be a future date"})
					return
				}

				deadlineDate = &parsedDate
			}
		}

		// Build dynamic UPDATE query
		updateQuery := `
			UPDATE savings_goals SET
				name = COALESCE($1, name),
				description = COALESCE($2, description),
				target_amount = COALESCE($3, target_amount),
				saved_in = COALESCE($4, saved_in),
				deadline = CASE 
					WHEN $5::boolean THEN NULL
					WHEN $6::date IS NOT NULL THEN $6::date
					ELSE deadline
				END,
				is_active = COALESCE($7, is_active),
				updated_at = NOW()
			WHERE id = $8 AND account_id = $9
			RETURNING id, account_id, name, description, target_amount, 
			          current_amount, currency, saved_in, deadline, 
			          is_active, created_at, updated_at
		`

		var goal SavingsGoalResponse
		var description, savedIn *string
		var deadline *time.Time
		var createdAt, updatedAt time.Time

		err = db.QueryRow(ctx, updateQuery,
			req.Name, req.Description, req.TargetAmount, req.SavedIn,
			clearDeadline, deadlineDate, req.IsActive,
			goalID, accountID,
		).Scan(
			&goal.ID, &goal.AccountID, &goal.Name, &description,
			&goal.TargetAmount, &goal.CurrentAmount, &goal.Currency,
			&savedIn, &deadline, &goal.IsActive, &createdAt, &updatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update savings goal: " + err.Error()})
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

		c.JSON(http.StatusOK, gin.H{
			"message":      "Meta de ahorro actualizada exitosamente",
			"savings_goal": goal,
		})
	}
}
