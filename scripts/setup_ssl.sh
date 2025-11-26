#!/bin/bash

#############################################
# 🔒 Setup SSL con Let's Encrypt + Certbot
#############################################
# 
# Este script automatiza la instalación y configuración
# de certificados SSL gratuitos con Let's Encrypt.
#
# PREREQUISITOS:
# - Ubuntu 20.04+ / Debian 10+
# - Nginx instalado y corriendo
# - Dominio apuntando a la IP del servidor (DNS configurado)
# - Puerto 80 (HTTP) y 443 (HTTPS) abiertos en firewall
#
# USO:
#   sudo ./setup_ssl.sh tu-dominio.com [www.tu-dominio.com]
#
# EJEMPLO:
#   sudo ./setup_ssl.sh stock-in-order.com www.stock-in-order.com
#
#############################################

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Banner
echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════╗"
echo "║   🔒 CONFIGURACIÓN SSL CON LET'S ENCRYPT       ║"
echo "║   Certbot + Nginx + Auto-renovación            ║"
echo "╚════════════════════════════════════════════════╝"
echo -e "${NC}"

#############################################
# 1. VALIDAR PREREQUISITOS
#############################################

echo -e "${YELLOW}[1/6] Validando prerequisitos...${NC}"

# Verificar que se ejecuta como root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}❌ ERROR: Este script debe ejecutarse como root (sudo)${NC}"
    exit 1
fi

