-- Seeding inicial de datos para desarrollo
-- Este archivo no tiene un .down.sql asociado para mantener los datos de prueba

-- ============================================
-- 1. Usuario de prueba
-- ============================================
-- Email: test@example.com
-- Password: password123
INSERT INTO users (name, email, password_hash, created_at)
VALUES (
    'Usuario de Prueba',
    'test@example.com',
    decode('24326124313024495571796c4f6b49704f696963574d3930764a36712e556534473279444f4f4944687033304875486477385258353972

4c43456369', 'hex'),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- ============================================
-- 2. Proveedores, Clientes y Productos
-- ============================================

-- Insertar proveedores usando el usuario de prueba
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
)
INSERT INTO suppliers (name, contact_person, email, phone, user_id, created_at)
SELECT s.name, s.contact_person, s.email, s.phone, tu.id, NOW()
FROM test_user tu,
(VALUES
    ('Distribuidora Tech S.A.', 'Juan Pérez', 'juan.perez@techsa.com', '+54 11 4567-8901'),
    ('Electrónica Global', 'María González', 'maria.gonzalez@elecglobal.com', '+54 11 4567-8902'),
    ('Importadora ABC', 'Carlos Rodríguez', 'carlos.rodriguez@abc.com', '+54 11 4567-8903'),
    ('Mayorista Digital', 'Ana Martínez', 'ana.martinez@mayoristadigital.com', '+54 11 4567-8904'),
    ('Proveedor Express', 'Luis Fernández', 'luis.fernandez@express.com', '+54 11 4567-8905')
) AS s(name, contact_person, email, phone)
WHERE NOT EXISTS (SELECT 1 FROM suppliers WHERE suppliers.email = s.email);

-- Insertar clientes
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
)
INSERT INTO customers (name, email, phone, address, user_id, created_at)
SELECT c.name, c.email, c.phone, c.address, tu.id, NOW()
FROM test_user tu,
(VALUES
    ('Tienda Central', 'ventas@tiendacentral.com', '+54 11 5678-1234', 'Av. Corrientes 1234, CABA'),
    ('Supermercado Norte', 'compras@supernorte.com', '+54 11 5678-1235', 'Av. Cabildo 5678, CABA'),
    ('Comercial Sur', 'info@comercialsur.com', '+54 11 5678-1236', 'Av. Rivadavia 9012, CABA'),
    ('Kiosco Digital', 'contacto@kioscodigital.com', '+54 11 5678-1237', 'Av. Santa Fe 3456, CABA'),
    ('Distribuidora Este', 'ventas@distribuidoraeste.com', '+54 11 5678-1238', 'Av. Libertador 7890, CABA')
) AS c(name, email, phone, address)
WHERE NOT EXISTS (SELECT 1 FROM customers WHERE customers.email = c.email);

-- Insertar productos
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
)
INSERT INTO products (name, sku, quantity, user_id, created_at)
SELECT p.name, p.sku, p.quantity, tu.id, NOW()
FROM test_user tu,
(VALUES
    ('Teclado Mecánico RGB', 'TEC-001', 150),
    ('Mouse Inalámbrico', 'MOU-001', 200),
    ('Monitor LED 24"', 'MON-001', 50),
    ('Auriculares Bluetooth', 'AUR-001', 120),
    ('Webcam Full HD', 'WEB-001', 80),
    ('Micrófono USB', 'MIC-001', 60),
    ('Mousepad Gaming XXL', 'PAD-001', 300),
    ('Cable HDMI 2.0', 'CAB-001', 250),
    ('Hub USB 3.0 4 Puertos', 'HUB-001', 100),
    ('Adaptador USB-C a HDMI', 'ADP-001', 90),
    ('Disco SSD 500GB', 'SSD-001', 70),
    ('Memoria RAM DDR4 16GB', 'RAM-001', 85),
    ('Fuente de Poder 650W', 'PSU-001', 45),
    ('Cooler CPU RGB', 'COO-001', 55),
    ('Pasta Térmica', 'PAS-001', 180),
    ('Gabinete ATX RGB', 'GAB-001', 30),
    ('Placa de Video GTX 1660', 'GPU-001', 20),
    ('Motherboard B450', 'MOB-001', 25),
    ('Procesador Ryzen 5', 'CPU-001', 35),
    ('Kit Limpieza PC', 'KIT-001', 220)
) AS p(name, sku, quantity)
WHERE NOT EXISTS (
    SELECT 1 FROM products pr 
    JOIN users u ON pr.user_id = u.id 
    WHERE pr.sku = p.sku AND u.email = 'test@example.com'
);

-- ============================================
-- 3. Órdenes de Venta
-- ============================================

-- Orden de Venta #1
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
),
new_order AS (
    INSERT INTO sales_orders (customer_id, order_date, status, user_id)
    SELECT 
        c.id,
        NOW() - INTERVAL '5 days',
        'completed',
        tu.id
    FROM test_user tu
    CROSS JOIN customers c
    WHERE c.email = 'ventas@tiendacentral.com'
    AND NOT EXISTS (
        SELECT 1 FROM sales_orders so
        WHERE so.customer_id = c.id
        AND so.order_date::DATE = (NOW() - INTERVAL '5 days')::DATE
    )
    RETURNING id
)
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
SELECT 
    no.id,
    p.id,
    items.qty,
    items.price
FROM new_order no
CROSS JOIN test_user tu
CROSS JOIN (VALUES
    ('TEC-001', 10, 45.99),
    ('MOU-001', 15, 25.50),
    ('PAD-001', 20, 12.99)
) AS items(sku, qty, price)
JOIN products p ON p.sku = items.sku AND p.user_id = tu.id;

