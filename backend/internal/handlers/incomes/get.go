package incomes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetIncome(db *pgxpool.Pool) gin.HandlerFunc {
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

		// Query income ensuring it belongs to the user's account
		var income IncomeResponse
		var familyMemberID, categoryID, categoryName *string
		var date, endDate *time.Time
		var createdAt time.Time

		query := `
			SELECT i.id, i.account_id, i.family_member_id, i.category_id, ic.name as category_name, i.description, 
			       i.amount, i.currency, i.exchange_rate, i.amount_in_primary_currency,
			       i.income_type, i.date, i.end_date, i.created_at
			FROM incomes i
			LEFT JOIN income_categories ic ON i.category_id = ic.id
			WHERE i.id = $1 AND i.account_id = $2
		`

		err := db.QueryRow(c.Request.Context(), query, incomeID, accountID).Scan(
			&income.ID,
			&income.AccountID,
			&familyMemberID,
			&categoryID,
			&categoryName,
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

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "income not found or does not belong to this account"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch income: " + err.Error()})
			return
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

		c.JSON(http.StatusOK, income)
	}
}
