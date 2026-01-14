package dashboard

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CategoryExpense represents expenses grouped by category
type CategoryExpense struct {
	CategoryID    *string `json:"category_id,omitempty"`
	CategoryName  *string `json:"category_name,omitempty"`
	CategoryIcon  *string `json:"category_icon,omitempty"`
	CategoryColor *string `json:"category_color,omitempty"`
	Total         float64 `json:"total"`
	Percentage    float64 `json:"percentage"`
}

// TopExpense represents a single top expense
type TopExpense struct {
	ID                      string  `json:"id"`
	Description             string  `json:"description"`
	Amount                  float64 `json:"amount"`
	Currency                string  `json:"currency"`
	AmountInPrimaryCurrency float64 `json:"amount_in_primary_currency"`
	CategoryName            *string `json:"category_name,omitempty"`
	Date                    string  `json:"date"`
}

// RecentTransaction represents a recent transaction (expense or income)
type RecentTransaction struct {
	ID                      string  `json:"id"`
	Type                    string  `json:"type"` // "expense" or "income"
	Description             string  `json:"description"`
	Amount                  float64 `json:"amount"`
	Currency                string  `json:"currency"`
	AmountInPrimaryCurrency float64 `json:"amount_in_primary_currency"`
	CategoryName            *string `json:"category_name,omitempty"`
	Date                    string  `json:"date"`
	CreatedAt               string  `json:"created_at"`
}

// DashboardSummaryResponse represents the complete dashboard summary
type DashboardSummaryResponse struct {
	Period               string              `json:"period"` // YYYY-MM format
	PrimaryCurrency      string              `json:"primary_currency"`
	TotalIncome          float64             `json:"total_income"`
	TotalExpenses        float64             `json:"total_expenses"`
	TotalAssignedToGoals float64             `json:"total_assigned_to_goals"` // Always 0 for now
	AvailableBalance     float64             `json:"available_balance"`
	ExpensesByCategory   []CategoryExpense   `json:"expenses_by_category"`
	TopExpenses          []TopExpense        `json:"top_expenses"`
	RecentTransactions   []RecentTransaction `json:"recent_transactions"`
}

