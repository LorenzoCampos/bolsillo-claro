package recurring_expenses

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RecurringExpenseListItem representa un item en la lista
type RecurringExpenseListItem struct {
	ID                      string   `json:"id"`
	Description             string   `json:"description"`
	Amount                  float64  `json:"amount"`
	Currency                string   `json:"currency"`
	CategoryName            *string  `json:"category_name,omitempty"`
	FamilyMemberName        *string  `json:"family_member_name,omitempty"`
	RecurrenceFrequency     string   `json:"recurrence_frequency"`
	RecurrenceInterval      int      `json:"recurrence_interval"`
	RecurrenceDayOfMonth    *int     `json:"recurrence_day_of_month,omitempty"`
	RecurrenceDayOfWeek     *int     `json:"recurrence_day_of_week,omitempty"`
	StartDate               string   `json:"start_date"`
	EndDate                 *string  `json:"end_date,omitempty"`
	TotalOccurrences        *int     `json:"total_occurrences,omitempty"`
	CurrentOccurrence       int      `json:"current_occurrence"`
	IsActive                bool     `json:"is_active"`
	CreatedAt               string   `json:"created_at"`
}

// ListRecurringExpenses maneja GET /api/recurring-expenses
func ListRecurringExpenses(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener account_id del contexto
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Account-ID header requerido",
			})
			return
		}

		ctx := c.Request.Context()

		// Query params opcionales
		isActiveParam := c.DefaultQuery("is_active", "true")
		frequency := c.Query("frequency")
		
		// Paginación
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}

		offset := (page - 1) * limit

		// Construir query dinámicamente
		baseQuery := `
			SELECT 
				re.id,
				re.description,
				re.amount,
				re.currency,
				ec.name AS category_name,
				fm.name AS family_member_name,
				re.recurrence_frequency,
				re.recurrence_interval,
				re.recurrence_day_of_month,
				re.recurrence_day_of_week,
				re.start_date,
				re.end_date,
				re.total_occurrences,
				re.current_occurrence,
				re.is_active,
				re.created_at
			FROM recurring_expenses re
			LEFT JOIN expense_categories ec ON re.category_id = ec.id
			LEFT JOIN family_members fm ON re.family_member_id = fm.id
			WHERE re.account_id = $1
		`

		args := []interface{}{accountID}
		argCount := 1

		// Filtro por is_active
		if isActiveParam != "all" {
			argCount++
			isActive := isActiveParam == "true"
			baseQuery += " AND re.is_active = $" + strconv.Itoa(argCount)
			args = append(args, isActive)
		}

		// Filtro por frequency
		if frequency != "" {
			argCount++
			baseQuery += " AND re.recurrence_frequency = $" + strconv.Itoa(argCount)
			args = append(args, frequency)
		}

		// Ordenar por created_at DESC
		baseQuery += " ORDER BY re.created_at DESC"

		// Paginación
		argCount++
		baseQuery += " LIMIT $" + strconv.Itoa(argCount)
		args = append(args, limit)

		argCount++
		baseQuery += " OFFSET $" + strconv.Itoa(argCount)
		args = append(args, offset)

		// Ejecutar query
		rows, err := pool.Query(ctx, baseQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error consultando gastos recurrentes",
				"details": err.Error(),
			})
			return
		}
		defer rows.Close()

		// Parsear resultados
		var recurringExpenses []RecurringExpenseListItem

		for rows.Next() {
			var item RecurringExpenseListItem
			var categoryName, familyMemberName *string
			var dayOfMonth, dayOfWeek, totalOccurrences *int
			var startDate, endDate, createdAt interface{}

			err := rows.Scan(
				&item.ID,
				&item.Description,
				&item.Amount,
				&item.Currency,
				&categoryName,
				&familyMemberName,
				&item.RecurrenceFrequency,
				&item.RecurrenceInterval,
				&dayOfMonth,
				&dayOfWeek,
				&startDate,
				&endDate,
				&totalOccurrences,
				&item.CurrentOccurrence,
				&item.IsActive,
				&createdAt,
			)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Error parseando resultados",
					"details": err.Error(),
				})
				return
			}

			// Asignar opcionales
			item.CategoryName = categoryName
			item.FamilyMemberName = familyMemberName
			item.RecurrenceDayOfMonth = dayOfMonth
			item.RecurrenceDayOfWeek = dayOfWeek
			item.TotalOccurrences = totalOccurrences
			
			// Convertir dates a string
			if startDate != nil {
				item.StartDate = fmt.Sprint(startDate)
			}
			
			if endDate != nil {
				endDateStr := fmt.Sprint(endDate)
				item.EndDate = &endDateStr
			}
			
			if createdAt != nil {
				item.CreatedAt = fmt.Sprint(createdAt)
			}

			recurringExpenses = append(recurringExpenses, item)
		}

		// Contar total (para paginación)
		countQuery := `
			SELECT COUNT(*) 
			FROM recurring_expenses 
			WHERE account_id = $1
		`
		
		countArgs := []interface{}{accountID}
		
		if isActiveParam != "all" {
			isActive := isActiveParam == "true"
			countQuery += " AND is_active = $2"
			countArgs = append(countArgs, isActive)
		}

		var total int
		err = pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
		if err != nil {
			total = 0
		}

		// Retornar respuesta
		c.JSON(http.StatusOK, gin.H{
			"recurring_expenses": recurringExpenses,
			"count":              len(recurringExpenses),
			"total":              total,
			"page":               page,
			"limit":              limit,
		})
	}
}
