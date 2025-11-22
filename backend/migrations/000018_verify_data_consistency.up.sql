-- Migración: Actualizar products, customers y suppliers para usar organization_id del usuario creador
-- Esto es necesario porque ahora filtramos por organization_id, no por user_id

-- 1. Actualizar productos: el user_id debe reflejar el organization_id del usuario que los creó
-- Para el admin (ID=5), organization_id=5, así que los productos quedan con user_id=5
-- Esto funciona porque el admin tiene organization_id = su propio ID

-- Verificar el estado actual
SELECT 
    'BEFORE UPDATE' as stage,
    u.id as user_id, 
    u.email, 
    u.organization_id,
    COUNT(p.id) as product_count
FROM users u
LEFT JOIN products p ON p.user_id = u.id
GROUP BY u.id, u.email, u.organization_id
ORDER BY u.id;

-- Los productos ya están correctos porque:
-- - El admin (ID=5) tiene organization_id=5
-- - Los productos tienen user_id=5
-- - Cuando buscamos por organization_id=5, debemos buscar user_id=5

-- El problema es que el código busca productos donde user_id = organization_id del usuario actual
-- Si un vendedor tiene organization_id=5, debe buscar productos con user_id=5

-- NO necesitamos migración de datos, el problema está en el código
-- Los productos con user_id=5 están bien para organization_id=5

SELECT 
    'VERIFICATION' as stage,
    u.id as user_id,
    u.email,
    u.role,
    u.organization_id,
    COUNT(p.id) as products_count
FROM users u
LEFT JOIN products p ON p.user_id = u.organization_id
WHERE u.organization_id IS NOT NULL
GROUP BY u.id, u.email, u.role, u.organization_id
ORDER BY u.id;