// GetSummary handles GET /api/dashboard/summary
// Returns a complete summary of the user's financial data for a given month
func GetSummary(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context (set by AccountMiddleware)
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Parse query parameters (optional month/year, defaults to current month)
		month := c.DefaultQuery("month", time.Now().Format("2006-01"))

		// Validate month format (YYYY-MM)
		_, err := time.Parse("2006-01", month)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format, use YYYY-MM"})
			return
		}

		ctx := c.Request.Context()

		// Get primary currency of the account
		var primaryCurrency string
		err = db.QueryRow(ctx, `SELECT currency FROM accounts WHERE id = $1`, accountID).Scan(&primaryCurrency)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account currency"})
			return
		}

		// ============================================================================
		// 1. CALCULATE TOTAL INCOME (sum of amount_in_primary_currency)
		// ============================================================================
		var totalIncome float64
		incomeQuery := `
			SELECT COALESCE(SUM(amount_in_primary_currency), 0)
			FROM incomes
			WHERE account_id = $1
			  AND TO_CHAR(date, 'YYYY-MM') = $2
		`
		err = db.QueryRow(ctx, incomeQuery, accountID, month).Scan(&totalIncome)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate total income"})
			return
		}

		// ============================================================================
		// 2. CALCULATE TOTAL EXPENSES (sum of amount_in_primary_currency)
		// ============================================================================
		var totalExpenses float64
		expensesQuery := `
			SELECT COALESCE(SUM(amount_in_primary_currency), 0)
			FROM expenses
			WHERE account_id = $1
			  AND TO_CHAR(date, 'YYYY-MM') = $2
		`
		err = db.QueryRow(ctx, expensesQuery, accountID, month).Scan(&totalExpenses)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate total expenses"})
			return
		}

		// ============================================================================
		// 3. EXPENSES BY CATEGORY (with percentages)
		// ============================================================================
		categoryQuery := `
			SELECT 
				e.category_id,
				ec.name as category_name,
				ec.icon as category_icon,
				ec.color as category_color,
				SUM(e.amount_in_primary_currency) as total
			FROM expenses e
			LEFT JOIN expense_categories ec ON e.category_id = ec.id
			WHERE e.account_id = $1
			  AND TO_CHAR(e.date, 'YYYY-MM') = $2
			GROUP BY e.category_id, ec.name, ec.icon, ec.color
			HAVING SUM(e.amount_in_primary_currency) > 0
			ORDER BY total DESC
		`

		rows, err := db.Query(ctx, categoryQuery, accountID, month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get expenses by category"})
			return
		}
		defer rows.Close()

		expensesByCategory := []CategoryExpense{}
		for rows.Next() {
			var cat CategoryExpense
			err := rows.Scan(&cat.CategoryID, &cat.CategoryName, &cat.CategoryIcon, &cat.CategoryColor, &cat.Total)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse category expense"})
				return
			}

			// Calculate percentage
			if totalExpenses > 0 {
				cat.Percentage = (cat.Total / totalExpenses) * 100
			} else {
				cat.Percentage = 0
			}

			expensesByCategory = append(expensesByCategory, cat)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading category expenses"})
			return
		}

		// ============================================================================
		// 4. TOP 5 EXPENSES (ordered by amount_in_primary_currency)
		// ============================================================================
		topExpensesQuery := `
			SELECT 
				e.id,
				e.description,
				e.amount,
				e.currency,
				e.amount_in_primary_currency,
				ec.name as category_name,
				e.date::TEXT
			FROM expenses e
			LEFT JOIN expense_categories ec ON e.category_id = ec.id
			WHERE e.account_id = $1
			  AND TO_CHAR(e.date, 'YYYY-MM') = $2
			ORDER BY e.amount_in_primary_currency DESC
			LIMIT 5
		`

		rows, err = db.Query(ctx, topExpensesQuery, accountID, month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get top expenses"})
			return
		}
		defer rows.Close()

		topExpenses := []TopExpense{}
		for rows.Next() {
			var exp TopExpense
			err := rows.Scan(&exp.ID, &exp.Description, &exp.Amount, &exp.Currency,
				&exp.AmountInPrimaryCurrency, &exp.CategoryName, &exp.Date)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse top expense"})
				return
			}
			topExpenses = append(topExpenses, exp)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading top expenses"})
			return
		}

		// ============================================================================
		// 5. RECENT TRANSACTIONS (last 10 expenses + incomes, merged and sorted)
		// ============================================================================
		recentTransactionsQuery := `
			(
				SELECT 
					e.id,
					'expense' as type,
					e.description,
					e.amount,
					e.currency,
					e.amount_in_primary_currency,
					ec.name as category_name,
					e.date::TEXT,
					e.created_at::TEXT
				FROM expenses e
				LEFT JOIN expense_categories ec ON e.category_id = ec.id
				WHERE e.account_id = $1
				  AND TO_CHAR(e.date, 'YYYY-MM') = $2
			)
			UNION ALL
			(
				SELECT 
					i.id,
					'income' as type,
					i.description,
					i.amount,
					i.currency,
					i.amount_in_primary_currency,
					ic.name as category_name,
					i.date::TEXT,
					i.created_at::TEXT
				FROM incomes i
				LEFT JOIN income_categories ic ON i.category_id = ic.id
				WHERE i.account_id = $1
				  AND TO_CHAR(i.date, 'YYYY-MM') = $2
			)
			ORDER BY created_at DESC
			LIMIT 10
		`

		rows, err = db.Query(ctx, recentTransactionsQuery, accountID, month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get recent transactions"})
			return
		}
		defer rows.Close()

		recentTransactions := []RecentTransaction{}
		for rows.Next() {
			var txn RecentTransaction
			err := rows.Scan(&txn.ID, &txn.Type, &txn.Description, &txn.Amount, &txn.Currency,
				&txn.AmountInPrimaryCurrency, &txn.CategoryName, &txn.Date, &txn.CreatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse transaction"})
				return
			}
			recentTransactions = append(recentTransactions, txn)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading recent transactions"})
			return
		}

		// ============================================================================
		// 6. CALCULATE TOTAL ASSIGNED TO SAVINGS GOALS
		// ============================================================================
		var totalAssignedToGoals float64
		goalsQuery := `
		SELECT COALESCE(SUM(current_amount), 0)
		FROM savings_goals
		WHERE account_id = $1 AND is_active = true
	`
		err = db.QueryRow(ctx, goalsQuery, accountID).Scan(&totalAssignedToGoals)
		if err != nil {
			// If there's an error, just set to 0 instead of failing the entire request
			totalAssignedToGoals = 0
		}

		// ============================================================================
		// 7. CALCULATE AVAILABLE BALANCE
		// ============================================================================
		availableBalance := totalIncome - totalExpenses - totalAssignedToGoals

		// ============================================================================
		// BUILD RESPONSE
		// ============================================================================
		response := DashboardSummaryResponse{
			Period:               month,
			PrimaryCurrency:      primaryCurrency,
			TotalIncome:          totalIncome,
			TotalExpenses:        totalExpenses,
			TotalAssignedToGoals: totalAssignedToGoals,
			AvailableBalance:     availableBalance,
			ExpensesByCategory:   expensesByCategory,
			TopExpenses:          topExpenses,
			RecentTransactions:   recentTransactions,
		}

		c.JSON(http.StatusOK, response)
	}
}
