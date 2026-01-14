package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB encapsula el pool de conexiones a PostgreSQL
// Usamos un pool en lugar de conexiones individuales para mejor rendimiento
type DB struct {
	Pool *pgxpool.Pool // Pool de conexiones reutilizables
}

// New crea una nueva conexión a PostgreSQL usando un connection pool
// databaseURL debe tener formato: postgresql://usuario:password@host:puerto/nombre_db
func New(databaseURL string) (*DB, error) {
	// context.Background() crea un contexto vacío
	// En Go, los contextos se usan para manejar timeouts y cancelaciones
	ctx := context.Background()

	// Configuración del pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parseando DATABASE_URL: %w", err)
	}

	// Configurar límites del pool
	// MaxConns: máximo de conexiones abiertas simultáneamente
	// MinConns: conexiones que se mantienen abiertas siempre (warm pool)
	// MaxConnLifetime: cuánto tiempo vive una conexión antes de reciclarse
	// MaxConnIdleTime: cuánto tiempo puede estar idle antes de cerrarse
	config.MaxConns = 10                      // Máximo 10 conexiones simultáneas
	config.MinConns = 2                       // Mínimo 2 conexiones siempre abiertas
	config.MaxConnLifetime = time.Hour        // Reciclar conexiones cada hora
	config.MaxConnIdleTime = time.Minute * 30 // Cerrar conexiones idle después de 30 min

	// Crear el pool de conexiones
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creando pool de conexiones: %w", err)
	}

	// Verificar que podemos conectarnos haciendo un ping
	// Ping envía una query simple para verificar conectividad
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close() // Si falla, cerrar el pool antes de retornar error
		return nil, fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	fmt.Println("✅ Conexión a PostgreSQL establecida correctamente")

	return &DB{Pool: pool}, nil
}

// Close cierra el pool de conexiones
// Debe llamarse cuando la aplicación se apaga para liberar recursos
func (db *DB) Close() {
	db.Pool.Close()
	fmt.Println("🔌 Pool de conexiones PostgreSQL cerrado")
}

// Ping verifica que la conexión a la base de datos sigue activa
// Útil para health checks
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats retorna estadísticas del pool de conexiones
// Útil para debugging y monitoreo
func (db *DB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}
