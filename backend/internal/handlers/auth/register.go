package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/LorenzoCampos/bolsillo-claro/internal/config"
	"github.com/LorenzoCampos/bolsillo-claro/internal/database"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/auth"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
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
			"details": err.Error(),
		})
		return
	}

	if exists {
		logger.LogRegisterFailed(req.Email, c.ClientIP(), "email_already_exists")
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

	// Generar tokens JWT para auto-login después del registro
	accessTokenExpiry, err := time.ParseDuration(h.config.JWTAccessExpiry)
	if err != nil {
		accessTokenExpiry = 15 * time.Minute // Fallback
	}

	refreshTokenExpiry, err := time.ParseDuration(h.config.JWTRefreshExpiry)
	if err != nil {
		refreshTokenExpiry = 7 * 24 * time.Hour // Fallback
	}

	jwtSecret := h.config.JWTSecret

	// Generar access token
	accessToken, err := auth.GenerateAccessToken(userID.String(), req.Email, jwtSecret, accessTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando token",
		})
		return
	}

	// Generar refresh token
	refreshToken, err := auth.GenerateRefreshToken(userID.String(), jwtSecret, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando refresh token",
		})
		return
	}

	// Log de registro exitoso
	logger.LogRegisterSuccess(userID.String(), req.Email, c.ClientIP())

	// Retornar el usuario creado CON tokens (auto-login)
	c.JSON(http.StatusCreated, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserInfo{
			ID:    userID.String(),
			Email: req.Email,
			Name:  req.Name,
		},
	})
}
