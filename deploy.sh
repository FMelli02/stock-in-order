#!/bin/bash

# Stock-in-Order Deployment Script
# Este script facilita el despliegue en producción

set -e

COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

echo -e "${COLOR_BLUE}================================${COLOR_RESET}"
echo -e "${COLOR_BLUE}Stock-in-Order Deployment${COLOR_RESET}"
echo -e "${COLOR_BLUE}================================${COLOR_RESET}"

# Verificar que existe .env.prod
if [ ! -f .env.prod ]; then
    echo -e "${COLOR_RED}Error: .env.prod no encontrado${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}Copia .env.prod.example a .env.prod y configúralo:${COLOR_RESET}"
    echo "  cp .env.prod.example .env.prod"
    echo "  nano .env.prod"
    exit 1
fi

# Cargar variables de entorno
source .env.prod

echo -e "${COLOR_GREEN}✓ Configuración cargada${COLOR_RESET}"

# Verificar variables críticas
if [ -z "$JWT_SECRET" ]; then
    echo -e "${COLOR_RED}Error: JWT_SECRET no configurado en .env.prod${COLOR_RESET}"
    exit 1
fi

if [ -z "$DB_PASSWORD" ] || [ "$DB_PASSWORD" == "CAMBIAR_ESTE_PASSWORD_SEGURO" ]; then
    echo -e "${COLOR_RED}Error: DB_PASSWORD debe ser configurado en .env.prod${COLOR_RESET}"
    exit 1
fi

echo -e "${COLOR_GREEN}✓ Variables críticas configuradas${COLOR_RESET}"

# Login a GHCR si es necesario
if [ ! -z "$GITHUB_TOKEN" ]; then
    echo -e "${COLOR_BLUE}Logging in to GitHub Container Registry...${COLOR_RESET}"
    echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin
    echo -e "${COLOR_GREEN}✓ Login exitoso${COLOR_RESET}"
fi

# Pull de las últimas imágenes
echo -e "${COLOR_BLUE}Pulling latest images...${COLOR_RESET}"
docker compose -f docker-compose.prod.yml --env-file .env.prod pull

# Detener contenedores existentes
echo -e "${COLOR_BLUE}Stopping existing containers...${COLOR_RESET}"
docker compose -f docker-compose.prod.yml --env-file .env.prod down

# Iniciar nuevos contenedores
echo -e "${COLOR_BLUE}Starting containers...${COLOR_RESET}"
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

# Esperar a que los servicios estén healthy
echo -e "${COLOR_BLUE}Waiting for services to be healthy...${COLOR_RESET}"
sleep 10

# Verificar estado
echo -e "${COLOR_BLUE}Checking container status...${COLOR_RESET}"
docker compose -f docker-compose.prod.yml ps

# Verificar salud de la API
echo -e "${COLOR_BLUE}Checking API health...${COLOR_RESET}"
API_PORT=${API_PORT:-8080}
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$API_PORT/api/v1/health || echo "000")

if [ "$HEALTH_CHECK" == "200" ]; then
    echo -e "${COLOR_GREEN}✓ API is healthy${COLOR_RESET}"
else
    echo -e "${COLOR_RED}✗ API health check failed (status: $HEALTH_CHECK)${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}Check logs:${COLOR_RESET}"
    echo "  docker compose -f docker-compose.prod.yml logs api"
    exit 1
fi

echo ""
echo -e "${COLOR_GREEN}================================${COLOR_RESET}"
echo -e "${COLOR_GREEN}✓ Deployment successful!${COLOR_RESET}"
echo -e "${COLOR_GREEN}================================${COLOR_RESET}"
echo ""
echo -e "${COLOR_BLUE}Services:${COLOR_RESET}"
echo "  API:      http://localhost:$API_PORT"
echo "  Frontend: http://localhost:${FRONTEND_PORT:-80}"
echo ""
echo -e "${COLOR_BLUE}Commands:${COLOR_RESET}"
echo "  View logs:    docker compose -f docker-compose.prod.yml logs -f"
echo "  Stop:         docker compose -f docker-compose.prod.yml down"
echo "  Restart:      docker compose -f docker-compose.prod.yml restart"
echo ""
