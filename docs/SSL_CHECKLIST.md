# ✅ Checklist: Configuración SSL con Let's Encrypt

## 📋 Antes de Comenzar

### Prerequisitos del Servidor
- [ ] VPS con Ubuntu 20.04+ o Debian 10+
- [ ] Acceso SSH con permisos sudo/root
- [ ] Nginx instalado: `sudo apt install nginx -y`
- [ ] Nginx corriendo: `sudo systemctl status nginx`
- [ ] IP pública asignada al servidor

### Configuración DNS (CRÍTICO ⚠️)
- [ ] Dominio registrado y activo
- [ ] Registro A creado: `tu-dominio.com → IP_DEL_SERVIDOR`
- [ ] Registro A creado (www): `www.tu-dominio.com → IP_DEL_SERVIDOR`
- [ ] DNS propagado (verificar con: `dig +short tu-dominio.com`)
- [ ] Dominio resuelve a la IP correcta
- [ ] Tiempo de propagación transcurrido (puede tardar hasta 48h)

### Firewall y Puertos
- [ ] Puerto 80 (HTTP) abierto
- [ ] Puerto 443 (HTTPS) abierto
- [ ] Verificado con: `sudo netstat -tuln | grep ':80\|:443'`
- [ ] UFW configurado: `sudo ufw allow 80/tcp && sudo ufw allow 443/tcp`

### Configuración de Nginx
- [ ] Archivo de configuración creado en `/etc/nginx/sites-available/`
- [ ] `server_name` coincide con tu dominio
- [ ] Link simbólico creado en `/etc/nginx/sites-enabled/`
- [ ] Sintaxis válida: `sudo nginx -t`
- [ ] Nginx recargado: `sudo systemctl reload nginx`

---

## 🚀 Instalación

### Opción A: Script Automatizado (⭐ Recomendado)

- [ ] Script copiado al servidor
- [ ] Permisos de ejecución: `chmod +x setup_ssl.sh`
- [ ] Ejecutado con sudo: `sudo ./setup_ssl.sh tu-dominio.com www.tu-dominio.com`
- [ ] Email proporcionado para notificaciones
- [ ] Script completado sin errores
- [ ] Mensaje de éxito mostrado

### Opción B: Manual

- [ ] Certbot instalado: `sudo apt install certbot python3-certbot-nginx -y`
- [ ] Comando ejecutado: `sudo certbot --nginx -d tu-dominio.com -d www.tu-dominio.com`
- [ ] Términos de servicio aceptados
- [ ] Email registrado (opcional)
- [ ] Opción de redirección HTTPS seleccionada (opción 2)
- [ ] Certificado obtenido exitosamente
- [ ] Nginx recargado automáticamente

---

## ✅ Verificación

### Certificado Instalado
- [ ] Comando exitoso: `sudo certbot certificates`
- [ ] Certificado aparece en lista
- [ ] Fecha de expiración válida (90 días desde hoy)
- [ ] Dominios correctos listados
- [ ] Rutas de certificado correctas

### HTTPS Funcionando
- [ ] Navegador abre: `https://tu-dominio.com`
- [ ] Candadito verde 🔒 visible en barra de direcciones
- [ ] Sin advertencias de seguridad
- [ ] Contenido carga correctamente
- [ ] API funciona si aplica

### Redirección HTTP → HTTPS
- [ ] `http://tu-dominio.com` redirige a `https://`
- [ ] Código de respuesta 301 (Moved Permanently)
- [ ] Verificado con: `curl -I http://tu-dominio.com`
- [ ] Header `Location:` apunta a HTTPS

### Configuración Nginx
- [ ] Archivo modificado por Certbot
- [ ] Bloque HTTP (puerto 80) con redirección
- [ ] Bloque HTTPS (puerto 443) con SSL
- [ ] Certificados referenciados correctamente
- [ ] Sintaxis válida: `sudo nginx -t`
- [ ] Comentarios "# managed by Certbot" presentes

### Auto-renovación
- [ ] Timer activo: `sudo systemctl status certbot.timer`
- [ ] Dry-run exitoso: `sudo certbot renew --dry-run`
- [ ] Sin errores en logs: `sudo journalctl -u certbot.service`
- [ ] Configuración de renovación existe: `/etc/letsencrypt/renewal/tu-dominio.com.conf`

---

## 🔬 Tests de Calidad

### Test con SSL Labs
- [ ] Acceder a: https://www.ssllabs.com/ssltest/
- [ ] Ingresar dominio y analizar
- [ ] Grado obtenido: **A** o **A+** ⭐
- [ ] Sin vulnerabilidades críticas
- [ ] Protocolos modernos (TLS 1.2/1.3)

### Test con cURL
```bash
# Certificado válido
- [ ] curl -vI https://tu-dominio.com 2>&1 | grep "SSL certificate verify ok"

# Redirección HTTP → HTTPS
- [ ] curl -I http://tu-dominio.com | grep "301 Moved"

# Headers de seguridad (si configuraste)
- [ ] curl -I https://tu-dominio.com | grep "Strict-Transport-Security"
```

