#!/bin/bash

echo "🧪 Probando el flujo completo de delegación..."
echo ""

# 1. Login para obtener token
echo "1️⃣ Login como usuario de prueba..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Error: No se pudo obtener el token"
  echo "Respuesta: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Token obtenido"
echo ""

# 2. Solicitar reporte de productos por email
echo "2️⃣ Solicitando reporte de productos por email..."
REPORT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/reports/products/email \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")

HTTP_CODE=$(echo "$REPORT_RESPONSE" | tail -n1)
BODY=$(echo "$REPORT_RESPONSE" | head -n1)

echo "HTTP Status: $HTTP_CODE"
echo "Response: $BODY"
echo ""

if [ "$HTTP_CODE" = "202" ]; then
  echo "✅ Solicitud aceptada correctamente (202 Accepted)"
else
  echo "❌ Error: Se esperaba HTTP 202, se recibió $HTTP_CODE"
  exit 1
fi

# 3. Verificar que el worker procesó el mensaje
echo "3️⃣ Verificando logs del worker..."
sleep 2
docker logs stock_in_order_worker --tail 10

echo ""
echo "✅ Prueba completada!"
echo "📧 El worker debería haber generado el reporte y procesado el mensaje"
