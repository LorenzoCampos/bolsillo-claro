-- Migration 004: Crear tabla family_members
-- Almacena los miembros de cuentas familiares

CREATE TABLE IF NOT EXISTS family_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_family_members_account_id ON family_members(account_id);
CREATE INDEX IF NOT EXISTS idx_family_members_is_active ON family_members(is_active);

-- Constraint: no puede haber dos miembros con el mismo nombre en la misma cuenta
CREATE UNIQUE INDEX IF NOT EXISTS idx_family_members_unique_name_per_account 
    ON family_members(account_id, name) WHERE is_active = true;

-- Comentarios
COMMENT ON TABLE family_members IS 'Miembros de cuentas familiares - no son usuarios del sistema, son etiquetas';
COMMENT ON COLUMN family_members.id IS 'Identificador único del miembro';
COMMENT ON COLUMN family_members.account_id IS 'Cuenta familiar a la que pertenece';
COMMENT ON COLUMN family_members.name IS 'Nombre del miembro como "Mamá", "Papá", "Juan"';
COMMENT ON COLUMN family_members.email IS 'Email opcional para funcionalidades futuras';
COMMENT ON COLUMN family_members.is_active IS 'Si el miembro está activo - los inactivos no aparecen en formularios';
