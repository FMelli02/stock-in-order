-- Crear tipo ENUM para los roles de usuario
CREATE TYPE user_role AS ENUM ('admin', 'vendedor', 'repositor');

-- Añadir columna role a la tabla users
ALTER TABLE users 
ADD COLUMN role user_role NOT NULL DEFAULT 'vendedor';

-- Actualizar el usuario admin existente (si existe) para que tenga rol admin
UPDATE users 
SET role = 'admin' 
WHERE email = 'admin@stockinorder.com';
