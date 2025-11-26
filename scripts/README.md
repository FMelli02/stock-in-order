# 📜 Scripts de Despliegue

Este directorio contiene scripts de automatización para el despliegue y configuración del servidor de producción.

## 📋 Scripts Disponibles

### 🔒 `setup_ssl.sh` - Configuración Automática de SSL

**Propósito:** Instala y configura certificados SSL gratuitos de Let's Encrypt con Certbot.

**Prerequisitos:**
- Ubuntu 20.04+ / Debian 10+
- Nginx instalado y corriendo
- Dominio apuntando a la IP del servidor (DNS configurado)
- Puertos 80 y 443 abiertos en firewall

**Uso:**
```bash
# Con dominio principal solamente
sudo ./setup_ssl.sh tu-dominio.com

# Con dominio principal + www
sudo ./setup_ssl.sh tu-dominio.com www.tu-dominio.com

# Ejemplo real
sudo ./setup_ssl.sh stock-in-order.com www.stock-in-order.com
```

**Qué hace:**
1. ✅ Valida prerequisitos (Nginx, DNS, puertos)
2. ✅ Instala Certbot y plugin de Nginx
3. ✅ Obtiene certificado SSL de Let's Encrypt
4. ✅ Configura Nginx para HTTPS automáticamente
5. ✅ Activa redirección HTTP → HTTPS
6. ✅ Configura renovación automática cada 60 días
7. ✅ Verifica que todo funcione correctamente

**Output esperado:**
- Certificado SSL válido por 90 días
- Sitio accesible en `https://tu-dominio.com`
- Candadito verde 🔒 en navegadores
- Auto-renovación configurada

**Documentación completa:** Ver [`../docs/SSL_SETUP_GUIDE.md`](../docs/SSL_SETUP_GUIDE.md)

---

## 🚀 Uso en Servidor de Producción

### 1. Copiar Scripts al Servidor

```bash
# Desde tu máquina local
scp setup_ssl.sh usuario@tu-servidor:/tmp/

# O clonar el repositorio en el servidor
ssh usuario@tu-servidor
cd /opt
sudo git clone https://github.com/tu-usuario/stock-in-order.git
cd stock-in-order/scripts
```

### 2. Dar Permisos de Ejecución

```bash
chmod +x setup_ssl.sh
```

### 3. Ejecutar Script

```bash
sudo ./setup_ssl.sh tu-dominio.com www.tu-dominio.com
```

---

## ⚠️ Notas Importantes

### DNS Debe Estar Configurado

**CRÍTICO**: El dominio **DEBE** apuntar a la IP del servidor **ANTES** de ejecutar el script.

Verificar DNS:
```bash
# Obtener IP del servidor
curl ifconfig.me

# Verificar que el dominio apunta a esa IP
dig +short tu-dominio.com

# Deben coincidir
```

### Puertos Abiertos

Asegurar que el firewall permita tráfico en:
- Puerto **80** (HTTP) - Requerido para validación de Let's Encrypt
- Puerto **443** (HTTPS) - Requerido para servir el sitio con SSL

```bash
# Ubuntu con UFW
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw status
```

### Ejecutar Como Root

Los scripts requieren permisos de superusuario:
```bash
# CORRECTO ✅
sudo ./setup_ssl.sh tu-dominio.com

# INCORRECTO ❌
./setup_ssl.sh tu-dominio.com
```

---

## 🔧 Troubleshooting

### Error: "DNS no apunta al servidor"

**Causa:** Registro A del dominio no está configurado o no se propagó.

**Solución:**
1. Configurar registro A en proveedor DNS
2. Esperar propagación (puede tardar hasta 48h)
3. Verificar con: `dig +short tu-dominio.com`

### Error: "Puerto 80 bloqueado"

**Causa:** Firewall o security group bloqueando puerto 80.

**Solución:**
```bash
# Abrir puerto en UFW
sudo ufw allow 80/tcp
sudo ufw reload

# Verificar que Nginx escucha
sudo netstat -tuln | grep ':80'
```

### Error: "Nginx no está instalado"

**Solución:**
```bash
sudo apt update
sudo apt install nginx -y
sudo systemctl start nginx
sudo systemctl enable nginx
```

---

## 📚 Recursos Adicionales

- **Guía completa de SSL**: [`../docs/SSL_SETUP_GUIDE.md`](../docs/SSL_SETUP_GUIDE.md)
- **Ejemplos de configuración Nginx**: [`../docs/nginx-config-examples/`](../docs/nginx-config-examples/)
- **Let's Encrypt**: https://letsencrypt.org/
- **Certbot**: https://certbot.eff.org/

---

## 🤝 Contribuir

Para agregar nuevos scripts de automatización:

1. Crear archivo `.sh` en este directorio
2. Incluir banner y documentación interna
3. Agregar validaciones de prerequisitos
4. Incluir mensajes de error descriptivos
5. Actualizar este README
6. Probar en servidor limpio antes de commit

---

**Desarrollado con 🔧 para simplificar el despliegue**
