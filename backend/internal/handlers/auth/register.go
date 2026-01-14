package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/LorenzoCampos/bolsillo-claro/internal/config"
	"github.com/LorenzoCampos/bolsillo-claro/internal/database"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/auth"
)

// RegisterRequest representa el JSON que el cliente envía para registrarse
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// RegisterResponse representa el JSON que retornamos al cliente
type RegisterResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// Handler encapsula las dependencias necesarias para los handlers de auth
type Handler struct {
	db     *database.DB
	config *config.Config
}

// NewHandler crea una nueva instancia del handler de auth
func NewHandler(db *database.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:     db,
		config: cfg,
	}
}

// Register maneja el endpoint POST /api/auth/register
// Crea un nuevo usuario en la base de datos
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	// Validar el JSON recibido usando las reglas de binding
	// Si falla, Gin automáticamente retorna 400 Bad Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"details": err.Error(),
		})
		return
	}

	// Normalizar email a minúsculas
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	ctx := c.Request.Context()

	// Verificar si el email ya existe
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := h.db.Pool.QueryRow(ctx, checkQuery, req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error verificando email",
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "El email ya está registrado",
		})
		return
	}

	// Hashear la contraseña
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error procesando la contraseña",
		})
		return
	}

	// Generar UUID para el nuevo usuario
	userID := uuid.New()

	// Insertar el usuario en la base de datos
	insertQuery := `
		INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at::TEXT
	`

	var createdAt string
	err = h.db.Pool.QueryRow(
		ctx,
		insertQuery,
		userID,
		req.Email,
		passwordHash,
		req.Name,
	).Scan(&createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error creando usuario",
			"details": err.Error(),
		})
		return
	}

	// Retornar el usuario creado (sin el password hash!)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Usuario creado exitosamente",
		"user": RegisterResponse{
			ID:        userID.String(),
			Email:     req.Email,
			Name:      req.Name,
			CreatedAt: createdAt,
		},
	})
}
