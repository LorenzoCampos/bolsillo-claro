package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword toma una contraseña en texto plano y retorna su hash bcrypt
// El hash resultante puede almacenarse de forma segura en la base de datos
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword genera un hash salted de la contraseña
	// Cost factor 12 es un buen balance entre seguridad y performance
	// A mayor cost, más seguro pero más lento (exponencial)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("error hasheando password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compara una contraseña en texto plano con un hash bcrypt
// Retorna nil si coinciden, o error si no coinciden
func CheckPassword(password, hash string) error {
	// bcrypt.CompareHashAndPassword compara de forma segura
	// Internamente rehashea la password con el mismo salt y compara
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("password incorrecta")
	}
	return nil
}
