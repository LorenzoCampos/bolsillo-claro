package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/auth"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// RefreshRequest representa el JSON que el cliente envía para renovar el token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse representa el JSON que retornamos con los nuevos tokens
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Refresh maneja el endpoint POST /api/auth/refresh
// Valida un refresh token y genera un nuevo par de tokens (access + refresh)
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest

	// Validar el JSON recibido
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"details": err.Error(),
		})
		return
	}

	// Validar el refresh token
	claims, err := auth.ValidateToken(req.RefreshToken, h.config.JWTSecret)
	if err != nil {
		logger.LogRefreshFailed(c.ClientIP(), "invalid_token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Refresh token inválido o expirado",
		})
		return
	}

	// Verificar que el usuario aún existe en la base de datos
	// (podría haber sido eliminado después de generar el refresh token)
	ctx := c.Request.Context()
	var email, name string
	query := "SELECT email, name FROM users WHERE id = $1"
	err = h.db.Pool.QueryRow(ctx, query, claims.UserID).Scan(&email, &name)

	if err != nil {
		logger.LogRefreshFailed(c.ClientIP(), "user_not_found")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuario no encontrado",
		})
		return
	}

	// Usuario válido - generar nuevos tokens
	accessTokenExpiry, err := time.ParseDuration(h.config.JWTAccessExpiry)
	if err != nil {
		accessTokenExpiry = 15 * time.Minute // Fallback
	}

	refreshTokenExpiry, err := time.ParseDuration(h.config.JWTRefreshExpiry)
	if err != nil {
		refreshTokenExpiry = 7 * 24 * time.Hour // Fallback
	}

	// Generar nuevo access token
	newAccessToken, err := auth.GenerateAccessToken(claims.UserID, email, h.config.JWTSecret, accessTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando access token",
		})
		return
	}

	// Generar nuevo refresh token (rotación de tokens)
	// Best practice: siempre rotar el refresh token para prevenir reuso
	newRefreshToken, err := auth.GenerateRefreshToken(claims.UserID, h.config.JWTSecret, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando refresh token",
		})
		return
	}

	// Log de refresh exitoso
	logger.LogRefreshSuccess(claims.UserID, email, c.ClientIP())

	// Retornar los nuevos tokens
	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

// RefreshFromHeader maneja el refresh desde el header Authorization
// Alternativa: POST /api/auth/refresh sin body, lee el refresh token del header
// Útil si el frontend guarda el refresh token en localStorage/cookies
func (h *Handler) RefreshFromHeader(c *gin.Context) {
	// Extraer el token del header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Header Authorization requerido",
		})
		return
	}

	// El formato esperado es: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Formato de Authorization inválido. Use: Bearer <token>",
		})
		return
	}

	refreshToken := parts[1]

	// Validar el refresh token
	claims, err := auth.ValidateToken(refreshToken, h.config.JWTSecret)
	if err != nil {
		logger.LogRefreshFailed(c.ClientIP(), "invalid_token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Refresh token inválido o expirado",
		})
		return
	}

	// Verificar que el usuario aún existe
	ctx := c.Request.Context()
	var email, name string
	query := "SELECT email, name FROM users WHERE id = $1"
	err = h.db.Pool.QueryRow(ctx, query, claims.UserID).Scan(&email, &name)

	if err != nil {
		logger.LogRefreshFailed(c.ClientIP(), "user_not_found")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuario no encontrado",
		})
		return
	}

	// Generar nuevos tokens
	accessTokenExpiry, err := time.ParseDuration(h.config.JWTAccessExpiry)
	if err != nil {
		accessTokenExpiry = 15 * time.Minute
	}

	refreshTokenExpiry, err := time.ParseDuration(h.config.JWTRefreshExpiry)
	if err != nil {
		refreshTokenExpiry = 7 * 24 * time.Hour
	}

	newAccessToken, err := auth.GenerateAccessToken(claims.UserID, email, h.config.JWTSecret, accessTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando access token",
		})
		return
	}

	newRefreshToken, err := auth.GenerateRefreshToken(claims.UserID, h.config.JWTSecret, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generando refresh token",
		})
		return
	}

	logger.LogRefreshSuccess(claims.UserID, email, c.ClientIP())

	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}
