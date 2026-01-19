package savings_goals

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// CreateSavingsGoalRequest represents the request to create a savings goal
type CreateSavingsGoalRequest struct {
	Name         string  `json:"name" binding:"required,min=1,max=255"`
	Description  *string `json:"description,omitempty"`
	TargetAmount float64 `json:"target_amount" binding:"required,gt=0"`
	SavedIn      *string `json:"saved_in,omitempty" binding:"omitempty,max=255"`
	Deadline     *string `json:"deadline,omitempty"` // Format: YYYY-MM-DD
}

// SavingsGoalResponse represents a savings goal
type SavingsGoalResponse struct {
	ID                     string   `json:"id"`
	AccountID              string   `json:"account_id"`
	Name                   string   `json:"name"`
	Description            *string  `json:"description,omitempty"`
	TargetAmount           float64  `json:"target_amount"`
	CurrentAmount          float64  `json:"current_amount"`
	Currency               string   `json:"currency"`
	SavedIn                *string  `json:"saved_in,omitempty"`
	Deadline               *string  `json:"deadline,omitempty"`
	ProgressPercentage     float64  `json:"progress_percentage"`
	RequiredMonthlySavings *float64 `json:"required_monthly_savings,omitempty"`
	IsActive               bool     `json:"is_active"`
	CreatedAt              string   `json:"created_at"`
	UpdatedAt              string   `json:"updated_at"`
}

// calculateRequiredMonthlySavings calcula cuánto hay que ahorrar por mes para alcanzar la meta
// Retorna nil si no hay deadline o si ya pasó la fecha
func calculateRequiredMonthlySavings(currentAmount, targetAmount float64, deadline *time.Time) *float64 {
	if deadline == nil {
		return nil // Sin deadline no se puede calcular
	}

	now := time.Now()
	if deadline.Before(now) {
		return nil // Deadline ya pasó
	}

	// Calcular meses restantes
	years := deadline.Year() - now.Year()
	months := int(deadline.Month() - now.Month())
	totalMonths := years*12 + months

	// Si es el mismo mes o ya pasó, retornar nil
	if totalMonths <= 0 {
		return nil
	}

	// Calcular cuánto falta
	amountRemaining := targetAmount - currentAmount
	if amountRemaining <= 0 {
		// Ya alcanzó la meta
		zero := 0.0
		return &zero
	}

	// Calcular cuánto hay que ahorrar por mes
	requiredMonthly := amountRemaining / float64(totalMonths)
	return &requiredMonthly
}

// CreateSavingsGoal handles POST /api/savings-goals
func CreateSavingsGoal(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context (set by AccountMiddleware)
		accountID, exists := middleware.GetAccountID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Parse request
		var req CreateSavingsGoalRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Validate deadline (if provided, must be future date)
		var deadlineDate *time.Time
		if req.Deadline != nil {
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

		// Get account currency (savings goal inherits currency from account)
		var currency string
		err := db.QueryRow(ctx, `SELECT currency FROM accounts WHERE id = $1`, accountID).Scan(&currency)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account currency"})
			return
		}

		// Check if a goal with the same name already exists for this account
		var goalExists bool
		err = db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM savings_goals WHERE account_id = $1 AND name = $2 AND is_active = true)`,
			accountID, req.Name,
		).Scan(&goalExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing goal"})
			return
		}
		if goalExists {
			c.JSON(http.StatusConflict, gin.H{"error": "ya existe una meta de ahorro con ese nombre"})
			return
		}

		// Insert savings goal
		var goalID uuid.UUID
		var createdAt, updatedAt time.Time
		insertQuery := `
			INSERT INTO savings_goals (
				account_id, name, description, target_amount, 
				current_amount, currency, saved_in, deadline, is_active
			) VALUES ($1, $2, $3, $4, 0, $5, $6, $7, true)
			RETURNING id, created_at, updated_at
		`

		err = db.QueryRow(ctx, insertQuery,
			accountID, req.Name, req.Description, req.TargetAmount,
			currency, req.SavedIn, deadlineDate,
		).Scan(&goalID, &createdAt, &updatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create savings goal: " + err.Error()})
			return
		}

		// Obtener user_id del contexto para logging
		userID, _ := middleware.GetUserID(c)

		// Log de creación exitosa
		logger.Info("savings_goal.created", "Meta de ahorro creada", map[string]interface{}{
			"goal_id":       goalID.String(),
			"account_id":    accountID,
			"user_id":       userID,
			"name":          req.Name,
			"target_amount": req.TargetAmount,
			"deadline":      req.Deadline,
			"ip":            c.ClientIP(),
		})

		// Calcular required_monthly_savings si hay deadline
		requiredMonthlySavings := calculateRequiredMonthlySavings(0, req.TargetAmount, deadlineDate)

		// Build response
		response := SavingsGoalResponse{
			ID:                     goalID.String(),
			AccountID:              accountID,
			Name:                   req.Name,
			Description:            req.Description,
			TargetAmount:           req.TargetAmount,
			CurrentAmount:          0,
			Currency:               currency,
			SavedIn:                req.SavedIn,
			Deadline:               req.Deadline,
			ProgressPercentage:     0,
			RequiredMonthlySavings: requiredMonthlySavings,
			IsActive:               true,
			CreatedAt:              createdAt.Format(time.RFC3339),
			UpdatedAt:              updatedAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Meta de ahorro creada exitosamente",
			"savings_goal": response,
		})
	}
}
