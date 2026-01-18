package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// DeactivateMember maneja PATCH /api/accounts/:id/members/:member_id/deactivate
// Desactiva un miembro (soft delete)
func (h *Handler) DeactivateMember(c *gin.Context) {
	toggleMemberStatus(h, c, false)
}

// ReactivateMember maneja PATCH /api/accounts/:id/members/:member_id/reactivate
// Reactiva un miembro previamente desactivado
func (h *Handler) ReactivateMember(c *gin.Context) {
	toggleMemberStatus(h, c, true)
}

// toggleMemberStatus es una función helper que maneja tanto deactivate como reactivate
func toggleMemberStatus(h *Handler, c *gin.Context, newStatus bool) {
	// Extraer user_id del contexto
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuario no autenticado",
		})
		return
	}

	// Obtener IDs desde la URL
	accountID := c.Param("id")
	memberID := c.Param("member_id")

	if accountID == "" || memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID de cuenta y miembro requeridos",
		})
		return
	}

	ctx := c.Request.Context()

	// Verificar que la cuenta existe y pertenece al usuario
	var accountExists bool
	checkAccountQuery := `
		SELECT EXISTS(
			SELECT 1 
			FROM accounts 
			WHERE id = $1 AND user_id = $2
		)
	`
	err := h.db.Pool.QueryRow(ctx, checkAccountQuery, accountID, userID).Scan(&accountExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando cuenta",
			"details": err.Error(),
		})
		return
	}

	if !accountExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cuenta no encontrada o no pertenece al usuario",
		})
		return
	}

	// Verificar que el miembro existe y pertenece a esta cuenta
	var currentStatus bool
	checkMemberQuery := `
		SELECT is_active
		FROM family_members 
		WHERE id = $1 AND account_id = $2
	`
	err = h.db.Pool.QueryRow(ctx, checkMemberQuery, memberID, accountID).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Miembro no encontrado en esta cuenta",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando miembro",
			"details": err.Error(),
		})
		return
	}

	// Validar que el estado actual es diferente al nuevo estado
	if currentStatus == newStatus {
		action := "activo"
		if !newStatus {
			action = "inactivo"
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El miembro ya está " + action,
		})
		return
	}

	// Si estamos reactivando, verificar que no exista otro miembro activo con el mismo nombre
	if newStatus {
		var memberName string
		var duplicateExists bool
		
		// Obtener el nombre del miembro a reactivar
		getNameQuery := `SELECT name FROM family_members WHERE id = $1`
		err = h.db.Pool.QueryRow(ctx, getNameQuery, memberID).Scan(&memberName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error obteniendo nombre del miembro",
				"details": err.Error(),
			})
			return
		}

		// Verificar duplicados
		checkDuplicateQuery := `
			SELECT EXISTS(
				SELECT 1 
				FROM family_members 
				WHERE account_id = $1 
				  AND LOWER(name) = LOWER($2) 
				  AND is_active = true
				  AND id != $3
			)
		`
		err = h.db.Pool.QueryRow(ctx, checkDuplicateQuery, accountID, memberName, memberID).Scan(&duplicateExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error verificando duplicados",
				"details": err.Error(),
			})
			return
		}

		if duplicateExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Ya existe un miembro activo con ese nombre. Desactívelo primero o cambie el nombre de este miembro antes de reactivarlo.",
			})
			return
		}
	}

	// Actualizar el estado
	updateQuery := `
		UPDATE family_members 
		SET is_active = $1
		WHERE id = $2 AND account_id = $3
		RETURNING id, name, email, is_active
	`

	var member struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		IsActive bool   `json:"isActive"`
	}

	err = h.db.Pool.QueryRow(ctx, updateQuery, newStatus, memberID, accountID).Scan(
		&member.ID,
		&member.Name,
		&member.Email,
		&member.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Miembro no encontrado",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error actualizando estado del miembro",
			"details": err.Error(),
		})
		return
	}

	// Logging estructurado
	action := "deactivated"
	logMessage := "Miembro desactivado"
	if newStatus {
		action = "reactivated"
		logMessage = "Miembro reactivado"
	}
	logger.Info("member."+action, logMessage, map[string]interface{}{
		"member_id":  member.ID,
		"account_id": accountID,
		"user_id":    userID,
		"ip":         c.ClientIP(),
	})

	message := "Miembro desactivado exitosamente"
	if newStatus {
		message = "Miembro reactivado exitosamente"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"member":  member,
	})
}
