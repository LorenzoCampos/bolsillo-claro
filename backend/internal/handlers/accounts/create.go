package accounts

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/LorenzoCampos/bolsillo-claro/internal/database"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// CreateAccountRequest representa el JSON para crear una cuenta
type CreateAccountRequest struct {
	Name     string        `json:"name" binding:"required"`
	Type     string        `json:"type" binding:"required,oneof=personal family"`
	Currency string        `json:"currency" binding:"required,oneof=ARS USD"`
	Members  []MemberInput `json:"members"`
}

// MemberInput representa un miembro familiar en la request
type MemberInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email"`
}

// AccountResponse representa la cuenta creada
type AccountResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	Currency  string           `json:"currency"`
	Members   []MemberResponse `json:"members,omitempty"`
	CreatedAt string           `json:"created_at"`
}

// MemberResponse representa un miembro en la response
type MemberResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// Handler encapsula las dependencias
type Handler struct {
	db *database.DB
}

// NewHandler crea una instancia del handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// CreateAccount maneja POST /api/accounts
// Crea una nueva cuenta y automáticamente crea la meta de Ahorro General
func (h *Handler) CreateAccount(c *gin.Context) {
	// Extraer user_id del contexto (viene del middleware de auth)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuario no autenticado",
		})
		return
	}

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"details": err.Error(),
		})
		return
	}

	// Validar que el nombre no esté vacío
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El nombre de la cuenta no puede estar vacío",
		})
		return
	}

	// Si es cuenta familiar, debe tener al menos un miembro
	if req.Type == "family" && len(req.Members) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Las cuentas familiares deben tener al menos un miembro",
		})
		return
	}

	// Si es cuenta personal, no debe tener miembros
	if req.Type == "personal" && len(req.Members) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Las cuentas personales no pueden tener miembros",
		})
		return
	}

	ctx := c.Request.Context()

	// Generar ID para la cuenta
	accountID := uuid.New()

	// Iniciar una transacción
	// Necesitamos transacción porque vamos a:
	// 1. Insertar la cuenta
	// 2. Insertar miembros (si es familiar)
	// 3. Insertar meta de Ahorro General
	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error iniciando transacción",
		})
		return
	}
	// defer garantiza que si algo falla, hacemos rollback
	defer tx.Rollback(ctx)

	// Insertar la cuenta
	insertAccountQuery := `
		INSERT INTO accounts (id, user_id, name, type, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at::TEXT
	`

	var createdAt string
	err = tx.QueryRow(
		ctx,
		insertAccountQuery,
		accountID,
		userID,
		req.Name,
		req.Type,
		req.Currency,
	).Scan(&createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error creando cuenta",
			"details": err.Error(),
		})
		return
	}

	// Insertar miembros si es cuenta familiar
	var members []MemberResponse
	if req.Type == "family" {
		for _, member := range req.Members {
			memberID := uuid.New()
			insertMemberQuery := `
				INSERT INTO family_members (id, account_id, name, email, is_active, created_at)
				VALUES ($1, $2, $3, $4, true, NOW())
			`

			_, err = tx.Exec(
				ctx,
				insertMemberQuery,
				memberID,
				accountID,
				strings.TrimSpace(member.Name),
				strings.TrimSpace(member.Email),
			)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Error creando miembro",
					"details": err.Error(),
				})
				return
			}

			members = append(members, MemberResponse{
				ID:    memberID.String(),
				Name:  member.Name,
				Email: member.Email,
			})
		}
	}

	// Crear la meta de Ahorro General automáticamente
	savingsGoalID := uuid.New()
	insertSavingsGoalQuery := `
		INSERT INTO savings_goals (
			id, account_id, name, target_amount, current_amount, 
			currency, deadline, is_general, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`

	// Meta general: target muy alto, sin deadline, is_general = true
	_, err = tx.Exec(
		ctx,
		insertSavingsGoalQuery,
		savingsGoalID,
		accountID,
		"Ahorro General",
		9999999999.99, // Target amount muy alto
		0,             // Current amount empieza en 0
		req.Currency,  // Misma moneda que la cuenta
		nil,           // Sin deadline
		true,          // is_general = true
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error creando meta de ahorro general",
			"details": err.Error(),
		})
		return
	}

	// Todo OK - hacer commit de la transacción
	err = tx.Commit(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error confirmando la transacción",
		})
		return
	}

	// Retornar la cuenta creada
	response := AccountResponse{
		ID:        accountID.String(),
		Name:      req.Name,
		Type:      req.Type,
		Currency:  req.Currency,
		CreatedAt: createdAt,
	}

	if req.Type == "family" {
		response.Members = members
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Cuenta creada exitosamente",
		"account": response,
	})
}