-- Orden de Venta #2
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
),
new_order AS (
    INSERT INTO sales_orders (customer_id, order_date, status, user_id)
    SELECT 
        c.id,
        NOW() - INTERVAL '3 days',
        'completed',
        tu.id
    FROM test_user tu
    CROSS JOIN customers c
    WHERE c.email = 'compras@supernorte.com'
    AND NOT EXISTS (
        SELECT 1 FROM sales_orders so
        WHERE so.customer_id = c.id
        AND so.order_date::DATE = (NOW() - INTERVAL '3 days')::DATE
    )
    RETURNING id
)
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
SELECT 
    no.id,
    p.id,
    items.qty,
    items.price
FROM new_order no
CROSS JOIN test_user tu
CROSS JOIN (VALUES
    ('MON-001', 5, 199.99),
    ('AUR-001', 12, 35.00),
    ('WEB-001', 8, 55.00)
) AS items(sku, qty, price)
JOIN products p ON p.sku = items.sku AND p.user_id = tu.id;

-- Orden de Venta #3
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
),
new_order AS (
    INSERT INTO sales_orders (customer_id, order_date, status, user_id)
    SELECT 
        c.id,
        NOW() - INTERVAL '1 day',
        'completed',
        tu.id
    FROM test_user tu
    CROSS JOIN customers c
    WHERE c.email = 'info@comercialsur.com'
    AND NOT EXISTS (
        SELECT 1 FROM sales_orders so
        WHERE so.customer_id = c.id
        AND so.order_date::DATE = (NOW() - INTERVAL '1 day')::DATE
    )
    RETURNING id
)
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
SELECT 
    no.id,
    p.id,
    items.qty,
    items.price
FROM new_order no
CROSS JOIN test_user tu
CROSS JOIN (VALUES
    ('CAB-001', 30, 8.99),
    ('HUB-001', 10, 18.50),
    ('ADP-001', 15, 22.00),
    ('KIT-001', 25, 9.99)
) AS items(sku, qty, price)
JOIN products p ON p.sku = items.sku AND p.user_id = tu.id;

-- ============================================
-- 4. Órdenes de Compra
-- ============================================

-- Orden de Compra #1
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
),
new_order AS (
    INSERT INTO purchase_orders (supplier_id, order_date, status, user_id)
    SELECT 
        s.id,
        NOW() - INTERVAL '2 days',
        'pending',
        tu.id
    FROM test_user tu
    CROSS JOIN suppliers s
    WHERE s.email = 'juan.perez@techsa.com'
    AND NOT EXISTS (
        SELECT 1 FROM purchase_orders po
        WHERE po.supplier_id = s.id
        AND po.order_date::DATE = (NOW() - INTERVAL '2 days')::DATE
    )
    RETURNING id
)
INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity, unit_cost)
SELECT 
    no.id,
    p.id,
    items.qty,
    items.cost
FROM new_order no
CROSS JOIN test_user tu
CROSS JOIN (VALUES
    ('TEC-001', 50, 30.00),
    ('MOU-001', 100, 15.00),
    ('PAD-001', 150, 7.50)
) AS items(sku, qty, cost)
JOIN products p ON p.sku = items.sku AND p.user_id = tu.id;

-- Orden de Compra #2
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
),
new_order AS (
    INSERT INTO purchase_orders (supplier_id, order_date, status, user_id)
    SELECT 
        s.id,
        NOW() - INTERVAL '7 days',
        'completed',
        tu.id
    FROM test_user tu
    CROSS JOIN suppliers s
    WHERE s.email = 'maria.gonzalez@elecglobal.com'
    AND NOT EXISTS (
        SELECT 1 FROM purchase_orders po
        WHERE po.supplier_id = s.id
        AND po.order_date::DATE = (NOW() - INTERVAL '7 days')::DATE
    )
    RETURNING id
)
INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity, unit_cost)
SELECT 
    no.id,
    p.id,
    items.qty,
    items.cost
FROM new_order no
CROSS JOIN test_user tu
CROSS JOIN (VALUES
    ('MON-001', 30, 120.00),
    ('AUR-001', 50, 20.00),
    ('WEB-001', 40, 35.00)
) AS items(sku, qty, cost)
JOIN products p ON p.sku = items.sku AND p.user_id = tu.id;

-- Orden de Compra #3
WITH test_user AS (
    SELECT id FROM users WHERE email = 'test@example.com' LIMIT 1
),
new_order AS (
    INSERT INTO purchase_orders (supplier_id, order_date, status, user_id)
    SELECT 
        s.id,
        NOW(),
        'pending',
        tu.id
    FROM test_user tu
    CROSS JOIN suppliers s
    WHERE s.email = 'carlos.rodriguez@abc.com'
    AND NOT EXISTS (
        SELECT 1 FROM purchase_orders po
        WHERE po.supplier_id = s.id
        AND po.order_date::DATE = NOW()::DATE
    )
    RETURNING id
)
INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity, unit_cost)
SELECT 
    no.id,
    p.id,
    items.qty,
    items.cost
FROM new_order no
CROSS JOIN test_user tu
CROSS JOIN (VALUES
    ('SSD-001', 40, 65.00),
    ('RAM-001', 30, 55.00),
    ('PSU-001', 20, 80.00),
    ('GPU-001', 10, 250.00)
) AS items(sku, qty, cost)
JOIN products p ON p.sku = items.sku AND p.user_id = tu.id;

-- ============================================
-- Fin del seeding
-- ============================================
