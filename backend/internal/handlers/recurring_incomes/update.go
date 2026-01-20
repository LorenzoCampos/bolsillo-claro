package recurring_incomes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// UpdateRecurringIncomeRequest representa el JSON para actualizar
// Todos los campos son opcionales (solo se actualiza lo que se envía)
type UpdateRecurringIncomeRequest struct {
	Description            *string  `json:"description"`
	Amount                 *float64 `json:"amount" binding:"omitempty,gt=0"`
	Currency               *string  `json:"currency" binding:"omitempty,oneof=ARS USD EUR"`
	CategoryID             *string  `json:"category_id"`
	FamilyMemberID         *string  `json:"family_member_id"`
	RecurrenceInterval     *int     `json:"recurrence_interval" binding:"omitempty,gt=0"`
	RecurrenceDayOfMonth   *int     `json:"recurrence_day_of_month" binding:"omitempty,gte=1,lte=31"`
	RecurrenceDayOfWeek    *int     `json:"recurrence_day_of_week" binding:"omitempty,gte=0,lte=6"`
	EndDate                *string  `json:"end_date"` // YYYY-MM-DD o null para eliminar
	TotalOccurrences       *int     `json:"total_occurrences" binding:"omitempty,gt=0"`
	IsActive               *bool    `json:"is_active"` // Para activar/desactivar
}

