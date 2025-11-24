-- Script para configurar la base de datos de tests
-- Ejecutar como usuario postgres

-- Crear base de datos de test (si no existe)
DROP DATABASE IF EXISTS stock_in_order_test;
CREATE DATABASE stock_in_order_test;

-- Conectar a la base de datos de test
\c stock_in_order_test;

-- Aplicar las migraciones necesarias para los tests
-- (Este script debe ejecutarse después de crear la base de datos)
