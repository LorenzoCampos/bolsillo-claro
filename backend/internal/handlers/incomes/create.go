package incomes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateIncomeRequest struct {
	FamilyMemberID *string `json:"family_member_id"` // Optional: for family accounts
	CategoryID     *string `json:"category_id"`      // Optional
	Description    string  `json:"description" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Currency       string  `json:"currency" binding:"required,oneof=ARS USD EUR"`
	IncomeType     string  `json:"income_type" binding:"required,oneof=one-time recurring"`
	Date           string  `json:"date" binding:"required"` // Format: YYYY-MM-DD
	EndDate        *string `json:"end_date"`                // Optional: for recurring

	// Multi-currency fields (Modo 3: Flexibilidad Total)
	ExchangeRate            *float64 `json:"exchange_rate,omitempty"`              // Optional: tasa de conversión
	AmountInPrimaryCurrency *float64 `json:"amount_in_primary_currency,omitempty"` // Optional: monto REAL recibido en moneda primaria
}

type IncomeResponse struct {
	ID                      string  `json:"id"`
	AccountID               string  `json:"account_id"`
	FamilyMemberID          *string `json:"family_member_id,omitempty"`
	CategoryID              *string `json:"category_id,omitempty"`
	CategoryName            *string `json:"category_name,omitempty"`
	Description             string  `json:"description"`
	Amount                  float64 `json:"amount"`
	Currency                string  `json:"currency"`
	ExchangeRate            float64 `json:"exchange_rate"`              // Tasa de conversión (snapshot)
	AmountInPrimaryCurrency float64 `json:"amount_in_primary_currency"` // Monto en moneda primaria
	IncomeType              string  `json:"income_type"`
	Date                    string  `json:"date"`
	EndDate                 *string `json:"end_date,omitempty"`
	CreatedAt               string  `json:"created_at"`
}

func CreateIncome(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateIncomeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get account_id from context (set by AccountMiddleware)
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Validate date format
		incomeDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}

		// Validate end_date logic
		if req.IncomeType == "one-time" && req.EndDate != nil {
			// One-time incomes should NOT have end_date
			c.JSON(http.StatusBadRequest, gin.H{"error": "one-time incomes cannot have an end_date"})
			return
		}

		// If recurring has end_date, validate it
		if req.IncomeType == "recurring" && req.EndDate != nil {
			endDate, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
				return
			}

			if endDate.Before(incomeDate) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after or equal to date"})
				return
			}
		}

		// If family_member_id is provided, validate it belongs to this account
		if req.FamilyMemberID != nil {
			var exists bool
			err := db.QueryRow(c.Request.Context(),
				`SELECT EXISTS(
				SELECT 1 FROM family_members 
				WHERE id = $1 AND account_id = $2
			)`,
				req.FamilyMemberID, accountID,
			).Scan(&exists)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate family member"})
				return
			}
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "family_member_id does not belong to this account"})
				return
			}
		}

		// ============================================================================
		// MULTI-CURRENCY LOGIC - Modo 3: Flexibilidad Total
		// ============================================================================
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

		var exchangeRate float64
		var amountInPrimaryCurrency float64

		// Modo 1: Same currency as primary (ARS = ARS)
		if req.Currency == primaryCurrency {
			exchangeRate = 1.0
			amountInPrimaryCurrency = req.Amount
		} else {
			// Modo 3: User provided amount_in_primary_currency (REAL amount received)
			if req.AmountInPrimaryCurrency != nil {
				// Calculate effective exchange rate
				amountInPrimaryCurrency = *req.AmountInPrimaryCurrency
				exchangeRate = amountInPrimaryCurrency / req.Amount
			} else if req.ExchangeRate != nil {
				// Modo 2: User provided exchange_rate
				exchangeRate = *req.ExchangeRate
				amountInPrimaryCurrency = req.Amount * exchangeRate
			} else {
				// Try to fetch rate from exchange_rates table
				err = db.QueryRow(c.Request.Context(),
					`SELECT rate FROM exchange_rates 
				 WHERE from_currency = $1 AND to_currency = $2 AND rate_date = $3
				 ORDER BY created_at DESC LIMIT 1`,
					req.Currency, primaryCurrency, req.Date,
				).Scan(&exchangeRate)

				if err != nil {
					// No rate found - require user to provide it
					c.JSON(http.StatusBadRequest, gin.H{
						"error":      "no exchange rate found for this date",
						"suggestion": "please provide either 'exchange_rate' or 'amount_in_primary_currency'",
						"details": map[string]string{
							"from_currency": req.Currency,
							"to_currency":   primaryCurrency,
							"date":          req.Date,
						},
					})
					return
				}

				amountInPrimaryCurrency = req.Amount * exchangeRate
			}
		}

		// Validate calculated values
		if exchangeRate <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "exchange_rate must be positive"})
			return
		}
		if amountInPrimaryCurrency <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount_in_primary_currency must be positive"})
			return
		}

		// Insert income with multi-currency fields
		var incomeID uuid.UUID
		var createdAt time.Time

		err = db.QueryRow(c.Request.Context(),
			`INSERT INTO incomes (
			account_id, family_member_id, category_id, description, 
			amount, currency, exchange_rate, amount_in_primary_currency,
			income_type, date, end_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`,
			accountID, req.FamilyMemberID, req.CategoryID, req.Description,
			req.Amount, req.Currency, exchangeRate, amountInPrimaryCurrency,
			req.IncomeType, req.Date, req.EndDate,
		).Scan(&incomeID, &createdAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create income: " + err.Error()})
			return
		}

		// Get category name if category_id was provided
		var categoryName *string
		if req.CategoryID != nil {
			var name string
			err = db.QueryRow(c.Request.Context(),
				`SELECT name FROM income_categories WHERE id = $1`,
				req.CategoryID,
			).Scan(&name)
			if err == nil {
				categoryName = &name
			}
		}

		// Build response
		response := IncomeResponse{
			ID:                      incomeID.String(),
			AccountID:               accountID.(string),
			FamilyMemberID:          req.FamilyMemberID,
			CategoryID:              req.CategoryID,
			CategoryName:            categoryName,
			Description:             req.Description,
			Amount:                  req.Amount,
			Currency:                req.Currency,
			ExchangeRate:            exchangeRate,
			AmountInPrimaryCurrency: amountInPrimaryCurrency,
			IncomeType:              req.IncomeType,
			Date:                    req.Date,
			EndDate:                 req.EndDate,
			CreatedAt:               createdAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusCreated, response)
	}
}