# Verificar que se proporcionó un dominio
if [ $# -eq 0 ]; then
    echo -e "${RED}❌ ERROR: Debes proporcionar al menos un dominio${NC}"
    echo -e "${YELLOW}Uso: sudo ./setup_ssl.sh tu-dominio.com [www.tu-dominio.com]${NC}"
    exit 1
fi

DOMAIN=$1
WWW_DOMAIN=${2:-}

echo -e "${GREEN}✓ Dominio principal: $DOMAIN${NC}"
if [ -n "$WWW_DOMAIN" ]; then
    echo -e "${GREEN}✓ Dominio secundario: $WWW_DOMAIN${NC}"
fi

# Verificar que Nginx está instalado
if ! command -v nginx &> /dev/null; then
    echo -e "${RED}❌ ERROR: Nginx no está instalado${NC}"
    echo -e "${YELLOW}Instala Nginx primero: sudo apt install nginx -y${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Nginx está instalado${NC}"

# Verificar que Nginx está corriendo
if ! systemctl is-active --quiet nginx; then
    echo -e "${YELLOW}⚠ Nginx no está corriendo. Iniciando...${NC}"
    systemctl start nginx
fi
echo -e "${GREEN}✓ Nginx está corriendo${NC}"

# Verificar puertos abiertos
if ! netstat -tuln | grep -q ':80 '; then
    echo -e "${RED}❌ ERROR: Puerto 80 no está escuchando${NC}"
    echo -e "${YELLOW}Verifica la configuración de Nginx y el firewall${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Puerto 80 (HTTP) está abierto${NC}"

#############################################
# 2. VERIFICAR DNS (CRÍTICO)
#############################################

echo ""
echo -e "${YELLOW}[2/6] Verificando configuración DNS...${NC}"

SERVER_IP=$(curl -s ifconfig.me || curl -s icanhazip.com)
echo -e "${BLUE}IP del servidor: $SERVER_IP${NC}"

# Resolver dominio
DOMAIN_IP=$(dig +short $DOMAIN @8.8.8.8 | tail -n1)

if [ -z "$DOMAIN_IP" ]; then
    echo -e "${RED}❌ ERROR: No se pudo resolver el dominio $DOMAIN${NC}"
    echo -e "${YELLOW}Verifica que:${NC}"
    echo -e "${YELLOW}  1. El dominio existe${NC}"
    echo -e "${YELLOW}  2. Los registros DNS están configurados${NC}"
    echo -e "${YELLOW}  3. Los cambios DNS se propagaron (puede tardar hasta 48h)${NC}"
    exit 1
fi

echo -e "${BLUE}IP del dominio $DOMAIN: $DOMAIN_IP${NC}"

if [ "$SERVER_IP" != "$DOMAIN_IP" ]; then
    echo -e "${RED}❌ ERROR: El dominio NO apunta a este servidor${NC}"
    echo -e "${YELLOW}Servidor: $SERVER_IP${NC}"
    echo -e "${YELLOW}Dominio:  $DOMAIN_IP${NC}"
    echo -e "${YELLOW}Configura un registro A en tu proveedor DNS apuntando a: $SERVER_IP${NC}"
    exit 1
fi

echo -e "${GREEN}✓ DNS configurado correctamente (dominio apunta a este servidor)${NC}"

# Verificar WWW si fue proporcionado
if [ -n "$WWW_DOMAIN" ]; then
    WWW_IP=$(dig +short $WWW_DOMAIN @8.8.8.8 | tail -n1)
    if [ "$SERVER_IP" != "$WWW_IP" ]; then
        echo -e "${YELLOW}⚠ ADVERTENCIA: $WWW_DOMAIN no apunta a este servidor${NC}"
        echo -e "${YELLOW}Certbot solo configurará SSL para el dominio principal${NC}"
        WWW_DOMAIN=""
    else
        echo -e "${GREEN}✓ $WWW_DOMAIN también apunta correctamente${NC}"
    fi
fi

#############################################
# 3. ACTUALIZAR SISTEMA E INSTALAR CERTBOT
#############################################

echo ""
echo -e "${YELLOW}[3/6] Actualizando sistema e instalando Certbot...${NC}"

echo -e "${BLUE}Actualizando lista de paquetes...${NC}"
apt update -qq

echo -e "${BLUE}Instalando Certbot y plugin de Nginx...${NC}"
apt install certbot python3-certbot-nginx -y

# Verificar instalación
if ! command -v certbot &> /dev/null; then
    echo -e "${RED}❌ ERROR: Certbot no se instaló correctamente${NC}"
    exit 1
fi

CERTBOT_VERSION=$(certbot --version 2>&1 | head -n1)
echo -e "${GREEN}✓ Certbot instalado: $CERTBOT_VERSION${NC}"

#############################################
# 4. VERIFICAR CONFIGURACIÓN DE NGINX
#############################################

echo ""
echo -e "${YELLOW}[4/6] Verificando configuración de Nginx...${NC}"

# Buscar archivo de configuración del dominio
NGINX_CONFIG="/etc/nginx/sites-available/$DOMAIN"
NGINX_ENABLED="/etc/nginx/sites-enabled/$DOMAIN"

if [ ! -f "$NGINX_CONFIG" ]; then
    echo -e "${YELLOW}⚠ No existe configuración específica para $DOMAIN${NC}"
    echo -e "${YELLOW}Certbot modificará /etc/nginx/sites-available/default${NC}"
    NGINX_CONFIG="/etc/nginx/sites-available/default"
fi

echo -e "${BLUE}Configuración de Nginx: $NGINX_CONFIG${NC}"

# Verificar sintaxis de Nginx
if nginx -t &> /dev/null; then
    echo -e "${GREEN}✓ Configuración de Nginx es válida${NC}"
else
    echo -e "${RED}❌ ERROR: Configuración de Nginx inválida${NC}"
    nginx -t
    exit 1
fi

#############################################
# 5. EJECUTAR CERTBOT (El Momento Mágico 🪄)
#############################################

echo ""
echo -e "${YELLOW}[5/6] Ejecutando Certbot para obtener certificado SSL...${NC}"
echo -e "${BLUE}Esto puede tardar unos minutos...${NC}"

# Construir comando de certbot
CERTBOT_CMD="certbot --nginx -d $DOMAIN"
if [ -n "$WWW_DOMAIN" ]; then
    CERTBOT_CMD="$CERTBOT_CMD -d $WWW_DOMAIN"
fi

# Agregar opciones no interactivas
CERTBOT_CMD="$CERTBOT_CMD --non-interactive --agree-tos --redirect"

# Solicitar email para notificaciones
read -p "$(echo -e ${BLUE}Email para notificaciones de Let\'s Encrypt: ${NC})" EMAIL
if [ -n "$EMAIL" ]; then
    CERTBOT_CMD="$CERTBOT_CMD --email $EMAIL"
else
    CERTBOT_CMD="$CERTBOT_CMD --register-unsafely-without-email"
fi

echo -e "${BLUE}Ejecutando: $CERTBOT_CMD${NC}"
echo ""

# Ejecutar certbot
if eval $CERTBOT_CMD; then
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   ✅ CERTIFICADO SSL INSTALADO EXITOSAMENTE    ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}"
else
    echo -e "${RED}❌ ERROR: Certbot falló al obtener el certificado${NC}"
    echo -e "${YELLOW}Posibles causas:${NC}"
    echo -e "${YELLOW}  1. DNS no propagado completamente${NC}"
    echo -e "${YELLOW}  2. Firewall bloqueando puerto 80/443${NC}"
    echo -e "${YELLOW}  3. Nginx no configurado correctamente${NC}"
    echo -e "${YELLOW}  4. Rate limit de Let's Encrypt alcanzado${NC}"
    exit 1
