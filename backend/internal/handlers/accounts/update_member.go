package accounts

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// UpdateMemberRequest representa la request para actualizar un miembro
type UpdateMemberRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}

// UpdateMemberResponse representa la response al actualizar un miembro
type UpdateMemberResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
}

// UpdateMember maneja PUT /api/accounts/:id/members/:member_id
// Actualiza nombre y/o email de un miembro existente
func (h *Handler) UpdateMember(c *gin.Context) {
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

	// Parse request body
	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"details": err.Error(),
		})
		return
	}

	// Validar que al menos un campo esté presente
	if req.Name == nil && req.Email == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Debe proporcionar al menos un campo para actualizar (name o email)",
		})
		return
	}

	// Trim y validar Name si está presente
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if len(trimmed) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "El nombre no puede estar vacío",
			})
			return
		}
		req.Name = &trimmed
	}

	// Trim Email si está presente
	if req.Email != nil {
		trimmed := strings.TrimSpace(*req.Email)
		req.Email = &trimmed
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
	var memberExists bool
	checkMemberQuery := `
		SELECT EXISTS(
			SELECT 1 
			FROM family_members 
			WHERE id = $1 AND account_id = $2
		)
	`
	err = h.db.Pool.QueryRow(ctx, checkMemberQuery, memberID, accountID).Scan(&memberExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando miembro",
			"details": err.Error(),
		})
		return
	}

	if !memberExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Miembro no encontrado en esta cuenta",
		})
		return
	}

	// Si se intenta cambiar el nombre, verificar que no exista otro miembro activo con ese nombre
	if req.Name != nil {
		var duplicateExists bool
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
		err = h.db.Pool.QueryRow(ctx, checkDuplicateQuery, accountID, *req.Name, memberID).Scan(&duplicateExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error verificando duplicados",
				"details": err.Error(),
			})
			return
		}

		if duplicateExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Ya existe otro miembro activo con ese nombre en esta cuenta",
			})
			return
		}
	}

	// Construir query dinámica de UPDATE
	query := "UPDATE family_members SET"
	args := []interface{}{}
	argPosition := 0
	needsComma := false

	if req.Name != nil {
		if needsComma {
			query += ","
		}
		argPosition++
		query += " name = $" + string(rune(argPosition+'0'))
		args = append(args, *req.Name)
		needsComma = true
	}

	if req.Email != nil {
		if needsComma {
			query += ","
		}
		argPosition++
		query += " email = $" + string(rune(argPosition+'0'))
		args = append(args, *req.Email)
		needsComma = true
	}

	// Agregar condiciones WHERE
	argPosition++
	query += " WHERE id = $" + string(rune(argPosition+'0'))
	args = append(args, memberID)

	argPosition++
	query += " AND account_id = $" + string(rune(argPosition+'0'))
	args = append(args, accountID)

	query += " RETURNING id, name, email, is_active"

	// Ejecutar UPDATE
	var member UpdateMemberResponse
	err = h.db.Pool.QueryRow(ctx, query, args...).Scan(
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
			"error":   "Error actualizando miembro",
			"details": err.Error(),
		})
		return
	}

	// Logging estructurado
	logger.Info("member.updated", "Miembro actualizado", map[string]interface{}{
		"member_id":  member.ID,
		"account_id": accountID,
		"user_id":    userID,
		"ip":         c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Miembro actualizado exitosamente",
		"member":  member,
	})
}
