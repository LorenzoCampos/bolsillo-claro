package accounts

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// AddMemberRequest representa la request para agregar un miembro
type AddMemberRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email"`
}

// AddMemberResponse representa la response al agregar un miembro
type AddMemberResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
}

// AddMember maneja POST /api/accounts/:id/members
// Agrega un nuevo miembro a una cuenta familiar
func (h *Handler) AddMember(c *gin.Context) {
	// Extraer user_id del contexto
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

	// Parse request body
	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"details": err.Error(),
		})
		return
	}

	// Validar que name no esté vacío después de trim
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El nombre del miembro no puede estar vacío",
		})
		return
	}

	// Trim email
	req.Email = strings.TrimSpace(req.Email)

	ctx := c.Request.Context()

	// Verificar que la cuenta existe, pertenece al usuario y es de tipo family
	var accountType string
	checkAccountQuery := `
		SELECT type 
		FROM accounts 
		WHERE id = $1 AND user_id = $2
	`
	err := h.db.Pool.QueryRow(ctx, checkAccountQuery, accountID, userID).Scan(&accountType)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cuenta no encontrada o no pertenece al usuario",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando cuenta",
			"details": err.Error(),
		})
		return
	}

	// Verificar que sea cuenta family
	if accountType != "family" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Solo se pueden agregar miembros a cuentas de tipo 'family'",
		})
		return
	}

	// Verificar que no exista un miembro activo con el mismo nombre en esta cuenta
	var exists bool
	checkDuplicateQuery := `
		SELECT EXISTS(
			SELECT 1 
			FROM family_members 
			WHERE account_id = $1 
			  AND LOWER(name) = LOWER($2) 
			  AND is_active = true
		)
	`
	err = h.db.Pool.QueryRow(ctx, checkDuplicateQuery, accountID, req.Name).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error verificando duplicados",
			"details": err.Error(),
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Ya existe un miembro activo con ese nombre en esta cuenta",
		})
		return
	}

	// Insertar nuevo miembro
	memberID := uuid.New()
	insertMemberQuery := `
		INSERT INTO family_members (id, account_id, name, email, is_active, created_at)
		VALUES ($1, $2, $3, $4, true, NOW())
		RETURNING id, name, email, is_active
	`

	var member AddMemberResponse
	err = h.db.Pool.QueryRow(
		ctx,
		insertMemberQuery,
		memberID,
		accountID,
		req.Name,
		req.Email,
	).Scan(&member.ID, &member.Name, &member.Email, &member.IsActive)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error creando miembro",
			"details": err.Error(),
		})
		return
	}

	// Logging estructurado
	logger.Info("member.added", "Miembro agregado a cuenta familiar", map[string]interface{}{
		"member_id":  member.ID,
		"account_id": accountID,
		"user_id":    userID,
		"name":       member.Name,
		"ip":         c.ClientIP(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Miembro agregado exitosamente",
		"member":  member,
	})
}
