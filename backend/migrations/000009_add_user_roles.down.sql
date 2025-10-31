-- Eliminar la columna role de la tabla users
ALTER TABLE users 
DROP COLUMN IF EXISTS role;

-- Eliminar el tipo ENUM user_role
DROP TYPE IF EXISTS user_role;
