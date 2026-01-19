package incomes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

type UpdateIncomeRequest struct {
	FamilyMemberID *string  `json:"family_member_id"` // Optional
	CategoryID     *string  `json:"category_id"`      // Optional: UUID
	Description    *string  `json:"description"`
	Amount         *float64 `json:"amount"`
	Currency       *string  `json:"currency"`
	IncomeType     *string  `json:"income_type"`
	Date           *string  `json:"date"`     // Format: YYYY-MM-DD
	EndDate        *string  `json:"end_date"` // Format: YYYY-MM-DD

	// Multi-currency fields (Modo 3)
	ExchangeRate            *float64 `json:"exchange_rate,omitempty"`
	AmountInPrimaryCurrency *float64 `json:"amount_in_primary_currency,omitempty"`
}

func UpdateIncome(db *pgxpool.Pool) gin.HandlerFunc {
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

		var req UpdateIncomeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// First, get existing income to recalculate multi-currency if needed
		var existingIncomeType, existingCurrency string
		var existingAmount, existingExchangeRate, existingAmountInPrimaryCurrency float64
		var existingDate string
		checkQuery := `SELECT income_type, amount, currency, exchange_rate, amount_in_primary_currency, date::TEXT 
		               FROM incomes WHERE id = $1 AND account_id = $2`
		err := db.QueryRow(c.Request.Context(), checkQuery, incomeID, accountID).Scan(
			&existingIncomeType, &existingAmount, &existingCurrency,
			&existingExchangeRate, &existingAmountInPrimaryCurrency, &existingDate)

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "income not found or does not belong to this account"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check income: " + err.Error()})
			return
		}

		// Validate income_type if provided
		if req.IncomeType != nil {
			if *req.IncomeType != "one-time" && *req.IncomeType != "recurring" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "income_type must be one-time or recurring"})
				return
			}
		}

		// Validate currency if provided
		if req.Currency != nil {
			validCurrencies := map[string]bool{"ARS": true, "USD": true}
			if !validCurrencies[*req.Currency] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be ARS or USD"})
				return
			}
		}

		// Validate amount if provided
		if req.Amount != nil && *req.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
			return
		}

		// Validate dates if provided
		var incomeDate time.Time
		if req.Date != nil {
			parsedDate, err := time.Parse("2006-01-02", *req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
				return
			}
			incomeDate = parsedDate
		}

		// Validate end_date logic
		finalIncomeType := existingIncomeType
		if req.IncomeType != nil {
			finalIncomeType = *req.IncomeType
		}

		if finalIncomeType == "one-time" && req.EndDate != nil && *req.EndDate != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "one-time incomes cannot have an end_date"})
			return
		}

		if req.EndDate != nil && *req.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
				return
			}

			// If date is being updated, check against new date
			if req.Date != nil && endDate.Before(incomeDate) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after or equal to date"})
				return
			}
		}

		// If family_member_id is provided, validate it belongs to this account
		if req.FamilyMemberID != nil && *req.FamilyMemberID != "" {
			var memberExists bool
			err := db.QueryRow(c.Request.Context(),
				`SELECT EXISTS(
					SELECT 1 FROM family_members 
					WHERE id = $1 AND account_id = $2
				)`,
				req.FamilyMemberID, accountID,
			).Scan(&memberExists)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate family member"})
				return
			}
			if !memberExists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "family_member_id does not belong to this account"})
				return
			}
		}

		// ============================================================================
		// MULTI-CURRENCY RECALCULATION - Modo 3
		// ============================================================================
		var finalExchangeRate *float64
		var finalAmountInPrimaryCurrency *float64

		currencyFieldsChanged := req.Amount != nil || req.Currency != nil || req.ExchangeRate != nil || req.AmountInPrimaryCurrency != nil || req.Date != nil

		if currencyFieldsChanged {
			var primaryCurrency string
			err = db.QueryRow(c.Request.Context(),
				`SELECT currency FROM accounts WHERE id = $1`,
				accountID,
			).Scan(&primaryCurrency)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account currency"})
				return
			}

			finalAmount := existingAmount
			if req.Amount != nil {
				finalAmount = *req.Amount
			}

			finalCurrency := existingCurrency
			if req.Currency != nil {
				finalCurrency = *req.Currency
			}

			finalDate := existingDate
			if req.Date != nil {
				finalDate = *req.Date
			}

			if finalCurrency == primaryCurrency {
				rate := 1.0
				amountPrimary := finalAmount
				finalExchangeRate = &rate
				finalAmountInPrimaryCurrency = &amountPrimary
			} else {
				if req.AmountInPrimaryCurrency != nil {
					amountPrimary := *req.AmountInPrimaryCurrency
					rate := amountPrimary / finalAmount
					finalExchangeRate = &rate
					finalAmountInPrimaryCurrency = &amountPrimary
				} else if req.ExchangeRate != nil {
					rate := *req.ExchangeRate
					amountPrimary := finalAmount * rate
					finalExchangeRate = &rate
					finalAmountInPrimaryCurrency = &amountPrimary
				} else {
					var rate float64
					err = db.QueryRow(c.Request.Context(),
						`SELECT rate FROM exchange_rates 
						 WHERE from_currency = $1 AND to_currency = $2 AND rate_date = $3
						 ORDER BY created_at DESC LIMIT 1`,
						finalCurrency, primaryCurrency, finalDate,
					).Scan(&rate)

					if err != nil {
						finalExchangeRate = &existingExchangeRate
						finalAmountInPrimaryCurrency = &existingAmountInPrimaryCurrency
					} else {
						amountPrimary := finalAmount * rate
						finalExchangeRate = &rate
						finalAmountInPrimaryCurrency = &amountPrimary
					}
				}
			}

			if *finalExchangeRate <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "exchange_rate must be positive"})
				return
			}
			if *finalAmountInPrimaryCurrency <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "amount_in_primary_currency must be positive"})
				return
			}
		}

		// Build dynamic UPDATE query
		updateQuery := `
			UPDATE incomes SET
				family_member_id = COALESCE($1, family_member_id),
				category_id = COALESCE($2, category_id),
				description = COALESCE($3, description),
				amount = COALESCE($4, amount),
				currency = COALESCE($5, currency),
				income_type = COALESCE($6, income_type),
				date = COALESCE($7, date),
				end_date = CASE 
					WHEN $8::text = 'CLEAR' THEN NULL
					WHEN $8::uuid IS NOT NULL THEN $8::date
					ELSE end_date
				END,
				exchange_rate = COALESCE($11, exchange_rate),
				amount_in_primary_currency = COALESCE($12, amount_in_primary_currency),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $9 AND account_id = $10
			RETURNING id, account_id, family_member_id, category_id, description, 
			          amount, currency, exchange_rate, amount_in_primary_currency,
			          income_type, date, end_date, created_at
		`

		// Handle end_date special case: empty string means clear it
		var endDateParam *string
		if req.EndDate != nil {
			if *req.EndDate == "" {
				clearValue := "CLEAR"
				endDateParam = &clearValue
			} else {
				endDateParam = req.EndDate
			}
		}

		var income IncomeResponse
		var familyMemberID, categoryID *string
		var date, endDate *time.Time
		var createdAt time.Time

		err = db.QueryRow(c.Request.Context(), updateQuery,
			req.FamilyMemberID, req.CategoryID, req.Description,
			req.Amount, req.Currency, req.IncomeType, req.Date,
			endDateParam, incomeID, accountID,
			finalExchangeRate, finalAmountInPrimaryCurrency,
		).Scan(
			&income.ID,
			&income.AccountID,
			&familyMemberID,
			&categoryID,
			&income.Description,
			&income.Amount,
			&income.Currency,
			&income.ExchangeRate,
			&income.AmountInPrimaryCurrency,
			&income.IncomeType,
			&date,
			&endDate,
			&createdAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update income: " + err.Error()})
			return
		}

		// Get category name if category_id exists
		var categoryName *string
		if categoryID != nil {
			var name string
			err = db.QueryRow(c.Request.Context(),
				`SELECT name FROM income_categories WHERE id = $1`,
				categoryID,
			).Scan(&name)
			if err == nil {
				categoryName = &name
			}
		}

		// Set optional fields
		income.FamilyMemberID = familyMemberID
		income.CategoryID = categoryID
		income.CategoryName = categoryName

		if date != nil {
			dateStr := date.Format("2006-01-02")
			income.Date = dateStr
		}

		if endDate != nil {
			endDateStr := endDate.Format("2006-01-02")
			income.EndDate = &endDateStr
		}

		income.CreatedAt = createdAt.Format(time.RFC3339)

		// Obtener user_id del contexto para logging
		userID, _ := middleware.GetUserID(c)

		// Log de actualización exitosa
		logger.Info("income.updated", "Ingreso actualizado", map[string]interface{}{
			"income_id":  incomeID,
			"account_id": accountID,
			"user_id":    userID,
			"ip":         c.ClientIP(),
		})

		c.JSON(http.StatusOK, income)
	}
}
