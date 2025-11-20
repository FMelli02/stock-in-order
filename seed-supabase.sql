-- ============================================
-- Script de Seed Data para Supabase
-- Stock In Order - Datos iniciales
-- ============================================

-- Obtener el ID del usuario admin
DO $$
DECLARE
    admin_user_id BIGINT;
BEGIN
    -- Obtener ID del usuario admin
    SELECT id INTO admin_user_id FROM users WHERE email = 'francoleproso1@gmail.com' LIMIT 1;
    
    IF admin_user_id IS NULL THEN
        RAISE EXCEPTION 'Usuario admin no encontrado. Por favor crea el usuario primero.';
    END IF;

    -- ============================================
    -- PROVEEDORES
    -- ============================================
    INSERT INTO suppliers (name, contact_person, email, phone, address, user_id, created_at)
    VALUES
        ('Distribuidora Tech Global S.A.', 'Juan Pérez', 'juan.perez@techglobal.com', '+54 11 4567-8901', 'Av. Corrientes 1234, CABA', admin_user_id, NOW()),
        ('Electrónica del Río', 'María González', 'maria.gonzalez@electrorio.com', '+54 11 4567-8902', 'Av. Libertador 5678, CABA', admin_user_id, NOW()),
        ('Importadora ABC Ltda', 'Carlos Rodríguez', 'carlos.rodriguez@abc.com.ar', '+54 11 4567-8903', 'Av. Rivadavia 9012, CABA', admin_user_id, NOW()),
        ('Mayorista Digital Express', 'Ana Martínez', 'ana.martinez@digitalexpress.com', '+54 11 4567-8904', 'Av. Santa Fe 3456, CABA', admin_user_id, NOW()),
        ('Proveedor Premium S.R.L.', 'Luis Fernández', 'luis.fernandez@premium.com', '+54 11 4567-8905', 'Av. Cabildo 7890, CABA', admin_user_id, NOW())
    ON CONFLICT DO NOTHING;

    -- ============================================
    -- CLIENTES
    -- ============================================
    INSERT INTO customers (name, email, phone, address, user_id, created_at)
    VALUES
        ('Tienda Central', 'ventas@tiendacentral.com', '+54 11 5678-1234', 'Av. Corrientes 1234, CABA', admin_user_id, NOW()),
        ('Supermercado Norte', 'compras@supernorte.com', '+54 11 5678-1235', 'Av. Cabildo 5678, CABA', admin_user_id, NOW()),
        ('Comercial Sur', 'info@comercialsur.com', '+54 11 5678-1236', 'Av. Rivadavia 9012, CABA', admin_user_id, NOW()),
        ('Kiosco Digital', 'contacto@kioscodigital.com', '+54 11 5678-1237', 'Av. Santa Fe 3456, CABA', admin_user_id, NOW()),
        ('Distribuidora Este', 'ventas@distribuidoraeste.com', '+54 11 5678-1238', 'Av. Libertador 7890, CABA', admin_user_id, NOW()),
        ('Comercio Minorista XYZ', 'info@xyz.com.ar', '+54 11 5678-1239', 'Av. Belgrano 2345, CABA', admin_user_id, NOW()),
        ('Casa de Electrodomésticos', 'ventas@casaelectro.com', '+54 11 5678-1240', 'Av. Juan B. Justo 4567, CABA', admin_user_id, NOW()),
        ('Tecnología y Más', 'contacto@tecnologiaymas.com', '+54 11 5678-1241', 'Av. Pueyrredón 8901, CABA', admin_user_id, NOW()),
        ('Retail Store SA', 'compras@retailstore.com', '+54 11 5678-1242', 'Av. Córdoba 1234, CABA', admin_user_id, NOW()),
        ('Punto de Venta Express', 'info@puntoventa.com', '+54 11 5678-1243', 'Av. Callao 5678, CABA', admin_user_id, NOW())
    ON CONFLICT DO NOTHING;

    -- ============================================
    -- PRODUCTOS
    -- ============================================
    INSERT INTO products (name, sku, description, user_id, stock_minimo, created_at)
    VALUES
        ('Teclado Mecánico RGB Gamer', 'TEC-001', 'Teclado mecánico con iluminación RGB, switches cherry MX', admin_user_id, 10, NOW()),
        ('Mouse Inalámbrico Ergonómico', 'MOU-001', 'Mouse inalámbrico con diseño ergonómico y 6 botones programables', admin_user_id, 15, NOW()),
        ('Monitor LED 24" Full HD', 'MON-001', 'Monitor LED de 24 pulgadas con resolución Full HD 1920x1080', admin_user_id, 5, NOW()),
        ('Auriculares Bluetooth Premium', 'AUR-001', 'Auriculares bluetooth con cancelación de ruido y micrófono', admin_user_id, 20, NOW()),
        ('Webcam Full HD 1080p', 'WEB-001', 'Cámara web con resolución Full HD y micrófono integrado', admin_user_id, 12, NOW()),
        ('Micrófono USB Profesional', 'MIC-001', 'Micrófono USB de condensador con brazo articulado', admin_user_id, 8, NOW()),
        ('Mousepad Gaming XXL', 'PAD-001', 'Mousepad tamaño XXL con base antideslizante y bordes cosidos', admin_user_id, 30, NOW()),
        ('Cable HDMI 2.0 Premium 2m', 'CAB-001', 'Cable HDMI 2.0 de alta velocidad, 2 metros de longitud', admin_user_id, 50, NOW()),
        ('Hub USB 3.0 4 Puertos', 'HUB-001', 'Hub USB 3.0 con 4 puertos de alta velocidad y LED indicador', admin_user_id, 20, NOW()),
        ('Adaptador USB-C a HDMI', 'ADP-001', 'Adaptador USB Type-C a HDMI 4K compatible', admin_user_id, 25, NOW()),
        ('Disco SSD 500GB SATA', 'SSD-001', 'Unidad de estado sólido SATA de 500GB con velocidad de lectura 550MB/s', admin_user_id, 8, NOW()),
        ('Memoria RAM DDR4 16GB 3200MHz', 'RAM-001', 'Módulo de memoria RAM DDR4 de 16GB a 3200MHz', admin_user_id, 10, NOW()),
        ('Fuente de Poder 650W 80+ Bronze', 'PSU-001', 'Fuente de alimentación modular de 650W con certificación 80+ Bronze', admin_user_id, 6, NOW()),
        ('Cooler CPU RGB con Disipador', 'COO-001', 'Sistema de refrigeración para CPU con iluminación RGB', admin_user_id, 12, NOW()),
        ('Gabinete ATX RGB Vidrio Templado', 'GAB-001', 'Gabinete ATX con panel lateral de vidrio templado e iluminación RGB', admin_user_id, 5, NOW())
    ON CONFLICT (user_id, sku) DO NOTHING;

    -- ============================================
    -- LOTES INICIALES DE PRODUCTOS
    -- ============================================
    INSERT INTO product_batches (product_id, user_id, quantity, lote_number, expiry_date, created_at)
    SELECT 
        p.id,
        admin_user_id,
        batch_data.qty,
        batch_data.lote,
        batch_data.expiry,
        NOW()
    FROM products p
    CROSS JOIN LATERAL (VALUES
        ('TEC-001', 50, 'LOTE-2024-001', NULL::DATE),
        ('MOU-001', 120, 'LOTE-2024-002', NULL::DATE),
        ('MON-001', 30, 'LOTE-2024-003', NULL::DATE),
        ('AUR-001', 80, 'LOTE-2024-004', NULL::DATE),
        ('WEB-001', 45, 'LOTE-2024-005', NULL::DATE),
        ('MIC-001', 35, 'LOTE-2024-006', NULL::DATE),
        ('PAD-001', 200, 'LOTE-2024-007', NULL::DATE),
        ('CAB-001', 300, 'LOTE-2024-008', NULL::DATE),
        ('HUB-001', 75, 'LOTE-2024-009', NULL::DATE),
        ('ADP-001', 90, 'LOTE-2024-010', NULL::DATE),
        ('SSD-001', 40, 'LOTE-2024-011', NULL::DATE),
        ('RAM-001', 60, 'LOTE-2024-012', NULL::DATE),
        ('PSU-001', 25, 'LOTE-2024-013', NULL::DATE),
        ('COO-001', 55, 'LOTE-2024-014', NULL::DATE),
        ('GAB-001', 20, 'LOTE-2024-015', NULL::DATE)
    ) AS batch_data(sku, qty, lote, expiry)
    WHERE p.sku = batch_data.sku
    AND p.user_id = admin_user_id
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Datos iniciales insertados correctamente';
END $$;
