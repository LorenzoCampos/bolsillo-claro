package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// DeleteAccount maneja DELETE /api/accounts/:id
// Elimina una cuenta solo si no tiene datos asociados (expenses, incomes, savings_goals)
// Si tiene datos, retorna un error 409 (Conflict)
func (h *Handler) DeleteAccount(c *gin.Context) {
	// Extraer user_id del contexto (viene del middleware de auth)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuario no autenticado",
		})
		return
	}

	// Obtener el ID de la cuenta desde la URL
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID de cuenta requerido",
		})
		return
	}

	ctx := c.Request.Context()

	// Verificar que la cuenta existe y pertenece al usuario
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2)`
	err := h.db.Pool.QueryRow(ctx, checkQuery, accountID, userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando cuenta",
			"details": err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cuenta no encontrada o no pertenece al usuario",
		})
		return
	}

	// Verificar que no tenga datos asociados
	// Checkeamos: expenses, incomes, savings_goals
	var hasExpenses, hasIncomes, hasGoals bool

	// Check expenses
	err = h.db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM expenses WHERE account_id = $1 LIMIT 1)`,
		accountID,
	).Scan(&hasExpenses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando gastos",
			"details": err.Error(),
		})
		return
	}

	// Check incomes
	err = h.db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM incomes WHERE account_id = $1 LIMIT 1)`,
		accountID,
	).Scan(&hasIncomes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando ingresos",
			"details": err.Error(),
		})
		return
	}

	// Check savings_goals
	err = h.db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM savings_goals WHERE account_id = $1 LIMIT 1)`,
		accountID,
	).Scan(&hasGoals)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando metas de ahorro",
			"details": err.Error(),
		})
		return
	}

	// Si tiene datos asociados, no permitir eliminación
	if hasExpenses || hasIncomes || hasGoals {
		conflicts := []string{}
		if hasExpenses {
			conflicts = append(conflicts, "gastos")
		}
		if hasIncomes {
			conflicts = append(conflicts, "ingresos")
		}
		if hasGoals {
			conflicts = append(conflicts, "metas de ahorro")
		}

		c.JSON(http.StatusConflict, gin.H{
			"error":      "No se puede eliminar la cuenta porque tiene datos asociados",
			"conflicts":  conflicts,
			"suggestion": "Elimine primero todos los gastos, ingresos y metas de ahorro antes de eliminar la cuenta",
		})
		return
	}

	// Iniciar transacción para eliminar la cuenta y sus family_members
	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error iniciando transacción",
			"details": err.Error(),
		})
		return
	}
	defer tx.Rollback(ctx) // Rollback automático si no se hace Commit

	// Eliminar family_members si existen (si es cuenta family)
	_, err = tx.Exec(ctx, `DELETE FROM family_members WHERE account_id = $1`, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error eliminando miembros de la familia",
			"details": err.Error(),
		})
		return
	}

	// Eliminar custom categories asociadas a esta cuenta
	// Las categorías del sistema (is_system = true) NO se eliminan
	_, err = tx.Exec(ctx, `DELETE FROM expense_categories WHERE account_id = $1 AND is_system = false`, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error eliminando categorías de gastos personalizadas",
			"details": err.Error(),
		})
		return
	}

	_, err = tx.Exec(ctx, `DELETE FROM income_categories WHERE account_id = $1 AND is_system = false`, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error eliminando categorías de ingresos personalizadas",
			"details": err.Error(),
		})
		return
	}

	// Eliminar la cuenta
	cmdTag, err := tx.Exec(ctx, `DELETE FROM accounts WHERE id = $1 AND user_id = $2`, accountID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error eliminando cuenta",
			"details": err.Error(),
		})
		return
	}

	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cuenta no encontrada",
		})
		return
	}

	// Commit de la transacción
	err = tx.Commit(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error confirmando eliminación",
			"details": err.Error(),
		})
		return
	}

	// Retornar éxito
	c.JSON(http.StatusOK, gin.H{
		"message":   "Cuenta eliminada exitosamente",
		"accountId": accountID,
	})
}