### Test con OpenSSL
```bash
# Conexión TLS exitosa
- [ ] openssl s_client -connect tu-dominio.com:443 -servername tu-dominio.com < /dev/null 2>/dev/null | grep "Verify return code: 0 (ok)"

# Fechas del certificado
- [ ] echo | openssl s_client -connect tu-dominio.com:443 2>/dev/null | openssl x509 -noout -dates
```

### Test desde Navegadores
- [ ] Chrome/Edge: Candadito verde visible
- [ ] Firefox: Candadito verde visible
- [ ] Safari: Candadito verde visible
- [ ] Mobile (iOS/Android): HTTPS funciona
- [ ] Click en candadito muestra "Conexión segura"

---

## 📊 Monitoreo

### Logs y Alertas
- [ ] Email de Let's Encrypt recibido (confirmación)
- [ ] Logs sin errores: `sudo cat /var/log/letsencrypt/letsencrypt.log`
- [ ] Nginx logs sin advertencias SSL
- [ ] Configurar alerta de vencimiento (30 días antes)

### Documentación
- [ ] IP del servidor documentada
- [ ] Dominios configurados documentados
- [ ] Fecha de instalación registrada
- [ ] Fecha de próxima renovación calculada (90 días)
- [ ] Email de contacto guardado

---

## 🛠️ Post-instalación

### Optimizaciones Opcionales
- [ ] Headers de seguridad agregados (HSTS, X-Frame-Options, etc.)
- [ ] HTTP/2 habilitado (incluido por Certbot)
- [ ] Gzip compression configurado
- [ ] Cache de assets estáticos configurado
- [ ] Rate limiting configurado

### Backups
- [ ] Backup de `/etc/letsencrypt/` creado
- [ ] Backup de configuración Nginx creado
- [ ] Procedimiento de restore documentado
- [ ] Backup programado (opcional)

### Equipo Notificado
- [ ] Team informado del cambio
- [ ] URLs HTTPS actualizadas en documentación
- [ ] Clientes notificados si aplica
- [ ] DNS TTL reducido temporalmente (opcional)

---

## 🚨 En Caso de Problemas

### Si el script falla:
1. [ ] Leer el mensaje de error completo
2. [ ] Verificar logs: `sudo cat /var/log/letsencrypt/letsencrypt.log`
3. [ ] Consultar sección Troubleshooting en `SSL_SETUP_GUIDE.md`
4. [ ] Verificar DNS nuevamente: `dig +short tu-dominio.com`
5. [ ] Esperar 1 hora y reintentar (DNS puede estar propagándose)

### Si la renovación falla:
1. [ ] Verificar timer: `sudo systemctl status certbot.timer`
2. [ ] Intentar renovación manual: `sudo certbot renew`
3. [ ] Verificar logs de error
4. [ ] Asegurar que puerto 80 esté abierto
5. [ ] Contactar soporte si persiste

### Si el sitio no carga:
1. [ ] Verificar sintaxis Nginx: `sudo nginx -t`
2. [ ] Revisar logs de Nginx: `sudo tail -f /var/log/nginx/error.log`
3. [ ] Recargar Nginx: `sudo systemctl reload nginx`
4. [ ] Verificar firewall: `sudo ufw status`
5. [ ] Verificar que el backend esté corriendo (si aplica)

---

## 📅 Mantenimiento Programado

### Semanal
- [ ] Verificar que el sitio cargue con HTTPS
- [ ] Revisar logs de Nginx por errores SSL

### Mensual
- [ ] Verificar estado de renovación: `sudo certbot certificates`
- [ ] Calcular días restantes hasta vencimiento
- [ ] Verificar que timer esté activo: `sudo systemctl status certbot.timer`

### Trimestral (90 días)
- [ ] Certificado renovado automáticamente
- [ ] Email de confirmación recibido
- [ ] Verificar nueva fecha de expiración
- [ ] Test SSL Labs para verificar grado A/A+

### Anual
- [ ] Revisar y actualizar headers de seguridad
- [ ] Verificar cambios en best practices de SSL
- [ ] Actualizar Certbot si hay nueva versión
- [ ] Documentar cambios realizados

---

## 🎯 Criterios de Éxito

Tu instalación SSL está completa cuando:

✅ Todos los checks de "Prerequisitos" están marcados  
✅ Todos los checks de "Instalación" están marcados  
✅ Todos los checks de "Verificación" están marcados  
✅ Test SSL Labs muestra grado **A** o **A+**  
✅ Renovación automática configurada y probada  
✅ Documentación actualizada  

---

## 📞 Contacto y Soporte

Si necesitas ayuda:
- 📖 Guía completa: `docs/SSL_SETUP_GUIDE.md`
- 🐛 Issues: GitHub Issues del proyecto
- 📧 Let's Encrypt Community: https://community.letsencrypt.org/
- 🔧 Certbot Help: https://certbot.eff.org/help/

---

**Fecha de instalación:** ___/___/______  
**Instalado por:** ___________________  
**Próxima renovación:** ___/___/______ (90 días después)  
**Email de contacto:** ___________________

---

✨ **¡Felicidades! Tu sitio ahora tiene el candadito verde 🔒** ✨
