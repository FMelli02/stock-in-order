# Guía de Migración a Supabase

Esta guía te ayudará a migrar tu base de datos PostgreSQL local a Supabase.

## 📋 Requisitos Previos

1. Cuenta en [Supabase](https://supabase.com)
2. Proyecto creado en Supabase
3. Credenciales de conexión de Supabase

## 🚀 Proceso de Migración

### Paso 1: Consolidar Migraciones

Ejecuta el script para consolidar todas las migraciones en un solo archivo:

```powershell
cd scripts
.\consolidate-migrations.ps1
```

Este script creará `supabase-schema.sql` en la raíz del proyecto.

### Paso 2: Crear el Schema en Supabase

1. Ve a tu dashboard de Supabase: https://supabase.com/dashboard
2. Selecciona tu proyecto
3. Ve a **SQL Editor** en el menú lateral
4. Crea una nueva query
5. Copia todo el contenido de `supabase-schema.sql`
6. Pégalo en el editor
7. Haz clic en **Run** para ejecutar el script

### Paso 3: Obtener Credenciales de Supabase

En tu proyecto de Supabase:

1. Ve a **Settings** → **Database**
2. En la sección **Connection Info**, encontrarás:
   - **Host**: `db.xxxxxxxxxxxxx.supabase.co`
   - **Database name**: `postgres`
   - **Port**: `5432`
   - **User**: `postgres`
   - **Password**: [tu password]

3. También puedes copiar el **Connection string** directamente:
   ```
   postgresql://postgres:[YOUR-PASSWORD]@db.xxxxxxxxxxxxx.supabase.co:5432/postgres
   ```

### Paso 4: Configurar Variables de Entorno

Crea o actualiza tu archivo `.env` en la raíz del proyecto:

```env
# Supabase Database
SUPABASE_DB_URL=postgresql://postgres:[YOUR-PASSWORD]@db.xxxxxxxxxxxxx.supabase.co:5432/postgres?sslmode=require

# Supabase API (opcional - si quieres usar la API de Supabase)
SUPABASE_URL=https://xxxxxxxxxxxxx.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_SERVICE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Otras variables existentes
JWT_SECRET=dev-jwt-secret-change-me
ENCRYPTION_KEY=dev-encryption-key-change-me32
RABBITMQ_URL=amqp://user:pass@rabbitmq:5672/
FRONTEND_URL=http://localhost:5173
```

⚠️ **Importante**: Reemplaza `[YOUR-PASSWORD]` y `xxxxxxxxxxxxx` con tus valores reales.

### Paso 5: Actualizar docker-compose.yml

El `docker-compose.yml` ya está actualizado para soportar Supabase. Simplemente:

1. Asegúrate de que el archivo `.env` existe con las variables correctas
2. Reinicia los servicios:

```powershell
docker compose down
docker compose up -d
```

Los servicios `api` y `worker` ahora usarán la variable `SUPABASE_DB_URL` si está configurada, o caerán back a la base de datos local.

### Paso 6: Migrar Datos Existentes (Opcional)

Si tienes datos en tu base de datos local que quieres migrar:

#### Opción A: Dump y Restore (Recomendado)

```powershell
# 1. Exportar datos de la BD local
docker compose exec postgres_db pg_dump -U user -d stock_db --data-only --inserts > backup-data.sql

# 2. Importar a Supabase
# Ve al SQL Editor de Supabase y pega el contenido de backup-data.sql
```

#### Opción B: Script de Migración de Datos

```powershell
# Ejecutar el script de migración de datos
.\scripts\migrate-data-to-supabase.ps1
```

### Paso 7: Verificar la Migración

1. Inicia la aplicación:
   ```powershell
   docker compose up -d
   ```

2. Verifica los logs del backend:
   ```powershell
   docker compose logs api -f
   ```

3. Busca el mensaje: `✅ Conectado a Supabase`

4. Prueba el login en el frontend y verifica que todo funcione correctamente

## 🔧 Configuración Avanzada

### Usar Supabase en Producción

Actualiza las variables de entorno en tu servidor de producción:

```bash
export DB_DSN="postgresql://postgres:[PASSWORD]@db.xxxxx.supabase.co:5432/postgres?sslmode=require"
export ENVIRONMENT="production"
```

### Row Level Security (RLS) en Supabase

Supabase recomienda usar RLS para seguridad. Para habilitarlo:

1. Ve a **Authentication** → **Policies** en Supabase
2. Habilita RLS en las tablas necesarias
3. Crea políticas según tus necesidades

Ejemplo de política para la tabla `products`:

```sql
-- Permitir a los usuarios ver solo sus propios productos
CREATE POLICY "Users can view their own products"
ON products FOR SELECT
USING (user_id = auth.uid());

-- Permitir a los usuarios crear productos
CREATE POLICY "Users can create products"
ON products FOR INSERT
WITH CHECK (user_id = auth.uid());
```

## 🔄 Rollback (Volver a Base de Datos Local)

Si necesitas volver a la base de datos local:

1. Comenta o elimina `SUPABASE_DB_URL` del archivo `.env`
2. Reinicia los servicios:
   ```powershell
   docker compose down
   docker compose up -d
   ```

## 📊 Monitoreo

### Logs de Supabase

Ve a tu proyecto de Supabase → **Logs** para ver:
- Consultas SQL ejecutadas
- Errores de base de datos
- Métricas de rendimiento

### Dashboard de Supabase

El dashboard proporciona:
- **Database**: Gestión de tablas y datos
- **SQL Editor**: Ejecutar queries
- **Table Editor**: Edición visual de datos
- **API Docs**: Documentación auto-generada de tu API

## ⚠️ Consideraciones Importantes

1. **SSL Requerido**: Supabase requiere conexiones SSL. Asegúrate de usar `sslmode=require` en la cadena de conexión.

2. **Límites de Conexiones**: 
   - Plan gratuito: 60 conexiones simultáneas
   - Plan Pro: 200 conexiones simultáneas
   - Configura connection pooling si es necesario

3. **Backups Automáticos**: Supabase hace backups automáticos diarios en el plan gratuito.

4. **Migraciones Futuras**: Para nuevas migraciones, puedes:
   - Ejecutar SQL directamente en el SQL Editor de Supabase
   - Usar herramientas como Flyway o Liquibase
   - Usar las migraciones nativas de Supabase

## 🆘 Troubleshooting

### Error: "connection refused"

- Verifica que la URL de Supabase sea correcta
- Asegúrate de incluir `sslmode=require`
- Verifica que tu IP no esté bloqueada en Supabase

### Error: "password authentication failed"

- Verifica el password en Settings → Database
- Resetea el password si es necesario

### Error: "relation does not exist"

- Asegúrate de haber ejecutado el schema completo
- Verifica en el Table Editor que las tablas existen

## 📚 Recursos Adicionales

- [Documentación de Supabase](https://supabase.com/docs)
- [Supabase Database Guide](https://supabase.com/docs/guides/database)
- [Connection Pooling](https://supabase.com/docs/guides/database/connecting-to-postgres#connection-pooler)

---

¿Problemas? Revisa los logs o contacta al equipo de soporte.
