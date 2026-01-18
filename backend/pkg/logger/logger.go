package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel representa el nivel de severidad del log
type LogLevel string

const (
	LevelInfo     LogLevel = "INFO"
	LevelWarning  LogLevel = "WARNING"
	LevelError    LogLevel = "ERROR"
	LevelSecurity LogLevel = "SECURITY" // Para eventos de seguridad específicos
)

// LogEntry representa una entrada de log estructurada
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Event     string                 `json:"event"`
	Message   string                 `json:"message,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Logger es la instancia global del logger
var Logger = log.New(os.Stdout, "", 0)

// logStructured escribe una entrada de log en formato JSON estructurado
func logStructured(level LogLevel, event string, message string, data map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Event:     event,
		Message:   message,
		Data:      data,
	}

	// Serializar a JSON
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback si el JSON falla
		Logger.Printf("[ERROR] Failed to marshal log entry: %v", err)
		return
	}

	Logger.Println(string(jsonBytes))
}

// Info registra un evento informativo
func Info(event string, message string, data map[string]interface{}) {
	logStructured(LevelInfo, event, message, data)
}

// Warning registra una advertencia
func Warning(event string, message string, data map[string]interface{}) {
	logStructured(LevelWarning, event, message, data)
}

// Error registra un error
func Error(event string, message string, data map[string]interface{}) {
	logStructured(LevelError, event, message, data)
}

// Security registra un evento de seguridad (auth, rate limits, etc.)
func Security(event string, message string, data map[string]interface{}) {
	logStructured(LevelSecurity, event, message, data)
}

// === Auth-specific logging helpers ===

// LogLoginSuccess registra un login exitoso
func LogLoginSuccess(userID, email, ip string) {
	Security("auth.login.success", "Usuario inició sesión exitosamente", map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"ip":      ip,
	})
}

// LogLoginFailed registra un intento fallido de login
func LogLoginFailed(email, ip, reason string) {
	Security("auth.login.failed", "Intento de login fallido", map[string]interface{}{
		"email":  email,
		"ip":     ip,
		"reason": reason,
	})
}

// LogRegisterSuccess registra un registro exitoso
func LogRegisterSuccess(userID, email, ip string) {
	Security("auth.register.success", "Nuevo usuario registrado", map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"ip":      ip,
	})
}

// LogRegisterFailed registra un intento fallido de registro
func LogRegisterFailed(email, ip, reason string) {
	Security("auth.register.failed", "Intento de registro fallido", map[string]interface{}{
		"email":  email,
		"ip":     ip,
		"reason": reason,
	})
}

// LogRateLimitExceeded registra cuando se excede el rate limit
func LogRateLimitExceeded(ip, endpoint string) {
	Security("ratelimit.exceeded", fmt.Sprintf("Rate limit excedido en %s", endpoint), map[string]interface{}{
		"ip":       ip,
		"endpoint": endpoint,
	})
}

// LogInvalidToken registra uso de token inválido
func LogInvalidToken(ip, reason string) {
	Security("auth.token.invalid", "Token inválido o expirado", map[string]interface{}{
		"ip":     ip,
		"reason": reason,
	})
}

// LogRefreshSuccess registra un refresh token exitoso
func LogRefreshSuccess(userID, email, ip string) {
	Security("auth.refresh.success", "Tokens renovados exitosamente", map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"ip":      ip,
	})
}

// LogRefreshFailed registra un intento fallido de refresh
func LogRefreshFailed(ip, reason string) {
	Security("auth.refresh.failed", "Intento de refresh fallido", map[string]interface{}{
		"ip":     ip,
		"reason": reason,
	})
}