fi

#############################################
# 6. VERIFICAR AUTO-RENOVACIÓN
#############################################

echo ""
echo -e "${YELLOW}[6/6] Configurando auto-renovación de certificados...${NC}"

# Verificar que el timer de renovación está activo
if systemctl is-active --quiet certbot.timer; then
    echo -e "${GREEN}✓ Timer de renovación automática está activo${NC}"
else
    echo -e "${YELLOW}⚠ Activando timer de renovación...${NC}"
    systemctl enable certbot.timer
    systemctl start certbot.timer
fi

# Mostrar status del timer
echo -e "${BLUE}Estado del timer de renovación:${NC}"
systemctl status certbot.timer --no-pager | grep -E "(Active|Trigger)"

# Hacer un dry-run de renovación
echo ""
echo -e "${BLUE}Probando renovación de certificado (dry-run)...${NC}"
if certbot renew --dry-run &> /dev/null; then
    echo -e "${GREEN}✓ La renovación automática funcionará correctamente${NC}"
else
    echo -e "${YELLOW}⚠ Hubo un problema en el dry-run, pero el certificado está instalado${NC}"
fi

#############################################
# 7. VERIFICAR CERTIFICADO
#############################################

echo ""
echo -e "${YELLOW}Verificando certificado instalado...${NC}"

# Mostrar información del certificado
certbot certificates | grep -A 5 "$DOMAIN" || true

#############################################
# 8. RECARGAR NGINX
#############################################

echo ""
echo -e "${YELLOW}Recargando configuración de Nginx...${NC}"
systemctl reload nginx
echo -e "${GREEN}✓ Nginx recargado con la nueva configuración SSL${NC}"

#############################################
# RESUMEN FINAL
#############################################

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                                                ║${NC}"
echo -e "${GREEN}║          🎉 ¡CONFIGURACIÓN COMPLETADA! 🎉      ║${NC}"
echo -e "${GREEN}║                                                ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}✓ Certificado SSL instalado para: $DOMAIN${NC}"
if [ -n "$WWW_DOMAIN" ]; then
    echo -e "${BLUE}✓ Certificado también válido para: $WWW_DOMAIN${NC}"
fi
echo -e "${BLUE}✓ Redirección automática HTTP → HTTPS configurada${NC}"
echo -e "${BLUE}✓ Auto-renovación activada (cada 12 horas)${NC}"
echo -e "${BLUE}✓ Certificado válido por 90 días${NC}"
echo ""
echo -e "${GREEN}🌐 Tu sitio ahora está accesible en:${NC}"
echo -e "${GREEN}   https://$DOMAIN${NC}"
if [ -n "$WWW_DOMAIN" ]; then
    echo -e "${GREEN}   https://$WWW_DOMAIN${NC}"
fi
echo ""
echo -e "${YELLOW}📋 Comandos útiles:${NC}"
echo -e "${YELLOW}   Ver certificados:       certbot certificates${NC}"
echo -e "${YELLOW}   Renovar manualmente:    certbot renew${NC}"
echo -e "${YELLOW}   Estado de renovación:   systemctl status certbot.timer${NC}"
echo -e "${YELLOW}   Ver logs:               journalctl -u certbot${NC}"
echo ""
echo -e "${GREEN}✨ ¡Tu sitio ahora tiene el candadito verde! 🔒✨${NC}"
echo ""
