package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims representa los datos que guardamos dentro del JWT
// jwt.RegisteredClaims incluye campos estándar como ExpiresAt, IssuedAt, etc.
type Claims struct {
	UserID string `json:"user_id"` // ID del usuario autenticado
	Email  string `json:"email"`   // Email del usuario (útil para debugging)
	jwt.RegisteredClaims
}

// GenerateAccessToken genera un JWT de corta duración (access token)
// Este token se usa en cada petición HTTP para autenticar al usuario
func GenerateAccessToken(userID, email, secret string, expiry time.Duration) (string, error) {
	// Crear los claims (datos del token)
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "bolsillo-claro",
		},
	}

	// Crear el token con el algoritmo HS256 (HMAC-SHA256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token con la clave secreta
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error firmando token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken genera un JWT de larga duración (refresh token)
// Este token se usa para obtener nuevos access tokens cuando expiran
func GenerateRefreshToken(userID, secret string, expiry time.Duration) (string, error) {
	// El refresh token solo necesita el userID
	// No incluimos datos adicionales para reducir tamaño
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "bolsillo-claro",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error firmando refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken valida un JWT y retorna los claims si es válido
// Verifica la firma, la expiración, y que sea del issuer correcto
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// Parsear el token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea el esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parseando token: %w", err)
	}

	// Extraer los claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}
