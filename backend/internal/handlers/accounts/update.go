package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// UpdateAccountRequest representa la estructura de datos para actualizar una cuenta
type UpdateAccountRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Currency *string `json:"currency,omitempty" binding:"omitempty,oneof=ARS USD EUR"`
}

// UpdateAccount maneja PUT /api/accounts/:id
// Permite actualizar el nombre y/o la moneda de una cuenta
// NOTA: No se permite cambiar el tipo de cuenta (personal/family) una vez creada
func (h *Handler) UpdateAccount(c *gin.Context) {
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

	// Parsear el body
	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"details": err.Error(),
		})
		return
	}

	// Validar que al menos un campo esté presente
	if req.Name == nil && req.Currency == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Debe proporcionar al menos un campo para actualizar (name o currency)",
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

	// Construir la query de actualización dinámicamente
	// Solo actualizamos los campos que vienen en el request
	query := `UPDATE accounts SET `
	args := []interface{}{}
	argPos := 1

	if req.Name != nil {
		query += `name = $` + string(rune(argPos+'0')) + `, `
		args = append(args, *req.Name)
		argPos++
	}

	if req.Currency != nil {
		query += `currency = $` + string(rune(argPos+'0')) + `, `
		args = append(args, *req.Currency)
		argPos++
	}

	// Remover la última coma y espacio
	query = query[:len(query)-2]

	// Agregar el WHERE y updated_at
	query += `, updated_at = NOW() WHERE id = $` + string(rune(argPos+'0')) + ` AND user_id = $` + string(rune(argPos+1+'0'))
	args = append(args, accountID, userID)

	// Ejecutar la actualización
	cmdTag, err := h.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error actualizando cuenta",
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

	// Obtener la cuenta actualizada
	getQuery := `
		SELECT 
			id,
			name,
			type,
			currency,
			created_at::TEXT,
			updated_at::TEXT
		FROM accounts
		WHERE id = $1
	`

	var account struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		Currency  string `json:"currency"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}

	err = h.db.Pool.QueryRow(ctx, getQuery, accountID).Scan(
		&account.ID,
		&account.Name,
		&account.Type,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cuenta no encontrada después de actualizar",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error obteniendo cuenta actualizada",
			"details": err.Error(),
		})
		return
	}

	// Retornar la cuenta actualizada
	c.JSON(http.StatusOK, gin.H{
		"message": "Cuenta actualizada exitosamente",
		"account": account,
	})
}
