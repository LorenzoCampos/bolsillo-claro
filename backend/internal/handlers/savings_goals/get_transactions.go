package savings_goals

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// GetTransactionsResponse represents the response for transactions endpoint
type GetTransactionsResponse struct {
	Transactions []SavingsGoalTransaction `json:"transactions"`
	Pagination   PaginationMetadata       `json:"pagination"`
}

// GetTransactions handles GET /api/savings-goals/:id/transactions
func GetTransactions(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context
		accountID, exists := middleware.GetAccountID(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		// Get savings goal ID from URL
		goalID := c.Param("id")
		if goalID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "savings_goal_id is required"})
			return
		}

		ctx := c.Request.Context()

		// Verify goal exists and belongs to this account
		checkQuery := `SELECT id FROM savings_goals WHERE id = $1 AND account_id = $2`
		var goalExists string
		err := db.QueryRow(ctx, checkQuery, goalID, accountID).Scan(&goalExists)

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "meta de ahorro no encontrada o no pertenece a esta cuenta"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify savings goal"})
			return
		}

		// Parse pagination parameters
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")
		transactionType := c.DefaultQuery("type", "all")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 20
		}
		// Max limit is 100 to prevent huge responses
		if limit > 100 {
			limit = 100
		}

		// Validate transaction type
		if transactionType != "all" && transactionType != "deposit" && transactionType != "withdrawal" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'all', 'deposit', or 'withdrawal'"})
			return
		}

		offset := (page - 1) * limit

		// Build count query with type filter
		countQuery := `SELECT COUNT(*) FROM savings_goal_transactions WHERE savings_goal_id = $1`
		countArgs := []interface{}{goalID}

		if transactionType != "all" {
			countQuery += ` AND transaction_type = $2`
			countArgs = append(countArgs, transactionType)
		}

		// Count total transactions
		var totalCount int
		err = db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count transactions"})
			return
		}

		// Calculate total pages
		totalPages := (totalCount + limit - 1) / limit
		if totalPages < 1 {
			totalPages = 1
		}

		// Build transactions query with type filter
		transactionsQuery := `
			SELECT 
				id, amount, transaction_type, description, 
				date::TEXT, created_at::TEXT
			FROM savings_goal_transactions
			WHERE savings_goal_id = $1`

		queryArgs := []interface{}{goalID}

		if transactionType != "all" {
			transactionsQuery += ` AND transaction_type = $2`
			queryArgs = append(queryArgs, transactionType)
		}

		transactionsQuery += `
			ORDER BY date DESC, created_at DESC
			LIMIT $` + strconv.Itoa(len(queryArgs)+1) + ` OFFSET $` + strconv.Itoa(len(queryArgs)+2)

		queryArgs = append(queryArgs, limit, offset)

		// Query transactions
		rows, err := db.Query(ctx, transactionsQuery, queryArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
			return
		}
		defer rows.Close()

		transactions := []SavingsGoalTransaction{}
		for rows.Next() {
			var txn SavingsGoalTransaction
			var description *string

			err := rows.Scan(
				&txn.ID, &txn.Amount, &txn.TransactionType,
				&description, &txn.Date, &txn.CreatedAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse transaction"})
				return
			}

			txn.Description = description

			// For display purposes, show withdrawals as negative amounts
			if txn.TransactionType == "withdrawal" {
				txn.Amount = -txn.Amount
			}

			transactions = append(transactions, txn)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading transactions"})
			return
		}

		// Build pagination metadata
		pagination := PaginationMetadata{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  totalCount,
			Limit:       limit,
		}

		// Build response
		response := GetTransactionsResponse{
			Transactions: transactions,
			Pagination:   pagination,
		}

		c.JSON(http.StatusOK, response)
	}
}
