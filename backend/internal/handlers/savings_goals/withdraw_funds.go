package savings_goals

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// WithdrawFundsRequest represents the request to withdraw funds from a savings goal
type WithdrawFundsRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description *string `json:"description,omitempty"`
	Date        string  `json:"date" binding:"required"` // Format: YYYY-MM-DD
}

// WithdrawFunds handles POST /api/savings-goals/:id/withdraw-funds
func WithdrawFunds(db *pgxpool.Pool) gin.HandlerFunc {
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
		var req WithdrawFundsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate date format
		transactionDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}

	// Check if date is not in the future
	if transactionDate.After(time.Now().Truncate(24 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "la fecha no puede ser futura"})
		return
	}

	ctx := c.Request.Context()

	// First, we need to check the goal's deadline before starting the transaction
	// to validate the transaction date
	var goalDeadline *time.Time
	preCheckQuery := `SELECT deadline FROM savings_goals WHERE id = $1 AND account_id = $2`
	err = db.QueryRow(ctx, preCheckQuery, goalID, accountID).Scan(&goalDeadline)
	
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "meta de ahorro no encontrada o no pertenece a esta cuenta"})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check savings goal"})
		return
	}

	// Validate that transaction date is not after goal deadline
	if goalDeadline != nil {
		deadlineDate := time.Date(goalDeadline.Year(), goalDeadline.Month(), goalDeadline.Day(), 0, 0, 0, 0, time.UTC)
		transactionDateUTC := time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, time.UTC)
		
		if transactionDateUTC.After(deadlineDate) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "no puedes retirar fondos con una fecha posterior al deadline de la meta",
				"transaction_date": req.Date,
				"goal_deadline": goalDeadline.Format("2006-01-02"),
			})
			return
		}
	}

		// Start transaction
		tx, err := db.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
			return
		}
		defer tx.Rollback(ctx)

	// Check if goal exists and belongs to this account
	var currentAmount, targetAmount float64
	var name string
	var deadline *time.Time
	checkQuery := `SELECT name, current_amount, target_amount, deadline FROM savings_goals WHERE id = $1 AND account_id = $2`
	err = tx.QueryRow(ctx, checkQuery, goalID, accountID).Scan(&name, &currentAmount, &targetAmount, &deadline)

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "meta de ahorro no encontrada o no pertenece a esta cuenta"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check savings goal"})
			return
		}

		// Validate that there are enough funds to withdraw
		if req.Amount > currentAmount {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "No hay suficientes fondos para retirar",
				"current_amount": currentAmount,
				"requested":      req.Amount,
				"available":      currentAmount,
			})
			return
		}

		// Create transaction record
		var transactionID uuid.UUID
		var createdAt time.Time
		insertTxnQuery := `
			INSERT INTO savings_goal_transactions (
				savings_goal_id, amount, transaction_type, description, date
			) VALUES ($1, $2, 'withdrawal', $3, $4)
			RETURNING id, created_at
		`

		err = tx.QueryRow(ctx, insertTxnQuery,
			goalID, req.Amount, req.Description, req.Date,
		).Scan(&transactionID, &createdAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
			return
		}

		// Update savings goal current_amount
		newAmount := currentAmount - req.Amount
		updateQuery := `
			UPDATE savings_goals 
			SET current_amount = $1, updated_at = NOW()
			WHERE id = $2
			RETURNING current_amount, updated_at
		`

		var updatedAmount float64
		var updatedAt time.Time
		err = tx.QueryRow(ctx, updateQuery, newAmount, goalID).Scan(&updatedAmount, &updatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update savings goal"})
			return
		}

		// Commit transaction
		err = tx.Commit(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
			return
		}

		// Calculate progress percentage
		progressPercentage := 0.0
		if targetAmount > 0 {
			progressPercentage = (updatedAmount / targetAmount) * 100
		}

		// Obtener user_id del contexto para logging
		userID, _ := middleware.GetUserID(c)

		// Log de retiro de fondos
		logger.Info("savings_goal.withdraw_funds", "Fondos retirados de meta de ahorro", map[string]interface{}{
			"goal_id":       goalID,
			"account_id":    accountID,
			"user_id":       userID,
			"amount":        req.Amount,
			"new_balance":   updatedAmount,
			"goal_name":     name,
			"ip":            c.ClientIP(),
		})

		// Build response
		c.JSON(http.StatusOK, gin.H{
			"message": "Fondos retirados exitosamente",
			"savings_goal": gin.H{
				"id":                  goalID,
				"name":                name,
				"current_amount":      updatedAmount,
				"target_amount":       targetAmount,
				"progress_percentage": progressPercentage,
				"updated_at":          updatedAt.Format(time.RFC3339),
			},
			"transaction": gin.H{
				"id":               transactionID.String(),
				"amount":           -req.Amount, // Negative for display
				"transaction_type": "withdrawal",
				"description":      req.Description,
				"date":             req.Date,
				"created_at":       createdAt.Format(time.RFC3339),
			},
		})
	}
}
