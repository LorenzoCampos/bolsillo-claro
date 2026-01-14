-- Seed default expense categories (account_id = NULL, is_system = TRUE)
INSERT INTO expense_categories (account_id, name, icon, color, is_system) VALUES
(NULL, 'Alimentación', '🍔', '#FF6B6B', TRUE),
(NULL, 'Transporte', '🚗', '#4ECDC4', TRUE),
(NULL, 'Salud', '⚕️', '#95E1D3', TRUE),
(NULL, 'Entretenimiento', '🎮', '#F38181', TRUE),
(NULL, 'Educación', '📚', '#AA96DA', TRUE),
(NULL, 'Hogar', '🏠', '#FCBAD3', TRUE),
(NULL, 'Servicios', '💡', '#A8D8EA', TRUE),
(NULL, 'Ropa', '👕', '#FFCCBC', TRUE),
(NULL, 'Mascotas', '🐶', '#C5E1A5', TRUE),
(NULL, 'Tecnología', '💻', '#90CAF9', TRUE),
(NULL, 'Viajes', '✈️', '#FFAB91', TRUE),
(NULL, 'Regalos', '🎁', '#F48FB1', TRUE),
(NULL, 'Impuestos', '🧾', '#BCAAA4', TRUE),
(NULL, 'Seguros', '🛡️', '#B39DDB', TRUE),
(NULL, 'Otro', '📦', '#B0BEC5', TRUE);

-- Seed default income categories (account_id = NULL, is_system = TRUE)
INSERT INTO income_categories (account_id, name, icon, color, is_system) VALUES
(NULL, 'Salario', '💼', '#66BB6A', TRUE),
(NULL, 'Freelance', '💻', '#42A5F5', TRUE),
(NULL, 'Inversiones', '📈', '#AB47BC', TRUE),
(NULL, 'Negocio', '🏢', '#FFA726', TRUE),
(NULL, 'Alquiler', '🏘️', '#26C6DA', TRUE),
(NULL, 'Regalo', '🎁', '#EC407A', TRUE),
(NULL, 'Venta', '🏷️', '#78909C', TRUE),
(NULL, 'Intereses', '💰', '#9CCC65', TRUE),
(NULL, 'Reembolso', '↩️', '#7E57C2', TRUE),
(NULL, 'Otro', '💵', '#8D6E63', TRUE);
