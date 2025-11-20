# 🚀 Migración a Supabase - Guía Rápida

## ✅ Archivos Creados

1. **`supabase-schema.sql`** - Schema completo consolidado de todas las migraciones
2. **`MIGRACION_SUPABASE.md`** - Guía detallada paso a paso
3. **`scripts/consolidate-migrations.ps1`** - Script para regenerar el schema
4. **`.env.example`** - Actualizado con variables de Supabase

## 📋 Pasos Rápidos

### 1. Crear Proyecto en Supabase

1. Ve a https://supabase.com/dashboard
2. Crea un nuevo proyecto (o usa uno existente)
3. Espera a que el proyecto se inicialice (1-2 minutos)

### 2. Ejecutar el Schema

1. En tu proyecto de Supabase, ve a **SQL Editor**
2. Abre el archivo `supabase-schema.sql` de este proyecto
3. Copia todo el contenido
4. Pégalo en el SQL Editor
5. Haz clic en **Run** (o presiona Ctrl/Cmd + Enter)

### 3. Obtener Credenciales

1. Ve a **Settings** → **Database**
2. Copia el **Connection string** (Connection Pooling)
3. El formato será algo como:
   ```
   postgresql://postgres.xxxxxxxxxxxxx:[YOUR-PASSWORD]@aws-0-us-east-1.pooler.supabase.com:6543/postgres
   ```

### 4. Configurar Aplicación

1. Crea un archivo `.env` en la raíz del proyecto (copia de `.env.example`)
2. Agrega tu connection string:
   ```env
   SUPABASE_DB_URL=postgresql://postgres.xxxxx:[PASSWORD]@aws-0-us-east-1.pooler.supabase.com:6543/postgres
   ```
3. Guarda el archivo

### 5. Reiniciar Aplicación

```powershell
# Reiniciar servicios
docker compose down
docker compose up -d

# Ver logs para verificar la conexión
docker compose logs api -f
```

Busca el mensaje: `✅ Conexión a base de datos establecida`

### 6. Probar

1. Ve a http://localhost:5173
2. Haz login o crea una cuenta nueva
3. Todo debería funcionar igual pero usando Supabase

## 🔄 Migrar Datos Existentes (Opcional)

Si tienes datos en la BD local que quieres migrar:

```powershell
# Exportar datos
docker compose exec postgres_db pg_dump -U user -d stock_db --data-only --inserts > backup-data.sql

# Luego pegar el contenido en el SQL Editor de Supabase
```

## 🔙 Volver a BD Local

Si quieres volver a usar la base de datos local:

1. Comenta la línea `SUPABASE_DB_URL` en `.env`
2. Reinicia: `docker compose restart`

## 📊 Ventajas de Supabase

- ✅ Base de datos en la nube sin configuración
- ✅ Backups automáticos diarios
- ✅ Dashboard visual para gestionar datos
- ✅ API REST auto-generada
- ✅ Autenticación integrada (opcional)
- ✅ Storage de archivos (opcional)
- ✅ Funciones Edge (opcional)

## 📚 Documentación Completa

Lee `MIGRACION_SUPABASE.md` para:
- Configuración avanzada
- Row Level Security (RLS)
- Monitoreo y logs
- Troubleshooting
- Connection pooling

## ⚠️ Notas Importantes

1. **SSL Requerido**: Supabase requiere `sslmode=require` en la conexión
2. **Connection Pooling**: Usa el puerto 6543 para pooling (recomendado)
3. **Límites**: Plan gratuito tiene 500MB de base de datos y 2GB de transferencia
4. **Backups**: Se hacen automáticamente, puedes restaurar desde el dashboard

---

¿Problemas? Revisa `MIGRACION_SUPABASE.md` o los logs con `docker compose logs api`
