package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/auth"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// LoginRequest representa el JSON que el cliente envía para hacer login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse representa el JSON que retornamos al cliente después del login
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

// UserInfo contiene información básica del usuario
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Login maneja el endpoint POST /api/auth/login
// Valida credenciales y retorna JWT tokens
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	// Validar el JSON recibido
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"details": err.Error(),
		})
		return
	}

	// Normalizar email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	ctx := c.Request.Context()

	// Buscar el usuario por email
	var userID, passwordHash, name string
	query := "SELECT id, password_hash, name FROM users WHERE email = $1"
	err := h.db.Pool.QueryRow(ctx, query, req.Email).Scan(&userID, &passwordHash, &name)

	if err != nil {
		// No revelar si el email existe o no (seguridad)
		logger.LogLoginFailed(req.Email, c.ClientIP(), "user_not_found")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Credenciales inválidas",
		})
		return
	}

	// Verificar la contraseña
	err = auth.CheckPassword(req.Password, passwordHash)
	if err != nil {
		logger.LogLoginFailed(req.Email, c.ClientIP(), "invalid_password")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Credenciales inválidas",
		})
		return
	}

	// Contraseña correcta - generar tokens
	// Parsear las duraciones de los tokens desde la config
	accessTokenExpiry, err := time.ParseDuration(h.config.JWTAccessExpiry)
	if err != nil {
		accessTokenExpiry = 15 * time.Minute // Fallback
	}

	refreshTokenExpiry, err := time.ParseDuration(h.config.JWTRefreshExpiry)
	if err != nil {
		refreshTokenExpiry = 7 * 24 * time.Hour // Fallback
	}

	// Obtener el JWT secret desde la config
	jwtSecret := h.config.JWTSecret

	// Generar access token
	accessToken, err := auth.GenerateAccessToken(userID, req.Email, jwtSecret, accessTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando token",
		})
		return
	}

	// Generar refresh token
	refreshToken, err := auth.GenerateRefreshToken(userID, jwtSecret, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando refresh token",
		})
		return
	}

	// Opcional: Guardar el refresh token en una cookie httpOnly (más seguro)
	// c.SetCookie(
	// 	"refresh_token",           // name
	// 	refreshToken,              // value
	// 	int(refreshTokenExpiry.Seconds()), // maxAge
	// 	"/",                       // path
	// 	"",                        // domain
	// 	false,                     // secure (true en producción con HTTPS)
	// 	true,                      // httpOnly
	// )

	// Log de login exitoso
	logger.LogLoginSuccess(userID, req.Email, c.ClientIP())

	// Retornar tokens y datos del usuario
	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserInfo{
			ID:    userID,
			Email: req.Email,
			Name:  name,
		},
	})
}
