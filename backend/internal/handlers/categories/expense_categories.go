package categories

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseCategoryResponse struct {
	ID        string  `json:"id"`
	AccountID *string `json:"account_id,omitempty"` // NULL si es system
	Name      string  `json:"name"`
	Icon      *string `json:"icon,omitempty"`
	Color     *string `json:"color,omitempty"`
	IsSystem  bool    `json:"is_system"`
	CreatedAt string  `json:"created_at"`
}

type CreateExpenseCategoryRequest struct {
	Name  string  `json:"name" binding:"required"`
	Icon  *string `json:"icon"`
	Color *string `json:"color"`
}

type UpdateExpenseCategoryRequest struct {
	Name  *string `json:"name"`
	Icon  *string `json:"icon"`
	Color *string `json:"color"`
}

// ListExpenseCategories devuelve todas las categorías:
// - Categorías del sistema (account_id = NULL, is_system = TRUE)
// - Categorías custom de esta cuenta (account_id = X-Account-ID)
func ListExpenseCategories(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context (set by AccountMiddleware)
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		query := `
			SELECT id, account_id, name, icon, color, is_system, created_at
			FROM expense_categories
			WHERE account_id IS NULL OR account_id = $1
			ORDER BY is_system DESC, name ASC
		`

		rows, err := db.Query(c.Request.Context(), query, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories: " + err.Error()})
			return
		}
		defer rows.Close()

		categories := []ExpenseCategoryResponse{}
		for rows.Next() {
			var cat ExpenseCategoryResponse
			var accountIDPtr *string
			var createdAt time.Time

			err := rows.Scan(
				&cat.ID,
				&accountIDPtr,
				&cat.Name,
				&cat.Icon,
				&cat.Color,
				&cat.IsSystem,
				&createdAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse category: " + err.Error()})
				return
			}

			cat.AccountID = accountIDPtr
			cat.CreatedAt = createdAt.Format(time.RFC3339)
			categories = append(categories, cat)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading categories"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
			"count":      len(categories),
		})
	}
}

// CreateExpenseCategory crea una categoría custom para esta cuenta
func CreateExpenseCategory(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		var req CreateExpenseCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Insert custom category
		var cat ExpenseCategoryResponse
		var accountIDPtr *string
		var createdAt time.Time

		query := `
			INSERT INTO expense_categories (account_id, name, icon, color, is_system)
			VALUES ($1, $2, $3, $4, FALSE)
			RETURNING id, account_id, name, icon, color, is_system, created_at
		`

		err := db.QueryRow(c.Request.Context(), query, accountID, req.Name, req.Icon, req.Color).Scan(
			&cat.ID,
			&accountIDPtr,
			&cat.Name,
			&cat.Icon,
			&cat.Color,
			&cat.IsSystem,
			&createdAt,
		)

		if err != nil {
			// Check for unique constraint violation
			if err.Error() == "ERROR: duplicate key value violates unique constraint \"unique_expense_category_custom\" (SQLSTATE 23505)" {
				c.JSON(http.StatusConflict, gin.H{"error": "category with this name already exists in this account"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category: " + err.Error()})
			return
		}

		cat.AccountID = accountIDPtr
		cat.CreatedAt = createdAt.Format(time.RFC3339)

		c.JSON(http.StatusCreated, cat)
	}
}

// UpdateExpenseCategory actualiza una categoría custom (solo si no es is_system)
func UpdateExpenseCategory(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		categoryID := c.Param("id")
		if categoryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category_id is required"})
			return
		}

		var req UpdateExpenseCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if category exists, belongs to this account, and is not system
		var isSystem bool
		var categoryAccountID *string
		checkQuery := `SELECT is_system, account_id FROM expense_categories WHERE id = $1`
		err := db.QueryRow(c.Request.Context(), checkQuery, categoryID).Scan(&isSystem, &categoryAccountID)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}

		if isSystem {
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot edit system categories"})
			return
		}

		if categoryAccountID == nil || *categoryAccountID != accountID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "category does not belong to this account"})
			return
		}

		// Update category
		updateQuery := `
			UPDATE expense_categories SET
				name = COALESCE($1, name),
				icon = COALESCE($2, icon),
				color = COALESCE($3, color),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
			RETURNING id, account_id, name, icon, color, is_system, created_at
		`

		var cat ExpenseCategoryResponse
		var accountIDPtr *string
		var createdAt time.Time

		err = db.QueryRow(c.Request.Context(), updateQuery, req.Name, req.Icon, req.Color, categoryID).Scan(
			&cat.ID,
			&accountIDPtr,
			&cat.Name,
			&cat.Icon,
			&cat.Color,
			&cat.IsSystem,
			&createdAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update category: " + err.Error()})
			return
		}

		cat.AccountID = accountIDPtr
		cat.CreatedAt = createdAt.Format(time.RFC3339)

		c.JSON(http.StatusOK, cat)
	}
}

// DeleteExpenseCategory elimina una categoría custom
// Solo permite borrar si:
// 1. No es is_system
// 2. Pertenece a esta cuenta
// 3. No tiene expenses asociados
func DeleteExpenseCategory(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get account_id from context
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account_id not found in context"})
			return
		}

		categoryID := c.Param("id")
		if categoryID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category_id is required"})
			return
		}

		// Check if category exists, belongs to this account, and is not system
		var isSystem bool
		var categoryAccountID *string
		checkQuery := `SELECT is_system, account_id FROM expense_categories WHERE id = $1`
		err := db.QueryRow(c.Request.Context(), checkQuery, categoryID).Scan(&isSystem, &categoryAccountID)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}

		if isSystem {
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete system categories"})
			return
		}

		if categoryAccountID == nil || *categoryAccountID != accountID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "category does not belong to this account"})
			return
		}

		// Check if category has associated expenses
		var expenseCount int
		countQuery := `SELECT COUNT(*) FROM expenses WHERE category_id = $1`
		err = db.QueryRow(c.Request.Context(), countQuery, categoryID).Scan(&expenseCount)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check category usage"})
			return
		}

		if expenseCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":         "cannot delete category with associated expenses",
				"expense_count": expenseCount,
			})
			return
		}

		// Delete category
		deleteQuery := `DELETE FROM expense_categories WHERE id = $1`
		_, err = db.Exec(c.Request.Context(), deleteQuery, categoryID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "category deleted successfully",
			"id":      categoryID,
		})
	}
}
