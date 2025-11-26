# 🔒 Guía Completa de SSL con Let's Encrypt y Certbot

## 📋 Tabla de Contenidos

1. [Introducción](#introducción)
2. [Prerequisitos](#prerequisitos)
3. [Configuración DNS](#configuración-dns)
4. [Instalación Automatizada](#instalación-automatizada)
5. [Instalación Manual](#instalación-manual)
6. [Cómo Funciona Certbot](#cómo-funciona-certbot)
7. [Modificaciones en Nginx](#modificaciones-en-nginx)
8. [Auto-renovación](#auto-renovación)
9. [Verificación](#verificación)
10. [Troubleshooting](#troubleshooting)
11. [Comandos Útiles](#comandos-útiles)

---

## 🎯 Introducción

Esta guía te permite obtener un **certificado SSL gratuito** de Let's Encrypt y configurarlo automáticamente en Nginx para que tu sitio tenga **HTTPS con el candadito verde 🔒**.

### ¿Por qué HTTPS?

- ✅ **Seguridad**: Encripta la comunicación entre cliente y servidor
- ✅ **Confianza**: Los navegadores muestran el candadito verde
- ✅ **SEO**: Google prioriza sitios HTTPS
- ✅ **Requisito moderno**: Muchas APIs solo funcionan con HTTPS
- ✅ **Gratis**: Let's Encrypt es 100% gratuito

### Características de Let's Encrypt

- 🆓 **Completamente gratuito**
- ⏱️ Certificados válidos por **90 días**
- 🔄 **Auto-renovación automática** cada 60 días
- 🌍 Reconocido por todos los navegadores
- 🤖 Proceso automatizado con Certbot

---

## ⚙️ Prerequisitos

### 1. Servidor VPS

Necesitas un servidor con:

| Requisito | Especificación |
|-----------|----------------|
| **OS** | Ubuntu 20.04+ / Debian 10+ |
| **Acceso** | SSH con permisos sudo/root |
| **Web Server** | Nginx instalado y corriendo |
| **IP Pública** | Asignada al servidor |

### 2. Dominio Configurado

**CRÍTICO**: El dominio **DEBE** apuntar a la IP del servidor **ANTES** de ejecutar Certbot.

#### Configuración DNS Necesaria

En tu proveedor de DNS (GoDaddy, Namecheap, Cloudflare, etc.):

```
Tipo    Nombre              Valor              TTL
A       @                   123.45.67.89       3600
A       www                 123.45.67.89       3600
```

Donde `123.45.67.89` es la IP de tu VPS.

#### Verificar DNS

```bash
# Obtener IP del servidor
curl ifconfig.me

# Verificar que el dominio apunta a esa IP
dig +short tu-dominio.com
dig +short www.tu-dominio.com

# Deben retornar la misma IP del servidor
```

**⏰ Importante**: Los cambios DNS pueden tardar hasta 48 horas en propagarse.

### 3. Firewall Configurado

Los puertos **80** (HTTP) y **443** (HTTPS) deben estar abiertos:

```bash
# Ubuntu con UFW
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw status

# Verificar que Nginx esté escuchando
sudo netstat -tuln | grep ':80'
sudo netstat -tuln | grep ':443'
```

### 4. Nginx con Configuración Base

Debe existir un archivo de configuración en:
- `/etc/nginx/sites-available/tu-dominio.com`
- O al menos `/etc/nginx/sites-available/default`

**Ejemplo mínimo:**

```nginx
server {
    listen 80;
    listen [::]:80;
    
    server_name tu-dominio.com www.tu-dominio.com;
    
    root /var/www/html;
    index index.html;
    
    location / {
        try_files $uri $uri/ =404;
    }
}
```

---

## 🚀 Instalación Automatizada

### Opción 1: Script Automatizado (⭐ Recomendado)

```bash
# 1. Descargar el script
cd /tmp
wget https://tu-repo.com/setup_ssl.sh
# O copiar manualmente el script a tu servidor

# 2. Dar permisos de ejecución
chmod +x setup_ssl.sh

# 3. Ejecutar con tu dominio
sudo ./setup_ssl.sh tu-dominio.com www.tu-dominio.com

# Ejemplo real:
sudo ./setup_ssl.sh stock-in-order.com www.stock-in-order.com
```

**El script hará automáticamente:**
1. ✅ Validar prerequisitos (Nginx, puertos, DNS)
2. ✅ Instalar Certbot y dependencias
3. ✅ Obtener certificado SSL de Let's Encrypt
4. ✅ Configurar Nginx para HTTPS
5. ✅ Activar redirección HTTP → HTTPS
6. ✅ Configurar auto-renovación
7. ✅ Verificar que todo funcione

### Output Esperado

```
╔════════════════════════════════════════════════╗
║   🔒 CONFIGURACIÓN SSL CON LET'S ENCRYPT       ║
║   Certbot + Nginx + Auto-renovación            ║
╚════════════════════════════════════════════════╝

[1/6] Validando prerequisitos...
✓ Dominio principal: stock-in-order.com
✓ Nginx está instalado
✓ Nginx está corriendo
✓ Puerto 80 (HTTP) está abierto

[2/6] Verificando configuración DNS...
IP del servidor: 123.45.67.89
IP del dominio: 123.45.67.89
✓ DNS configurado correctamente

[3/6] Actualizando sistema e instalando Certbot...
✓ Certbot instalado: certbot 2.7.4

[4/6] Verificando configuración de Nginx...
✓ Configuración de Nginx es válida

[5/6] Ejecutando Certbot para obtener certificado SSL...
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/stock-in-order.com/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/stock-in-order.com/privkey.pem

╔════════════════════════════════════════════════╗
║   ✅ CERTIFICADO SSL INSTALADO EXITOSAMENTE    ║
╚════════════════════════════════════════════════╝

[6/6] Configurando auto-renovación de certificados...
✓ Timer de renovación automática está activo
✓ La renovación automática funcionará correctamente

╔════════════════════════════════════════════════╗
║          🎉 ¡CONFIGURACIÓN COMPLETADA! 🎉      ║
╚════════════════════════════════════════════════╝

✓ Certificado SSL instalado para: stock-in-order.com
✓ Redirección automática HTTP → HTTPS configurada
✓ Auto-renovación activada

🌐 Tu sitio ahora está accesible en:
   https://stock-in-order.com
   https://www.stock-in-order.com

✨ ¡Tu sitio ahora tiene el candadito verde! 🔒✨
```

---

## 🛠️ Instalación Manual

Si prefieres hacerlo paso a paso sin el script:

### Paso 1: Instalar Certbot

```bash
# Actualizar sistema
sudo apt update

# Instalar Certbot con plugin de Nginx
sudo apt install certbot python3-certbot-nginx -y

# Verificar instalación
certbot --version
```

### Paso 2: Obtener Certificado

```bash
# Comando básico (con dominios)
sudo certbot --nginx -d tu-dominio.com -d www.tu-dominio.com

# Con email para notificaciones
sudo certbot --nginx -d tu-dominio.com -d www.tu-dominio.com --email admin@tu-dominio.com

# Modo no interactivo (para automatización)
sudo certbot --nginx -d tu-dominio.com -d www.tu-dominio.com \
  --non-interactive \
  --agree-tos \
  --email admin@tu-dominio.com \
  --redirect
```

### Paso 3: Responder Preguntas Interactivas

Durante la ejecución, Certbot preguntará:

1. **Email de contacto** (para notificaciones de vencimiento)
   - Ingresa tu email o deja vacío
   
2. **Aceptar términos de servicio**
   - Escribe `A` para aceptar
   
3. **Compartir email con EFF** (opcional)
   - Escribe `Y` (sí) o `N` (no)
   
4. **Redirección HTTP → HTTPS**
   - Opción `2`: Redirigir todo el tráfico a HTTPS (⭐ recomendado)

### Paso 4: Verificar Auto-renovación

```bash
# Ver estado del timer
sudo systemctl status certbot.timer

# Habilitar si no está activo
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Probar renovación (sin hacerla realmente)
sudo certbot renew --dry-run
```

---

## 🔍 Cómo Funciona Certbot

### Proceso de Validación (ACME Challenge)

Certbot usa el protocolo **ACME** para validar que controlas el dominio:

```
1. Certbot crea un archivo temporal en:
   /var/www/html/.well-known/acme-challenge/TOKEN

2. Contacta a Let's Encrypt con tu dominio

3. Let's Encrypt hace una petición HTTP a:
   http://tu-dominio.com/.well-known/acme-challenge/TOKEN

4. Si obtiene el archivo correcto, valida que controlas el dominio

5. Let's Encrypt genera el certificado y lo entrega a Certbot

6. Certbot instala el certificado en Nginx automáticamente
```

### Estructura de Archivos

Certbot almacena los certificados en:

```
/etc/letsencrypt/
├── live/
│   └── tu-dominio.com/
│       ├── fullchain.pem    → Certificado completo (público)
│       ├── privkey.pem      → Clave privada (secreto)
│       ├── cert.pem         → Certificado del dominio
│       └── chain.pem        → Certificados intermedios
│
├── archive/                  → Versiones históricas
│   └── tu-dominio.com/
│
├── renewal/                  → Configuración de renovación
│   └── tu-dominio.com.conf
│
└── accounts/                 → Credenciales de Let's Encrypt
```

**⚠️ IMPORTANTE**: `privkey.pem` es **MUY SENSIBLE**. Nunca lo compartas ni lo subas a git.

---

## 📝 Modificaciones en Nginx

### Antes de Certbot

Tu configuración original:

```nginx
server {
    listen 80;
    listen [::]:80;
    
    server_name tu-dominio.com www.tu-dominio.com;
    
    root /var/www/html;
    index index.html;
    
    location / {
        try_files $uri $uri/ =404;
    }
}
```

### Después de Certbot

Certbot **modifica automáticamente** el archivo y crea:

```nginx
# Bloque HTTP (Puerto 80) - Redirige a HTTPS
server {
    listen 80;
    listen [::]:80;
    
    server_name tu-dominio.com www.tu-dominio.com;
    
    # Certbot agrega esta redirección automáticamente
    return 301 https://$server_name$request_uri;
}

# Bloque HTTPS (Puerto 443) - Configuración SSL
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    
    server_name tu-dominio.com www.tu-dominio.com;
    
    root /var/www/html;
    index index.html;
    
    # Certificados SSL - Agregados por Certbot
    ssl_certificate /etc/letsencrypt/live/tu-dominio.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/tu-dominio.com/privkey.pem;
    
    # Opciones SSL recomendadas - Agregadas por Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
    
    location / {
        try_files $uri $uri/ =404;
    }
}
```

### Opciones SSL Configuradas

Certbot crea `/etc/letsencrypt/options-ssl-nginx.conf` con:

```nginx
# Protocolos SSL modernos
ssl_protocols TLSv1.2 TLSv1.3;

# Cifrados seguros
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256...';
ssl_prefer_server_ciphers off;

# Timeouts
ssl_session_timeout 1d;
ssl_session_cache shared:SSL:50m;
ssl_session_tickets off;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
ssl_trusted_certificate /etc/letsencrypt/live/tu-dominio.com/chain.pem;

# Resolvers DNS
resolver 8.8.8.8 8.8.4.4 valid=300s;
resolver_timeout 5s;
```

### Parámetros Diffie-Hellman

Certbot también genera `/etc/letsencrypt/ssl-dhparams.pem` para **Perfect Forward Secrecy**.

---

## 🔄 Auto-renovación

### Cómo Funciona

Los certificados de Let's Encrypt **vencen a los 90 días**. Certbot instala un **systemd timer** que intenta renovarlos automáticamente cada 12 horas.

### Verificar Timer

```bash
# Ver estado
sudo systemctl status certbot.timer

# Output esperado:
● certbot.timer - Run certbot twice daily
     Loaded: loaded (/lib/systemd/system/certbot.timer; enabled)
     Active: active (waiting)
    Trigger: Mon 2025-11-25 12:00:00 UTC; 11h left
```

### Proceso de Renovación

```
1. El timer ejecuta: certbot renew

2. Certbot verifica cada certificado:
   - Si faltan > 30 días: No hace nada
   - Si faltan ≤ 30 días: Renueva el certificado

3. Después de renovar, recarga Nginx automáticamente
```

### Probar Renovación Manualmente

```bash
# Dry-run (simula renovación sin hacerla)
sudo certbot renew --dry-run

# Forzar renovación (aunque no esté por vencer)
sudo certbot renew --force-renewal

# Renovación silenciosa (sin output)
sudo certbot renew --quiet
```

### Logs de Renovación

```bash
# Ver logs de Certbot
sudo cat /var/log/letsencrypt/letsencrypt.log

# Ver últimos intentos de renovación
sudo journalctl -u certbot.service
```

### Configuración de Renovación

Archivo: `/etc/letsencrypt/renewal/tu-dominio.com.conf`

```ini
# renew_before_expiry = 30 days
version = 2.7.4
archive_dir = /etc/letsencrypt/archive/tu-dominio.com
cert = /etc/letsencrypt/live/tu-dominio.com/cert.pem
privkey = /etc/letsencrypt/live/tu-dominio.com/privkey.pem
chain = /etc/letsencrypt/live/tu-dominio.com/chain.pem
fullchain = /etc/letsencrypt/live/tu-dominio.com/fullchain.pem

[renewalparams]
account = ACCOUNT_ID
authenticator = nginx
installer = nginx
server = https://acme-v02.api.letsencrypt.org/directory
```

---

## ✅ Verificación

### 1. Verificar en Navegador

Abre: `https://tu-dominio.com`

**Deberías ver:**
- 🔒 Candadito verde en la barra de direcciones
- Certificado emitido por "Let's Encrypt Authority"
- Conexión segura (TLS 1.2 o 1.3)

### 2. Verificar con SSL Labs

Herramienta online: https://www.ssllabs.com/ssltest/

```
Ingresa tu dominio y espera el análisis

Resultado esperado: Grado A o A+
```

### 3. Verificar con cURL

```bash
# Verificar certificado
curl -vI https://tu-dominio.com 2>&1 | grep -E "SSL|TLS|certificate"

# Verificar redirección HTTP → HTTPS
curl -I http://tu-dominio.com
# Debe retornar: 301 Moved Permanently
# Location: https://tu-dominio.com
```

### 4. Verificar con OpenSSL

```bash
# Ver detalles del certificado
openssl s_client -connect tu-dominio.com:443 -servername tu-dominio.com < /dev/null 2>/dev/null | openssl x509 -noout -text

# Ver fechas de validez
echo | openssl s_client -connect tu-dominio.com:443 2>/dev/null | openssl x509 -noout -dates
```

### 5. Ver Certificados Instalados

```bash
# Listar todos los certificados
sudo certbot certificates

# Output esperado:
Found the following certs:
  Certificate Name: tu-dominio.com
    Domains: tu-dominio.com www.tu-dominio.com
    Expiry Date: 2026-02-22 12:00:00+00:00 (VALID: 89 days)
    Certificate Path: /etc/letsencrypt/live/tu-dominio.com/fullchain.pem
    Private Key Path: /etc/letsencrypt/live/tu-dominio.com/privkey.pem
```

---

## 🚨 Troubleshooting

### Error 1: DNS no apunta al servidor

**Síntoma:**
```
Failed authorization procedure. tu-dominio.com (http-01): 
urn:ietf:params:acme:error:dns :: DNS problem: NXDOMAIN
```

**Causa:** El dominio no está apuntando a la IP del servidor.

**Solución:**
```bash
# Verificar DNS
dig +short tu-dominio.com

# Debe retornar la IP del servidor
# Si no, configura el registro A en tu proveedor DNS
```

**Esperar Propagación DNS:**
```bash
# Verificar en múltiples DNS públicos
dig @8.8.8.8 tu-dominio.com
dig @1.1.1.1 tu-dominio.com
dig @208.67.222.222 tu-dominio.com
```

---

### Error 2: Puerto 80 bloqueado

**Síntoma:**
```
Failed authorization procedure. tu-dominio.com (http-01): 
urn:ietf:params:acme:error:connection :: Connection refused
```

**Causa:** El firewall bloquea el puerto 80 o Nginx no está escuchando.

**Solución:**
```bash
# Verificar que Nginx escucha en puerto 80
sudo netstat -tuln | grep ':80'

# Abrir puerto 80 en firewall
sudo ufw allow 80/tcp
sudo ufw reload

# Verificar que Nginx esté corriendo
sudo systemctl status nginx

# Reiniciar Nginx si es necesario
sudo systemctl restart nginx
```

---

### Error 3: Nginx mal configurado

**Síntoma:**
```
nginx: [emerg] invalid number of arguments in "ssl_certificate"
```

**Causa:** Configuración de Nginx con errores de sintaxis.

**Solución:**
```bash
# Verificar sintaxis
sudo nginx -t

# Ver el error específico y corregir el archivo
sudo nano /etc/nginx/sites-available/tu-dominio.com

# Después de corregir, recargar
sudo systemctl reload nginx
```

---

### Error 4: Rate Limit de Let's Encrypt

**Síntoma:**
```
Error: urn:ietf:params:acme:error:rateLimited
Too many certificates already issued for this exact set of domains
```

**Causa:** Demasiados intentos en poco tiempo.

**Límites de Let's Encrypt:**
- 5 certificados por dominio por semana
- 50 subdominios por dominio registrado
- 300 cuentas nuevas por IP cada 3 horas

**Solución:**
```bash
# Esperar 1 semana y volver a intentar

# Mientras tanto, usar el entorno de staging
sudo certbot --staging -d tu-dominio.com
```

---

### Error 5: Certificado no se renueva automáticamente

**Síntoma:**
```bash
sudo certbot renew --dry-run
# Retorna errores
```

**Solución:**
```bash
# Verificar timer
sudo systemctl status certbot.timer

# Si no está activo, habilitarlo
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Ver logs de errores
sudo journalctl -u certbot.service -n 50

# Renovar manualmente
sudo certbot renew --force-renewal
```

---

### Error 6: Nginx no recarga después de renovar

**Síntoma:** Certificado renovado pero Nginx sigue usando el antiguo.

**Solución:**
```bash
# Recargar Nginx manualmente
sudo systemctl reload nginx

# Verificar configuración de renovación
sudo cat /etc/letsencrypt/renewal/tu-dominio.com.conf

# Asegurar que contenga:
post_hook = systemctl reload nginx
```

---

## 🔧 Comandos Útiles

### Gestión de Certificados

```bash
# Listar certificados instalados
sudo certbot certificates

# Renovar certificados manualmente
sudo certbot renew

# Renovar certificado específico
sudo certbot renew --cert-name tu-dominio.com

# Probar renovación (dry-run)
sudo certbot renew --dry-run

# Forzar renovación (aunque no esté por vencer)
sudo certbot renew --force-renewal

# Revocar certificado
sudo certbot revoke --cert-path /etc/letsencrypt/live/tu-dominio.com/cert.pem

# Eliminar certificado
sudo certbot delete --cert-name tu-dominio.com
```

### Información del Certificado

```bash
# Ver información completa
sudo certbot certificates

# Ver fecha de expiración
sudo openssl x509 -noout -dates -in /etc/letsencrypt/live/tu-dominio.com/cert.pem

# Ver dominios cubiertos
sudo openssl x509 -noout -text -in /etc/letsencrypt/live/tu-dominio.com/cert.pem | grep DNS

# Ver emisor del certificado
sudo openssl x509 -noout -issuer -in /etc/letsencrypt/live/tu-dominio.com/cert.pem
```

### Logs y Debugging

```bash
# Ver logs de Certbot
sudo cat /var/log/letsencrypt/letsencrypt.log

# Seguir logs en tiempo real
sudo tail -f /var/log/letsencrypt/letsencrypt.log

# Ver logs del servicio de renovación
sudo journalctl -u certbot.service

# Ver últimas 50 líneas de logs
sudo journalctl -u certbot.service -n 50

# Ver logs desde una fecha
sudo journalctl -u certbot.service --since "2025-11-01"
```

### Testing

```bash
# Test SSL con SSL Labs (navegador)
# https://www.ssllabs.com/ssltest/analyze.html?d=tu-dominio.com

# Test con cURL
curl -vI https://tu-dominio.com

# Test con OpenSSL
echo | openssl s_client -connect tu-dominio.com:443 -servername tu-dominio.com

# Verificar redirección HTTP → HTTPS
curl -I http://tu-dominio.com
```

---

## 📊 Checklist de Verificación

Antes de ejecutar Certbot, asegúrate de que:

- [ ] El dominio existe y está registrado
- [ ] El registro A apunta a la IP del servidor
- [ ] DNS propagado (puede tardar hasta 48h)
- [ ] Servidor VPS con Ubuntu/Debian
- [ ] Nginx instalado y corriendo
- [ ] Puerto 80 abierto en firewall
- [ ] Puerto 443 abierto en firewall
- [ ] Configuración base de Nginx creada
- [ ] `server_name` en Nginx coincide con el dominio
- [ ] Nginx reiniciado después de cambios

Después de instalar SSL:

- [ ] Navegador muestra candadito verde 🔒
- [ ] `https://tu-dominio.com` funciona
- [ ] Redirección HTTP → HTTPS activa
- [ ] Timer de renovación activo (`systemctl status certbot.timer`)
- [ ] Dry-run de renovación exitoso (`certbot renew --dry-run`)
- [ ] Certificado válido por 90 días
- [ ] Email de confirmación recibido

---

## 🎯 Resumen Ejecutivo

### Instalación en 3 Comandos

```bash
# 1. Instalar Certbot
sudo apt install certbot python3-certbot-nginx -y

# 2. Obtener certificado
sudo certbot --nginx -d tu-dominio.com -d www.tu-dominio.com

# 3. Verificar auto-renovación
sudo certbot renew --dry-run
```

### Resultado

- ✅ Certificado SSL instalado
- ✅ HTTPS funcionando
- ✅ Redirección automática
- ✅ Auto-renovación configurada
- ✅ Válido por 90 días
- ✅ Grado A/A+ en SSL Labs

---

## 🔗 Referencias

- **Let's Encrypt**: https://letsencrypt.org/
- **Certbot**: https://certbot.eff.org/
- **SSL Labs**: https://www.ssllabs.com/ssltest/
- **Mozilla SSL Config**: https://ssl-config.mozilla.org/
- **Nginx Docs**: https://nginx.org/en/docs/http/configuring_https_servers.html

---

**Desarrollado con 🔒 para hacer internet más seguro**
