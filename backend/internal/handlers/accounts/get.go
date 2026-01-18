package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// FamilyMemberDetail representa un miembro de la familia
type FamilyMemberDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
}

// AccountDetail representa el detalle completo de una cuenta
type AccountDetail struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Type      string               `json:"type"`
	Currency  string               `json:"currency"`
	CreatedAt string               `json:"createdAt"`
	Members   []FamilyMemberDetail `json:"members,omitempty"` // Solo para cuentas family
}

// GetAccount maneja GET /api/accounts/:id
// Retorna el detalle de una cuenta específica con sus miembros si es family
func (h *Handler) GetAccount(c *gin.Context) {
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

	// Query para obtener la cuenta
	// IMPORTANTE: Verificamos que pertenezca al usuario autenticado
	query := `
		SELECT 
			id,
			name,
			type,
			currency,
			created_at::TEXT
		FROM accounts
		WHERE id = $1 AND user_id = $2
	`

	var account AccountDetail
	err := h.db.Pool.QueryRow(ctx, query, accountID, userID).Scan(
		&account.ID,
		&account.Name,
		&account.Type,
		&account.Currency,
		&account.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cuenta no encontrada o no pertenece al usuario",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error obteniendo cuenta",
			"details": err.Error(),
		})
		return
	}

	// Si es cuenta family, obtener los miembros
	if account.Type == "family" {
		membersQuery := `
			SELECT 
				id,
				name,
				email,
				is_active
			FROM family_members
			WHERE account_id = $1 AND is_active = true
			ORDER BY created_at ASC
		`

		rows, err := h.db.Pool.Query(ctx, membersQuery, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error obteniendo miembros de la familia",
				"details": err.Error(),
			})
			return
		}
		defer rows.Close()

		members := []FamilyMemberDetail{}
		for rows.Next() {
			var member FamilyMemberDetail
			err := rows.Scan(&member.ID, &member.Name, &member.Email, &member.IsActive)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Error leyendo miembros",
					"details": err.Error(),
				})
				return
			}
			members = append(members, member)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error procesando miembros",
				"details": err.Error(),
			})
			return
		}

		account.Members = members
	}

	// Retornar el detalle de la cuenta
	c.JSON(http.StatusOK, account)
}