// UpdateRecurringIncome maneja PUT /api/recurring-expenses/:id
// IMPORTANTE: Actualizar el template NO afecta gastos ya generados (histórico preservado)
// Solo afecta FUTUROS gastos que se generen desde este template
func UpdateRecurringIncome(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recurringID := c.Param("id")

		var req UpdateRecurringIncomeRequest

		// Validar JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Datos inválidos",
				"details": err.Error(),
			})
			return
		}

		// Obtener account_id del contexto
		accountID, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Account-ID header requerido",
			})
			return
		}

		// Obtener user_id del contexto
		userID, _ := middleware.GetUserID(c)

		ctx := c.Request.Context()

		// Verificar que el recurring_expense existe y pertenece a esta cuenta
		var existsCheck bool
		var currentFrequency string
		checkQuery := `
			SELECT EXISTS(SELECT 1 FROM recurring_incomes WHERE id = $1 AND account_id = $2),
			       (SELECT recurrence_frequency FROM recurring_incomes WHERE id = $1)
		`
		err := pool.QueryRow(ctx, checkQuery, recurringID, accountID).Scan(&existsCheck, &currentFrequency)
		if err != nil || !existsCheck {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Ingreso recurrente no encontrado",
			})
			return
		}

		// Validar family_member_id si se está actualizando
		if req.FamilyMemberID != nil && *req.FamilyMemberID != "" {
			var memberExists bool
			checkMemberQuery := `
				SELECT EXISTS(
					SELECT 1 FROM family_members 
					WHERE id = $1 AND account_id = $2
				)
			`
			err := pool.QueryRow(ctx, checkMemberQuery, *req.FamilyMemberID, accountID).Scan(&memberExists)
			if err != nil || !memberExists {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "family_member_id no pertenece a esta cuenta",
				})
				return
			}
		}

		// Validar end_date formato si se está actualizando
		var endDate *time.Time
		if req.EndDate != nil && *req.EndDate != "" {
			parsed, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "end_date debe tener formato YYYY-MM-DD",
					"details": err.Error(),
				})
				return
			}
			endDate = &parsed
		}

		// Validación de negocio: si se actualiza day_of_month/day_of_week, verificar que concuerde con frequency
		if req.RecurrenceDayOfMonth != nil {
			if currentFrequency != "monthly" && currentFrequency != "yearly" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "recurrence_day_of_month solo aplica a frequency=monthly/yearly",
				})
				return
			}
		}

		if req.RecurrenceDayOfWeek != nil {
			if currentFrequency != "weekly" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "recurrence_day_of_week solo aplica a frequency=weekly",
				})
				return
			}
		}

		// Construir UPDATE dinámico (solo actualizar campos enviados)
		updateFields := []string{}
		args := []interface{}{}
		argCount := 1

		if req.Description != nil {
			updateFields = append(updateFields, "description = $"+itoa(argCount))
			args = append(args, *req.Description)
			argCount++
		}

		if req.Amount != nil {
			updateFields = append(updateFields, "amount = $"+itoa(argCount))
			args = append(args, *req.Amount)
			argCount++
		}

		if req.Currency != nil {
			updateFields = append(updateFields, "currency = $"+itoa(argCount))
			args = append(args, *req.Currency)
			argCount++
		}

		if req.CategoryID != nil {
			if *req.CategoryID == "" {
				updateFields = append(updateFields, "category_id = NULL")
			} else {
				updateFields = append(updateFields, "category_id = $"+itoa(argCount))
				args = append(args, *req.CategoryID)
				argCount++
			}
		}

		if req.FamilyMemberID != nil {
			if *req.FamilyMemberID == "" {
				updateFields = append(updateFields, "family_member_id = NULL")
			} else {
				updateFields = append(updateFields, "family_member_id = $"+itoa(argCount))
				args = append(args, *req.FamilyMemberID)
				argCount++
			}
		}

		if req.RecurrenceInterval != nil {
			updateFields = append(updateFields, "recurrence_interval = $"+itoa(argCount))
			args = append(args, *req.RecurrenceInterval)
			argCount++
		}

		if req.RecurrenceDayOfMonth != nil {
			updateFields = append(updateFields, "recurrence_day_of_month = $"+itoa(argCount))
			args = append(args, *req.RecurrenceDayOfMonth)
			argCount++
		}

		if req.RecurrenceDayOfWeek != nil {
			updateFields = append(updateFields, "recurrence_day_of_week = $"+itoa(argCount))
			args = append(args, *req.RecurrenceDayOfWeek)
			argCount++
		}

		if req.EndDate != nil {
			if *req.EndDate == "" {
				updateFields = append(updateFields, "end_date = NULL")
			} else {
				updateFields = append(updateFields, "end_date = $"+itoa(argCount))
				args = append(args, endDate)
				argCount++
			}
		}

		if req.TotalOccurrences != nil {
			updateFields = append(updateFields, "total_occurrences = $"+itoa(argCount))
			args = append(args, *req.TotalOccurrences)
			argCount++
		}

		if req.IsActive != nil {
			updateFields = append(updateFields, "is_active = $"+itoa(argCount))
			args = append(args, *req.IsActive)
			argCount++
		}

		// Si no hay campos para actualizar
		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No hay campos para actualizar",
			})
			return
		}

		// Agregar WHERE clause
		args = append(args, recurringID, accountID)
		whereClause := " WHERE id = $" + itoa(argCount) + " AND account_id = $" + itoa(argCount+1)

		// Construir query completo
		updateQuery := "UPDATE recurring_incomes SET " + join(updateFields, ", ") + whereClause + " RETURNING updated_at"

		var updatedAt time.Time
		err = pool.QueryRow(ctx, updateQuery, args...).Scan(&updatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error actualizando ingreso recurrente",
				"details": err.Error(),
			})
			return
		}

		// Log de actualización
		logger.Info("recurring_expense.updated", "Ingreso recurrente actualizado", map[string]interface{}{
			"recurring_income_id": recurringID,
			"account_id":           accountID,
			"user_id":              userID,
			"fields_updated":       len(updateFields),
			"ip":                   c.ClientIP(),
		})

		c.JSON(http.StatusOK, gin.H{
			"message":    "Ingreso recurrente actualizado exitosamente",
			"updated_at": updatedAt.Format(time.RFC3339),
			"note":       "Los gastos ya generados NO se modifican. Solo afecta futuros gastos.",
		})
	}
}

// Helper functions
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}
