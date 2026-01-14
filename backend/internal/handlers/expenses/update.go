package expenses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UpdateExpenseRequest struct {
	FamilyMemberID *string  `json:"family_member_id"` // Optional
	CategoryID     *string  `json:"category_id"`      // Optional: UUID
	Description    *string  `json:"description"`
	Amount         *float64 `json:"amount"`
	Currency       *string  `json:"currency"`
	ExpenseType    *string  `json:"expense_type"`
	Date           *string  `json:"date"`     // Format: YYYY-MM-DD
	EndDate        *string  `json:"end_date"` // Format: YYYY-MM-DD

	// Multi-currency fields (Modo 3)
	ExchangeRate            *float64 `json:"exchange_rate,omitempty"`
	AmountInPrimaryCurrency *float64 `json:"amount_in_primary_currency,omitempty"`
}

func UpdateExpense(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context (set by AccountMiddleware)
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Get expense ID from URL parameter
		expenseID := c.Param("id")
		if expenseID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expense_id is required"})
			return
		}

		var req UpdateExpenseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// First, get existing expense to recalculate multi-currency if needed
		var existingExpenseType, existingCurrency string
		var existingAmount, existingExchangeRate, existingAmountInPrimaryCurrency float64
		var existingDate string
		checkQuery := `SELECT expense_type, amount, currency, exchange_rate, amount_in_primary_currency, date::TEXT 
	               FROM expenses WHERE id = $1 AND account_id = $2`
		err := db.QueryRow(c.Request.Context(), checkQuery, expenseID, accountID).Scan(
			&existingExpenseType, &existingAmount, &existingCurrency,
			&existingExchangeRate, &existingAmountInPrimaryCurrency, &existingDate)

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "expense not found or does not belong to this account"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check expense: " + err.Error()})
			return
		}

		// Validate expense_type if provided
		if req.ExpenseType != nil {
			if *req.ExpenseType != "one-time" && *req.ExpenseType != "recurring" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "expense_type must be one-time or recurring"})
				return
			}
		}

		// Validate currency if provided
		if req.Currency != nil {
			validCurrencies := map[string]bool{"ARS": true, "USD": true, "EUR": true}
			if !validCurrencies[*req.Currency] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be ARS, USD, or EUR"})
				return
			}
		}

		// Validate amount if provided
		if req.Amount != nil && *req.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
			return
		}

		// Validate dates if provided
		var expenseDate time.Time
		if req.Date != nil {
			parsedDate, err := time.Parse("2006-01-02", *req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
				return
			}
			expenseDate = parsedDate
		}

		// Validate end_date logic
		finalExpenseType := existingExpenseType
		if req.ExpenseType != nil {
			finalExpenseType = *req.ExpenseType
		}

		if finalExpenseType == "one-time" && req.EndDate != nil && *req.EndDate != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "one-time expenses cannot have an end_date"})
			return
		}

		if req.EndDate != nil && *req.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
				return
			}

			// If date is being updated, check against new date
			if req.Date != nil && endDate.Before(expenseDate) {
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
		// If any currency-related field changed, recalculate exchange_rate and amount_in_primary_currency
		var finalExchangeRate *float64
		var finalAmountInPrimaryCurrency *float64

		// Check if currency-related fields were updated
		currencyFieldsChanged := req.Amount != nil || req.Currency != nil || req.ExchangeRate != nil || req.AmountInPrimaryCurrency != nil || req.Date != nil

		if currencyFieldsChanged {
			// Get primary currency of the account
			var primaryCurrency string
			err = db.QueryRow(c.Request.Context(),
				`SELECT currency FROM accounts WHERE id = $1`,
				accountID,
			).Scan(&primaryCurrency)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account currency"})
				return
			}

			// Use updated values or existing ones
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

			// Apply Modo 3 logic
			if finalCurrency == primaryCurrency {
				// Modo 1: Same currency
				rate := 1.0
				amountPrimary := finalAmount
				finalExchangeRate = &rate
				finalAmountInPrimaryCurrency = &amountPrimary
			} else {
				// Modo 3: User provided amount_in_primary_currency
				if req.AmountInPrimaryCurrency != nil {
					amountPrimary := *req.AmountInPrimaryCurrency
					rate := amountPrimary / finalAmount
					finalExchangeRate = &rate
					finalAmountInPrimaryCurrency = &amountPrimary
				} else if req.ExchangeRate != nil {
					// Modo 2: User provided exchange_rate
					rate := *req.ExchangeRate
					amountPrimary := finalAmount * rate
					finalExchangeRate = &rate
					finalAmountInPrimaryCurrency = &amountPrimary
				} else {
					// Try to fetch rate from exchange_rates table
					var rate float64
					err = db.QueryRow(c.Request.Context(),
						`SELECT rate FROM exchange_rates 
					 WHERE from_currency = $1 AND to_currency = $2 AND rate_date = $3
					 ORDER BY created_at DESC LIMIT 1`,
						finalCurrency, primaryCurrency, finalDate,
					).Scan(&rate)

					if err != nil {
						// Keep existing values if no new rate found
						finalExchangeRate = &existingExchangeRate
						finalAmountInPrimaryCurrency = &existingAmountInPrimaryCurrency
					} else {
						amountPrimary := finalAmount * rate
						finalExchangeRate = &rate
						finalAmountInPrimaryCurrency = &amountPrimary
					}
				}
			}

			// Validate
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
		UPDATE expenses SET
			family_member_id = COALESCE($1, family_member_id),
			category_id = COALESCE($2, category_id),
			description = COALESCE($3, description),
			amount = COALESCE($4, amount),
			currency = COALESCE($5, currency),
			expense_type = COALESCE($6, expense_type),
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
		          expense_type, date, end_date, created_at
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

		var expense ExpenseResponse
		var familyMemberID, categoryID *string
		var date, endDate *time.Time
		var createdAt time.Time

		err = db.QueryRow(c.Request.Context(), updateQuery,
			req.FamilyMemberID, req.CategoryID, req.Description,
			req.Amount, req.Currency, req.ExpenseType, req.Date,
			endDateParam, expenseID, accountID,
			finalExchangeRate, finalAmountInPrimaryCurrency,
		).Scan(
			&expense.ID,
			&expense.AccountID,
			&familyMemberID,
			&categoryID,
			&expense.Description,
			&expense.Amount,
			&expense.Currency,
			&expense.ExchangeRate,
			&expense.AmountInPrimaryCurrency,
			&expense.ExpenseType,
			&date,
			&endDate,
			&createdAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update expense: " + err.Error()})
			return
		}

		// Get category name if category_id exists
		var categoryName *string
		if categoryID != nil {
			var name string
			err = db.QueryRow(c.Request.Context(),
				`SELECT name FROM expense_categories WHERE id = $1`,
				categoryID,
			).Scan(&name)
			if err == nil {
				categoryName = &name
			}
		}

		// Set optional fields
		expense.FamilyMemberID = familyMemberID
		expense.CategoryID = categoryID
		expense.CategoryName = categoryName

		if date != nil {
			dateStr := date.Format("2006-01-02")
			expense.Date = dateStr
		}

		if endDate != nil {
			endDateStr := endDate.Format("2006-01-02")
			expense.EndDate = &endDateStr
		}

		expense.CreatedAt = createdAt.Format(time.RFC3339)

		c.JSON(http.StatusOK, expense)
	}
}
