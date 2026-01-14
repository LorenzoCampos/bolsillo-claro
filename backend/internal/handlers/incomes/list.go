package incomes

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListIncomesQuery struct {
	DateFrom       string `form:"date_from"`        // YYYY-MM-DD
	DateTo         string `form:"date_to"`          // YYYY-MM-DD
	IncomeType     string `form:"income_type"`      // one-time, recurring
	CategoryID     string `form:"category_id"`      // Categoría exacta
	FamilyMemberID string `form:"family_member_id"` // UUID
	SortBy         string `form:"sort_by"`          // date, amount, created_at
	Order          string `form:"order"`            // asc, desc
	Page           int    `form:"page"`             // Página (default: 1)
	Limit          int    `form:"limit"`            // Items por página (default: 20, max: 100)
}

type IncomeListItem struct {
	ID                      string  `json:"id"`
	FamilyMemberID          *string `json:"family_member_id,omitempty"`
	CategoryID              *string `json:"category_id,omitempty"`
	CategoryName            *string `json:"category_name,omitempty"`
	Description             string  `json:"description"`
	Amount                  float64 `json:"amount"`
	Currency                string  `json:"currency"`
	ExchangeRate            float64 `json:"exchange_rate"`
	AmountInPrimaryCurrency float64 `json:"amount_in_primary_currency"`
	IncomeType              string  `json:"income_type"`
	Date                    string  `json:"date"`
	EndDate                 *string `json:"end_date,omitempty"`
	CreatedAt               string  `json:"created_at"`
}

type ListIncomesResponse struct {
	Incomes    []IncomeListItem `json:"incomes"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

func ListIncomes(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context (set by AccountMiddleware)
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Parse query parameters
		var query ListIncomesQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set defaults
		if query.Page < 1 {
			query.Page = 1
		}
		if query.Limit < 1 {
			query.Limit = 20
		}
		if query.Limit > 100 {
			query.Limit = 100
		}
		if query.SortBy == "" {
			query.SortBy = "date"
		}
		if query.Order == "" {
			query.Order = "desc"
		}

		// Validate sort_by
		validSortFields := map[string]bool{
			"date":       true,
			"amount":     true,
			"created_at": true,
		}
		if !validSortFields[query.SortBy] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort_by field"})
			return
		}

		// Validate order
		if query.Order != "asc" && query.Order != "desc" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order must be asc or desc"})
			return
		}

		// Validate dates if provided
		if query.DateFrom != "" {
			if _, err := time.Parse("2006-01-02", query.DateFrom); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from format, use YYYY-MM-DD"})
				return
			}
		}
		if query.DateTo != "" {
			if _, err := time.Parse("2006-01-02", query.DateTo); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to format, use YYYY-MM-DD"})
				return
			}
		}

		// Validate income_type if provided
		if query.IncomeType != "" && query.IncomeType != "one-time" && query.IncomeType != "recurring" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "income_type must be one-time or recurring"})
			return
		}

		// Build WHERE clauses dynamically
		whereClauses := []string{"i.account_id = $1"}
		args := []interface{}{accountID}
		argIndex := 2

		if query.DateFrom != "" {
			whereClauses = append(whereClauses, "i.date >= $"+strconv.Itoa(argIndex))
			args = append(args, query.DateFrom)
			argIndex++
		}

		if query.DateTo != "" {
			whereClauses = append(whereClauses, "i.date <= $"+strconv.Itoa(argIndex))
			args = append(args, query.DateTo)
			argIndex++
		}

		if query.IncomeType != "" {
			whereClauses = append(whereClauses, "i.income_type = $"+strconv.Itoa(argIndex))
			args = append(args, query.IncomeType)
			argIndex++
		}

		if query.CategoryID != "" {
			whereClauses = append(whereClauses, "i.category_id = $"+strconv.Itoa(argIndex))
			args = append(args, query.CategoryID)
			argIndex++
		}

		if query.FamilyMemberID != "" {
			whereClauses = append(whereClauses, "i.family_member_id = $"+strconv.Itoa(argIndex))
			args = append(args, query.FamilyMemberID)
			argIndex++
		}

		whereClause := strings.Join(whereClauses, " AND ")

		// Get total count
		var totalCount int
		countQuery := "SELECT COUNT(*) FROM incomes i WHERE " + whereClause
		err := db.QueryRow(c.Request.Context(), countQuery, args...).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count incomes"})
			return
		}

		// Calculate pagination
		totalPages := (totalCount + query.Limit - 1) / query.Limit
		offset := (query.Page - 1) * query.Limit

		// Build main query with JOIN to get category name
		mainQuery := `
			SELECT i.id, i.family_member_id, i.category_id, ic.name as category_name,
			       i.description, i.amount, i.currency, i.exchange_rate, i.amount_in_primary_currency,
			       i.income_type, i.date, i.end_date, i.created_at
			FROM incomes i
			LEFT JOIN income_categories ic ON i.category_id = ic.id
			WHERE ` + whereClause + `
			ORDER BY i.` + query.SortBy + ` ` + strings.ToUpper(query.Order) + `
			LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

		args = append(args, query.Limit, offset)

		// Execute query
		rows, err := db.Query(c.Request.Context(), mainQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch incomes: " + err.Error()})
			return
		}
		defer rows.Close()

		// Parse results
		incomes := []IncomeListItem{}
		for rows.Next() {
			var income IncomeListItem
			var familyMemberID, categoryID, categoryName *string
			var date, endDate *time.Time
			var createdAt time.Time

			err := rows.Scan(
				&income.ID,
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
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse income: " + err.Error()})
				return
			}

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

			incomes = append(incomes, income)
		}

		// Check for errors during iteration
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading incomes"})
			return
		}

		// Build response
		response := ListIncomesResponse{
			Incomes:    incomes,
			TotalCount: totalCount,
			Page:       query.Page,
			Limit:      query.Limit,
			TotalPages: totalPages,
		}

		c.JSON(http.StatusOK, response)
	}
}
