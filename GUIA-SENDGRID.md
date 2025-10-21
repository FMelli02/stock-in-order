# 📧 Guía de Configuración de SendGrid

## ¿Qué es SendGrid?

**SendGrid** es un servicio de envío de emails transaccionales que utilizamos para enviar los reportes generados a los usuarios.

## ¿Por qué SendGrid?

- ✅ **Confiable**: Alta tasa de entrega (deliverability)
- ✅ **Escalable**: Miles de emails por minuto
- ✅ **Gratuito**: 100 emails/día gratis para siempre
- ✅ **Fácil**: API simple y bien documentada
- ✅ **Profesional**: No cae en spam, tracking incluido

## Plan Gratuito

SendGrid ofrece un plan gratuito con:
- 📧 **100 emails por día** (suficiente para desarrollo y proyectos pequeños)
- ✅ Email validation
- ✅ Templates HTML
- ✅ Analytics básico
- ✅ Sin tarjeta de crédito requerida

## 🚀 Paso a Paso: Configuración

### Paso 1: Crear Cuenta en SendGrid

1. Ve a [https://signup.sendgrid.com/](https://signup.sendgrid.com/)
2. Completa el formulario de registro:
   - Email
   - Contraseña
   - Nombre completo
3. Verifica tu email (revisa spam si no llega)
4. Completa el onboarding:
   - **Get Started with Email API** (opción recomendada)
   - Elige **Web API** como método de integración
   - Selecciona **Go** como lenguaje

### Paso 2: Crear API Key

1. Una vez dentro del dashboard, ve a:
   ```
   Settings → API Keys
   ```

2. Haz clic en **"Create API Key"**

3. Configura la API Key:
   - **API Key Name**: `stock-in-order-worker`
   - **API Key Permissions**: 
     - Selecciona **"Restricted Access"**
     - En la lista, expande **"Mail Send"**
     - Activa **"Mail Send"** (debe estar en ON)
     - Deja todo lo demás en OFF

4. Haz clic en **"Create & View"**

5. **⚠️ IMPORTANTE**: Copia la API Key inmediatamente
   ```
   SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   ```
   - Esta es la **única vez** que podrás verla
   - Guárdala en un lugar seguro (nunca en Git)

### Paso 3: Verificar Sender Identity

Para poder enviar emails, SendGrid requiere que verifiques tu identidad:

#### Opción A: Single Sender Verification (Más rápido - Recomendado para desarrollo)

1. Ve a:
   ```
   Settings → Sender Authentication → Single Sender Verification
   ```

2. Haz clic en **"Create New Sender"**

3. Completa el formulario:
   - **From Name**: `Stock in Order`
   - **From Email Address**: Tu email personal (ej: `tu@gmail.com`)
   - **Reply To**: El mismo email
   - **Company Address**: Tu dirección
   - **City**, **State**, **Zip Code**, **Country**

4. Haz clic en **"Create"**

5. **Verifica tu email**:
   - SendGrid te enviará un email de verificación
   - Haz clic en el link de verificación
   - Una vez verificado, podrás enviar emails desde ese correo

#### Opción B: Domain Authentication (Profesional - Requiere dominio propio)

Si tienes un dominio (ej: `tuempresa.com`):

1. Ve a:
   ```
   Settings → Sender Authentication → Domain Authentication
   ```

2. Haz clic en **"Authenticate Your Domain"**

3. Selecciona tu proveedor DNS (ej: Cloudflare, GoDaddy, etc.)

4. Ingresa tu dominio: `tuempresa.com`

5. Copia los registros DNS que SendGrid te proporciona

6. Agrégalos en tu proveedor de DNS:
   - Registros **CNAME** para validación
   - Registros **TXT** para SPF/DKIM

7. Espera hasta 48 horas para propagación DNS

8. Verifica en SendGrid que el dominio esté autenticado

### Paso 4: Configurar en el Proyecto

1. **Crea un archivo `.env`** en la raíz del proyecto:
   ```bash
   cp .env.example .env
   ```

2. **Edita el archivo `.env`** y agrega tu API Key:
   ```bash
   SENDGRID_API_KEY=SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   ```

3. **⚠️ NUNCA** subas el archivo `.env` a Git:
   - Ya está incluido en `.gitignore`
   - Verifica que esté ignorado

### Paso 5: Actualizar Email Remitente en el Código

Si usaste **Single Sender Verification**, debes actualizar el email remitente en el código:

1. Abre `worker/cmd/api/main.go`

2. Busca esta línea (aproximadamente línea 43):
   ```go
   emailClient := email.NewClient(
       cfg.SendGrid_APIKey,
       "noreply@stockinorder.com", // ← Cambiar este email
       "Stock in Order",
   )
   ```

3. Reemplázalo con el email que verificaste:
   ```go
   emailClient := email.NewClient(
       cfg.SendGrid_APIKey,
       "tu@gmail.com", // ← Tu email verificado
       "Stock in Order",
   )
   ```

### Paso 6: Probar el Envío

1. **Reconstruye los contenedores**:
   ```bash
   docker compose down
   docker compose up -d --build
   ```

2. **Verifica los logs del worker**:
   ```bash
   docker logs stock_in_order_worker -f
   ```

   Deberías ver:
   ```
   📧 Cliente de email configurado
   ```

3. **Prueba el envío**:
   - Ve al frontend: http://localhost:5173
   - Inicia sesión
   - Ve a "Productos"
   - Haz clic en **"Recibir por Email"**

4. **Monitorea los logs**:
   ```bash
   docker logs stock_in_order_worker --tail 20
   ```

   Deberías ver:
   ```
   📨 Mensaje recibido
   🔨 Generando reporte
   📊 Reporte generado: 6403 bytes
   ✅ Email enviado exitosamente a usuario@ejemplo.com (código: 202)
   ```

5. **Revisa tu bandeja de entrada**:
   - El email debería llegar en **menos de 30 segundos**
   - Si no llega, revisa **spam/correo no deseado**

## 🧪 Modo Desarrollo (Sin SendGrid)

Si **NO** configuras `SENDGRID_API_KEY`, el sistema funcionará en **modo desarrollo**:

- ✅ El worker procesa reportes normalmente
- ✅ Genera los archivos Excel
- ⚠️ **NO** envía emails reales
- 📋 Solo logea: `[MODO DEV] Email simulado a usuario@ejemplo.com`

Esto es útil para:
- Desarrollo local sin cuenta SendGrid
- Testing de la lógica sin consumir cuota
- Ambientes de CI/CD

## 📊 Monitoreo en SendGrid

Una vez configurado, puedes monitorear tus emails en el dashboard:

1. Ve a **Activity** en SendGrid

2. Verás cada email enviado con:
   - ✅ Entregado (Delivered)
   - 📬 Abierto (Opened)
   - 🖱️ Clics (Clicked)
   - ❌ Rebotado (Bounced)
   - 🚫 Spam

3. Si un email falla:
   - Revisa **Activity Feed**
   - Busca el email problemático
   - Lee el mensaje de error
   - Corrige el problema

## ⚠️ Límites del Plan Gratuito

- 📧 **100 emails por día**
- 🔄 Se resetea cada 24 horas
- 💳 Sin tarjeta requerida

Si necesitas más:
- **Essentials**: $19.95/mes (50,000 emails/mes)
- **Pro**: $89.95/mes (1,500,000 emails/mes)

## 🔐 Seguridad

### ✅ Buenas Prácticas

1. **Nunca** subas tu API Key a Git
2. **Rota** la API Key cada 90 días
3. **Usa** Restricted Access (solo Mail Send)
4. **Elimina** API Keys que no uses
5. **Monitorea** el uso en SendGrid Dashboard

### ❌ Lo que NO debes hacer

- ❌ Compartir tu API Key
- ❌ Hardcodear la API Key en el código
- ❌ Usar Full Access si no lo necesitas
- ❌ Dejar API Keys sin usar activas

## 🐛 Troubleshooting

### Problema: "Error 401 Unauthorized"

**Causa**: API Key inválida o sin permisos

**Solución**:
1. Verifica que copiaste la API Key completa
2. Verifica que tiene permisos de "Mail Send"
3. Genera una nueva API Key

### Problema: "Error 403 Forbidden"

**Causa**: Email remitente no verificado

**Solución**:
1. Completa Single Sender Verification
2. Verifica tu email haciendo clic en el link
3. Usa el email verificado en el código

### Problema: "Email no llega"

**Causa**: Email en spam o filtrado

**Solución**:
1. Revisa tu carpeta de spam
2. Verifica en SendGrid Activity si se envió
3. Si se envió pero no llega, contacta a tu proveedor de email
4. Considera usar Domain Authentication

### Problema: "Límite de 100 emails superado"

**Causa**: Cuota diaria agotada

**Solución**:
1. Espera 24 horas para reset
2. Considera upgrade a plan pago
3. En desarrollo, usa modo sin API Key

## 📚 Recursos Adicionales

- 📖 [Documentación oficial SendGrid](https://docs.sendgrid.com/)
- 🔑 [API Keys Guide](https://docs.sendgrid.com/ui/account-and-settings/api-keys)
- 📧 [Email Authentication](https://docs.sendgrid.com/ui/account-and-settings/how-to-set-up-domain-authentication)
- 💬 [SendGrid Support](https://support.sendgrid.com/)
- 🎓 [Tutorials y ejemplos](https://docs.sendgrid.com/for-developers)

## ✅ Checklist de Configuración

- [ ] Cuenta de SendGrid creada
- [ ] API Key generada con permisos "Mail Send"
- [ ] Single Sender Verification completada
- [ ] Email remitente verificado
- [ ] API Key agregada a archivo `.env`
- [ ] Email remitente actualizado en `main.go`
- [ ] Contenedores reconstruidos con `docker compose up -d --build`
- [ ] Prueba enviada desde el frontend
- [ ] Email recibido correctamente
- [ ] Logs del worker verificados

---

**¡Listo!** 🎉 Ahora tu sistema está configurado para enviar reportes por email.
