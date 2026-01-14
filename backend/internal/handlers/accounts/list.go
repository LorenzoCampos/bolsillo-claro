package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// AccountListItem representa una cuenta en la lista
type AccountListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Currency    string `json:"currency"`
	MemberCount *int   `json:"memberCount,omitempty"` // Solo para cuentas family
	CreatedAt   string `json:"createdAt"`
}

// ListAccounts maneja GET /api/accounts
// Retorna todas las cuentas del usuario autenticado
func (h *Handler) ListAccounts(c *gin.Context) {
	// Extraer user_id del contexto (viene del middleware de auth)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuario no autenticado",
		})
		return
	}

	ctx := c.Request.Context()

	// Query para obtener todas las cuentas del usuario
	// Incluimos un LEFT JOIN con family_members para contar miembros
	query := `
		SELECT 
			a.id,
			a.name,
			a.type,
			a.currency,
			a.created_at::TEXT,
			COUNT(fm.id) FILTER (WHERE a.type = 'family') as member_count
		FROM accounts a
		LEFT JOIN family_members fm ON a.id = fm.account_id AND fm.is_active = true
		WHERE a.user_id = $1
		GROUP BY a.id, a.name, a.type, a.currency, a.created_at
		ORDER BY a.created_at DESC
	`

	rows, err := h.db.Pool.Query(ctx, query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error obteniendo cuentas",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	// Leer las cuentas
	accounts := []AccountListItem{}
	for rows.Next() {
		var account AccountListItem
		var memberCount int

		err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Type,
			&account.Currency,
			&account.CreatedAt,
			&memberCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error leyendo cuentas",
				"details": err.Error(),
			})
			return
		}

		// Solo incluir memberCount si es cuenta familiar
		if account.Type == "family" {
			account.MemberCount = &memberCount
		}

		accounts = append(accounts, account)
	}

	// Verificar errores en la iteración
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error procesando cuentas",
			"details": err.Error(),
		})
		return
	}

	// Retornar las cuentas
	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
		"count":    len(accounts),
	})
}
