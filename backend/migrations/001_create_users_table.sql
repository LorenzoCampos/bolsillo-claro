-- Migration 001: Crear tabla users
-- Esta tabla almacena los usuarios del sistema

-- Habilitar la extensión uuid-ossp para generar UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Crear tabla users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Crear índice en email para búsquedas rápidas
-- Como email es UNIQUE, PostgreSQL crea un índice automáticamente,
-- pero lo declaramos explícitamente para claridad
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Comentarios para documentación
COMMENT ON TABLE users IS 'Usuarios del sistema - cada usuario puede administrar múltiples cuentas';
COMMENT ON COLUMN users.id IS 'Identificador único del usuario (UUID v4)';
COMMENT ON COLUMN users.email IS 'Email del usuario - usado para login - debe ser único';
COMMENT ON COLUMN users.password_hash IS 'Hash bcrypt de la contraseña - NUNCA almacenar password en texto plano';
COMMENT ON COLUMN users.name IS 'Nombre completo del usuario para mostrar en UI';
COMMENT ON COLUMN users.created_at IS 'Timestamp de cuando se creó la cuenta';
COMMENT ON COLUMN users.updated_at IS 'Timestamp de última actualización de datos del usuario';
