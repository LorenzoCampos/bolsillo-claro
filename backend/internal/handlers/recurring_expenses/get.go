package recurring_expenses

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RecurringExpenseDetail representa el detalle completo de un gasto recurrente
type RecurringExpenseDetail struct {
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
	UpdatedAt                 string   `json:"updated_at"`
	GeneratedExpensesCount    int      `json:"generated_expenses_count"` // Cuántos gastos se generaron
}

// GetRecurringExpense maneja GET /api/recurring-expenses/:id
func GetRecurringExpense(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recurringID := c.Param("id")

		// Obtener account_id del contexto
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Account-ID header requerido",
			})
			return
		}

		ctx := c.Request.Context()

		// Query principal
		query := `
			SELECT 
				re.id,
				re.account_id,
				re.description,
				re.amount,
				re.currency,
				re.category_id,
				ec.name AS category_name,
				re.family_member_id,
				fm.name AS family_member_name,
				re.recurrence_frequency,
				re.recurrence_interval,
				re.recurrence_day_of_month,
				re.recurrence_day_of_week,
				re.start_date,
				re.end_date,
				re.total_occurrences,
				re.current_occurrence,
				re.exchange_rate,
				re.amount_in_primary_currency,
				re.is_active,
				re.created_at,
				re.updated_at
			FROM recurring_expenses re
			LEFT JOIN expense_categories ec ON re.category_id = ec.id
			LEFT JOIN family_members fm ON re.family_member_id = fm.id
			WHERE re.id = $1 AND re.account_id = $2
		`

		var detail RecurringExpenseDetail
		var categoryID, categoryName, familyMemberID, familyMemberName *string
		var dayOfMonth, dayOfWeek, totalOccurrences *int
		var exchangeRate, amountInPrimaryCurrency *float64
		var startDate, endDate, createdAt, updatedAt interface{}

		err := pool.QueryRow(ctx, query, recurringID, accountID).Scan(
			&detail.ID,
			&detail.AccountID,
			&detail.Description,
			&detail.Amount,
			&detail.Currency,
			&categoryID,
			&categoryName,
			&familyMemberID,
			&familyMemberName,
			&detail.RecurrenceFrequency,
			&detail.RecurrenceInterval,
			&dayOfMonth,
			&dayOfWeek,
			&startDate,
			&endDate,
			&totalOccurrences,
			&detail.CurrentOccurrence,
			&exchangeRate,
			&amountInPrimaryCurrency,
			&detail.IsActive,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Gasto recurrente no encontrado",
			})
			return
		}

		// Asignar opcionales
		detail.CategoryID = categoryID
		detail.CategoryName = categoryName
		detail.FamilyMemberID = familyMemberID
		detail.FamilyMemberName = familyMemberName
		detail.RecurrenceDayOfMonth = dayOfMonth
		detail.RecurrenceDayOfWeek = dayOfWeek
		detail.TotalOccurrences = totalOccurrences
		detail.ExchangeRate = exchangeRate
		detail.AmountInPrimaryCurrency = amountInPrimaryCurrency
		
		// Convertir dates a string
		if startDate != nil {
			detail.StartDate = fmt.Sprint(startDate)
		}
		
		if endDate != nil {
			endDateStr := fmt.Sprint(endDate)
			detail.EndDate = &endDateStr
		}
		
		if createdAt != nil {
			detail.CreatedAt = fmt.Sprint(createdAt)
		}
		
		if updatedAt != nil {
			detail.UpdatedAt = fmt.Sprint(updatedAt)
		}

		// Contar cuántos gastos se generaron desde este template
		countQuery := `
			SELECT COUNT(*) 
			FROM expenses 
			WHERE recurring_expense_id = $1
		`
		err = pool.QueryRow(ctx, countQuery, recurringID).Scan(&detail.GeneratedExpensesCount)
		if err != nil {
			detail.GeneratedExpensesCount = 0
		}

		c.JSON(http.StatusOK, detail)
	}
}
