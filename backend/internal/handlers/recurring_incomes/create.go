package recurring_incomes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// CreateRecurringIncomeRequest representa el JSON para crear un ingreso recurrente
type CreateRecurringIncomeRequest struct {
	Description       string   `json:"description" binding:"required"`
	Amount            float64  `json:"amount" binding:"required,gt=0"`
	Currency          string   `json:"currency" binding:"required,oneof=ARS USD EUR"`
	CategoryID        *string  `json:"category_id"`
	FamilyMemberID    *string  `json:"family_member_id"`
	
	// Recurrence configuration
	RecurrenceFrequency   string `json:"recurrence_frequency" binding:"required,oneof=daily weekly monthly yearly"`
	RecurrenceInterval    int    `json:"recurrence_interval" binding:"omitempty,gt=0"`
	RecurrenceDayOfMonth  *int   `json:"recurrence_day_of_month" binding:"omitempty,gte=1,lte=31"`
	RecurrenceDayOfWeek   *int   `json:"recurrence_day_of_week" binding:"omitempty,gte=0,lte=6"`
	
	// Time boundaries
	StartDate         string  `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate           *string `json:"end_date"`                      // YYYY-MM-DD (nullable)
	TotalOccurrences  *int    `json:"total_occurrences" binding:"omitempty,gt=0"`
	
	// Multi-currency (optional)
	ExchangeRate              *float64 `json:"exchange_rate,omitempty"`
	AmountInPrimaryCurrency   *float64 `json:"amount_in_primary_currency,omitempty"`
}

// CreateRecurringIncomeResponse representa la respuesta después de crear
type CreateRecurringIncomeResponse struct {
	ID                        string   `json:"id"`
	AccountID                 string   `json:"account_id"`
	Description               string   `json:"description"`
	Amount                    float64  `json:"amount"`
	Currency                  string   `json:"currency"`
	CategoryID                *string  `json:"category_id,omitempty"`
	CategoryName              *string  `json:"category_name,omitempty"`
	FamilyMemberID            *string  `json:"family_member_id,omitempty"`
	FamilyMemberName          *string  `json:"family_member_name,omitempty"`
	RecurrenceFrequency       string   `json:"recurrence_frequency"`
	RecurrenceInterval        int      `json:"recurrence_interval"`
	RecurrenceDayOfMonth      *int     `json:"recurrence_day_of_month,omitempty"`
	RecurrenceDayOfWeek       *int     `json:"recurrence_day_of_week,omitempty"`
	StartDate                 string   `json:"start_date"`
	EndDate                   *string  `json:"end_date,omitempty"`
	TotalOccurrences          *int     `json:"total_occurrences,omitempty"`
	CurrentOccurrence         int      `json:"current_occurrence"`
	ExchangeRate              *float64 `json:"exchange_rate,omitempty"`
	AmountInPrimaryCurrency   *float64 `json:"amount_in_primary_currency,omitempty"`
	IsActive                  bool     `json:"is_active"`
	CreatedAt                 string   `json:"created_at"`
}

// CreateRecurringIncome maneja POST /api/recurring-expenses
func CreateRecurringIncome(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateRecurringIncomeRequest

		// Validar JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Datos inválidos",
				"details": err.Error(),
			})
			return
		}

		// Obtener account_id del contexto (middleware AccountMiddleware)
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

		// Validar start_date formato YYYY-MM-DD
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "start_date debe tener formato YYYY-MM-DD",
				"details": err.Error(),
			})
			return
		}

		// Validar end_date si existe
		var endDate *time.Time
		if req.EndDate != nil {
			parsed, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "end_date debe tener formato YYYY-MM-DD",
					"details": err.Error(),
				})
				return
			}
			
			// end_date debe ser >= start_date
			if parsed.Before(startDate) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "end_date debe ser mayor o igual a start_date",
				})
				return
			}
			
			endDate = &parsed
		}

		// Validación de negocio: monthly/yearly REQUIERE day_of_month
		if (req.RecurrenceFrequency == "monthly" || req.RecurrenceFrequency == "yearly") && req.RecurrenceDayOfMonth == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "monthly/yearly requiere recurrence_day_of_month (1-31)",
			})
			return
		}

		// Validación de negocio: weekly REQUIERE day_of_week
		if req.RecurrenceFrequency == "weekly" && req.RecurrenceDayOfWeek == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "weekly requiere recurrence_day_of_week (0=Domingo, 6=Sábado)",
			})
			return
		}

		// Validación: daily/yearly NO deben tener day_of_week
		if req.RecurrenceFrequency != "weekly" && req.RecurrenceDayOfWeek != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "recurrence_day_of_week solo aplica a frequency=weekly",
			})
			return
		}

		// Validación: daily/weekly NO deben tener day_of_month
		if req.RecurrenceFrequency != "monthly" && req.RecurrenceFrequency != "yearly" && req.RecurrenceDayOfMonth != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "recurrence_day_of_month solo aplica a frequency=monthly/yearly",
			})
			return
		}

		// Validar family_member_id si existe
		if req.FamilyMemberID != nil {
			var memberExists bool
			checkMemberQuery := `
				SELECT EXISTS(
					SELECT 1 FROM family_members 
					WHERE id = $1 AND account_id = $2
				)
			`
			err := pool.QueryRow(ctx, checkMemberQuery, *req.FamilyMemberID, accountID).Scan(&memberExists)
			if err != nil || !memberExists {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "family_member_id no pertenece a esta cuenta",
				})
				return
			}
		}

		// Obtener moneda primaria de la cuenta
		var primaryCurrency string
		accountQuery := "SELECT currency FROM accounts WHERE id = $1"
		err = pool.QueryRow(ctx, accountQuery, accountID).Scan(&primaryCurrency)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error obteniendo moneda de la cuenta",
			})
			return
		}

		// Calcular exchange_rate y amount_in_primary_currency (Modo 3 multi-currency)
		var exchangeRate float64
		var amountInPrimaryCurrency float64

		if req.Currency == primaryCurrency {
			// Modo 1: Misma moneda
			exchangeRate = 1.0
			amountInPrimaryCurrency = req.Amount
		} else if req.AmountInPrimaryCurrency != nil {
			// Modo 3: Usuario provee amount_in_primary_currency
			amountInPrimaryCurrency = *req.AmountInPrimaryCurrency
			exchangeRate = amountInPrimaryCurrency / req.Amount
		} else if req.ExchangeRate != nil {
			// Modo 2: Usuario provee exchange_rate
			exchangeRate = *req.ExchangeRate
			amountInPrimaryCurrency = req.Amount * exchangeRate
		} else {
			// Modo Auto: Buscar tasa en exchange_rates table
			rateQuery := `
				SELECT rate FROM exchange_rates 
				WHERE from_currency = $1 AND to_currency = $2 AND rate_date = $3
				ORDER BY created_at DESC LIMIT 1
			`
			err := pool.QueryRow(ctx, rateQuery, req.Currency, primaryCurrency, startDate.Format("2006-01-02")).Scan(&exchangeRate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "No se encontró tasa de cambio. Proporcione exchange_rate o amount_in_primary_currency",
				})
				return
			}
			amountInPrimaryCurrency = req.Amount * exchangeRate
		}

		// Validar valores calculados
		if exchangeRate <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "exchange_rate debe ser mayor a 0",
			})
			return
		}

		if amountInPrimaryCurrency <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "amount_in_primary_currency debe ser mayor a 0",
			})
			return
		}

		// Default recurrence_interval = 1
		interval := 1
		if req.RecurrenceInterval > 0 {
			interval = req.RecurrenceInterval
		}

		// INSERT en recurring_incomes
		insertQuery := `
			INSERT INTO recurring_incomes (
				account_id, description, amount, currency, category_id, family_member_id,
				recurrence_frequency, recurrence_interval, recurrence_day_of_month, recurrence_day_of_week,
				start_date, end_date, total_occurrences,
				exchange_rate, amount_in_primary_currency
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING id, current_occurrence, is_active, created_at
		`

		var recurringID string
		var currentOccurrence int
		var isActive bool
		var createdAt time.Time

		err = pool.QueryRow(
			ctx,
			insertQuery,
			accountID,
			req.Description,
			req.Amount,
			req.Currency,
			req.CategoryID,
			req.FamilyMemberID,
			req.RecurrenceFrequency,
			interval,
			req.RecurrenceDayOfMonth,
			req.RecurrenceDayOfWeek,
			startDate,
			endDate,
			req.TotalOccurrences,
			exchangeRate,
			amountInPrimaryCurrency,
		).Scan(&recurringID, &currentOccurrence, &isActive, &createdAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error creando ingreso recurrente",
				"details": err.Error(),
			})
			return
		}

		// Obtener nombres de category y family_member si existen (para response)
		var categoryName *string
		var familyMemberName *string

		if req.CategoryID != nil {
			var name string
			categoryQuery := "SELECT name FROM income_categories WHERE id = $1"
			err := pool.QueryRow(ctx, categoryQuery, *req.CategoryID).Scan(&name)
			if err == nil {
				categoryName = &name
			}
		}

		if req.FamilyMemberID != nil {
			var name string
			memberQuery := "SELECT name FROM family_members WHERE id = $1"
			err := pool.QueryRow(ctx, memberQuery, *req.FamilyMemberID).Scan(&name)
			if err == nil {
				familyMemberName = &name
			}
		}

		// Log de creación exitosa
		logger.Info("recurring_expense.created", "Ingreso recurrente creado", map[string]interface{}{
			"recurring_income_id": recurringID,
			"account_id":           accountID,
			"user_id":              userID,
			"description":          req.Description,
			"frequency":            req.RecurrenceFrequency,
			"amount":               req.Amount,
			"currency":             req.Currency,
			"ip":                   c.ClientIP(),
		})

		// Retornar respuesta
		response := CreateRecurringIncomeResponse{
			ID:                      recurringID,
			AccountID:               accountID.(string),
			Description:             req.Description,
			Amount:                  req.Amount,
			Currency:                req.Currency,
			CategoryID:              req.CategoryID,
			CategoryName:            categoryName,
			FamilyMemberID:          req.FamilyMemberID,
			FamilyMemberName:        familyMemberName,
			RecurrenceFrequency:     req.RecurrenceFrequency,
			RecurrenceInterval:      interval,
			RecurrenceDayOfMonth:    req.RecurrenceDayOfMonth,
			RecurrenceDayOfWeek:     req.RecurrenceDayOfWeek,
			StartDate:               startDate.Format("2006-01-02"),
			EndDate:                 req.EndDate,
			TotalOccurrences:        req.TotalOccurrences,
			CurrentOccurrence:       currentOccurrence,
			ExchangeRate:            &exchangeRate,
			AmountInPrimaryCurrency: &amountInPrimaryCurrency,
			IsActive:                isActive,
			CreatedAt:               createdAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":           "Ingreso recurrente creado exitosamente",
			"recurring_expense": response,
		})
	}
}
