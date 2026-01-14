package expenses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetExpense(db *pgxpool.Pool) gin.HandlerFunc {
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

		// Query expense with category name
		var expense ExpenseResponse
		var familyMemberID, categoryID, categoryName *string
		var date, endDate *time.Time
		var createdAt time.Time

		query := `
			SELECT e.id, e.account_id, e.family_member_id, e.category_id, 
			       ec.name as category_name, e.description, 
			       e.amount, e.currency, e.exchange_rate, e.amount_in_primary_currency,
			       e.expense_type, e.date, e.end_date, e.created_at
			FROM expenses e
			LEFT JOIN expense_categories ec ON e.category_id = ec.id
			WHERE e.id = $1 AND e.account_id = $2
		`

		err := db.QueryRow(c.Request.Context(), query, expenseID, accountID).Scan(
			&expense.ID,
			&expense.AccountID,
			&familyMemberID,
			&categoryID,
			&categoryName,
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

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "expense not found or does not belong to this account"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch expense: " + err.Error()})
			return
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
