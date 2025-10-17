# 🤖 CI/CD Pipeline Documentation

## 📋 Descripción General

Este proyecto implementa un **pipeline completo de CI/CD** usando **GitHub Actions** que automatiza testing, linting, seguridad, construcción y despliegue de imágenes Docker.

---

## 🏗️ Workflows Implementados

### 1. **CI/CD Pipeline** (`ci.yml`)

**Trigger:** Push o Pull Request a `main`/`master`

**Jobs:**

#### 📦 `test-backend` - Testing de Backend (Go)
- ✅ Checkout del código
- ✅ Setup Go 1.23
- ✅ Instalación de dependencias (`go mod download`)
- ✅ Ejecución de tests con race detector (`go test -race`)
- ✅ Generación de coverage report
- ✅ Upload de reporte como artifact

**Comandos ejecutados:**
```bash
go mod download
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 🎨 `test-frontend` - Testing de Frontend (React)
- ✅ Checkout del código
- ✅ Setup Node.js 20
- ✅ Instalación de dependencias (`npm ci`)
- ✅ Ejecución de tests
- ✅ Build de producción
- ✅ Upload de build artifacts

**Comandos ejecutados:**
```bash
npm ci
npm test -- --passWithNoTests --watchAll=false
npm run build
```

#### 🐳 `build-and-push-images` - Docker Build & Push
- ✅ **Dependencias:** Espera que `test-backend` y `test-frontend` pasen
- ✅ Setup Docker Buildx
- ✅ Login a GitHub Container Registry (GHCR)
- ✅ Extracción de metadata (tags, labels)
- ✅ Build y push de imagen Backend
- ✅ Build y push de imagen Frontend
- ✅ Cache de layers para builds más rápidos

**Tags generados:**
- `ghcr.io/[owner]/[repo]/backend:latest` (solo en branch default)
- `ghcr.io/[owner]/[repo]/backend:[branch]-[sha]`
- `ghcr.io/[owner]/[repo]/frontend:latest`
- `ghcr.io/[owner]/[repo]/frontend:[branch]-[sha]`

---

### 2. **PR Quality Checks** (`pr-checks.yml`)

**Trigger:** Pull Request a `main`/`master`

**Jobs:**

#### 🔍 `lint` - Code Quality
- ✅ **Backend:** golangci-lint con múltiples linters
- ✅ **Frontend:** ESLint + TypeScript compiler check
- ✅ Verifica formato, estilo y errores de tipo

**Linters habilitados:**
- `errcheck`, `gosimple`, `govet`, `ineffassign`
- `staticcheck`, `unused`, `typecheck`, `gofmt`, `gocritic`

#### 🔒 `security` - Security Scan
- ✅ Trivy vulnerability scanner
- ✅ Escaneo de filesystem completo
- ✅ Reporte en formato SARIF
- ✅ Upload a GitHub Security tab

#### ✅ `build-check` - Build Verification
- ✅ Verifica que backend compila (`go build`)
- ✅ Verifica que frontend hace build (`npm run build`)
- ✅ Reporte de tamaño del bundle

---

### 3. **Release** (`release.yml`)

**Trigger:** Push de tag con formato `v*.*.*` (ejemplo: `v1.0.0`)

**Jobs:**

#### 📝 `create-release` - GitHub Release
- ✅ Crea release automático en GitHub
- ✅ Incluye instrucciones de despliegue
- ✅ Links a imágenes Docker

#### 🚀 `build-release-images` - Release Images
- ✅ Build y push con tags de versión
- ✅ Tags múltiples: `v1.0.0`, `1.0.0`, `latest`
- ✅ Labels OCI estándar
- ✅ Metadata de versión

---

## 🔐 Configuración de Secrets

### GitHub Secrets Requeridos

Para que el pipeline funcione correctamente, configura los siguientes secrets en tu repositorio:

1. **`GITHUB_TOKEN`** 
   - ✅ **Auto-generado:** GitHub Actions lo provee automáticamente
   - Permisos: `packages: write`, `contents: read`

### Opcional: Docker Hub (en lugar de GHCR)

Si prefieres usar Docker Hub en lugar de GitHub Container Registry:

2. **`DOCKERHUB_USERNAME`**
   - Tu username de Docker Hub
   - Ejemplo: `fmelli02`

3. **`DOCKERHUB_TOKEN`**
   - Access token de Docker Hub
   - Crear en: https://hub.docker.com/settings/security

**Para usar Docker Hub, descomenta estas líneas en `ci.yml`:**
```yaml
- name: Log in to Docker Hub
  uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKERHUB_USERNAME }}
    password: ${{ secrets.DOCKERHUB_TOKEN }}
```

Y cambia:
```yaml
env:
  REGISTRY: docker.io
  IMAGE_NAME_BACKEND: [tu-username]/stock-in-order-backend
  IMAGE_NAME_FRONTEND: [tu-username]/stock-in-order-frontend
```

---

## 📊 Configuración de Permisos en GitHub

### 1. Habilitar GitHub Container Registry

En tu repositorio:
1. Ve a **Settings → Actions → General**
2. En **Workflow permissions**, selecciona:
   - ✅ **Read and write permissions**
3. Marca: ✅ **Allow GitHub Actions to create and approve pull requests**

### 2. Hacer Paquetes Públicos (opcional)

Para que las imágenes sean públicas:
1. Ve a tu perfil → **Packages**
2. Selecciona el paquete
3. **Package settings → Change visibility → Public**

---

## 🚀 Uso del Pipeline

### Push a Main/Master

Cada vez que hagas push a `main` o `master`:

```bash
git add .
git commit -m "feat: nueva funcionalidad"
git push origin main
```

**El pipeline ejecutará:**
1. ✅ Tests de backend (Go)
2. ✅ Tests de frontend (React)
3. ✅ Build de imágenes Docker
4. ✅ Push a GHCR con tags actualizados

### Pull Request

Al crear un PR:

```bash
git checkout -b feature/nueva-feature
git add .
git commit -m "feat: implementar X"
git push origin feature/nueva-feature
```

**El pipeline ejecutará:**
1. ✅ Tests completos
2. ✅ Linting y format checks
3. ✅ Security scan con Trivy
4. ✅ Build verification

### Release

Para crear un release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

**El pipeline ejecutará:**
1. ✅ Crea GitHub Release
2. ✅ Build de imágenes con tag de versión
3. ✅ Push con múltiples tags: `v1.0.0`, `1.0.0`, `latest`

---

## 📦 Pull de Imágenes

### Desde GitHub Container Registry

```bash
# Latest version
docker pull ghcr.io/fmelli02/stock-in-order/backend:latest
docker pull ghcr.io/fmelli02/stock-in-order/frontend:latest

# Versión específica
docker pull ghcr.io/fmelli02/stock-in-order/backend:v1.0.0
docker pull ghcr.io/fmelli02/stock-in-order/frontend:v1.0.0

# Por commit SHA
docker pull ghcr.io/fmelli02/stock-in-order/backend:master-abc1234
```

### Login en GHCR (si las imágenes son privadas)

```bash
# Crear Personal Access Token en GitHub con scope 'read:packages'
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

---

## 🔧 Configuración Local para Testing

### Verificar que los tests pasan localmente

**Backend:**
```bash
cd backend
go test -v -race ./...
```

**Frontend:**
```bash
cd frontend
npm test -- --passWithNoTests --watchAll=false
```

### Ejecutar linting

**Backend:**
```bash
cd backend
golangci-lint run
```

**Frontend:**
```bash
cd frontend
npm run lint
npx tsc --noEmit
```

### Build local de imágenes

```bash
# Backend
docker build -t stock-in-order-backend:local ./backend

# Frontend
docker build -t stock-in-order-frontend:local \
  --build-arg VITE_API_URL=http://localhost:8080/api/v1 \
  ./frontend
```

---

## 📈 Visualización de Resultados

### GitHub Actions Tab

1. Ve a tu repositorio en GitHub
2. Click en la pestaña **Actions**
3. Verás todos los workflow runs

### Artifacts

Los siguientes artifacts se generan y están disponibles por 30 días:

- **`backend-coverage`**: Reporte HTML de cobertura de tests
- **`frontend-build`**: Build de producción del frontend (7 días)

Para descargar:
1. Ve a **Actions** → selecciona un workflow run
2. Scroll down a **Artifacts**
3. Click para descargar

### Security Tab

Los resultados de Trivy se suben automáticamente a:
- **Security** → **Code scanning alerts**

---

## 🐛 Troubleshooting

### Error: "Permission denied while pushing to GHCR"

**Solución:**
1. Verifica que los permisos de workflow estén configurados correctamente
2. Ve a Settings → Actions → General
3. Marca "Read and write permissions"

### Error: "go test fails with module not found"

**Solución:**
```bash
cd backend
go mod tidy
git add go.mod go.sum
git commit -m "chore: update dependencies"
```

### Error: "npm test fails"

**Solución:**
```bash
cd frontend
npm install
npm run build
# Verifica que los tests pasen localmente
npm test
```

### Docker build es muy lento

**Optimización:**
- El pipeline usa cache de layers automáticamente
- En builds subsecuentes, solo se rebuildan layers modificados
- Primer build: ~3-5 minutos
- Builds incrementales: ~30-60 segundos

---

## 🎯 Best Practices

### Commits

Use **Conventional Commits** para mensajes claros:
```
feat: agregar nueva funcionalidad
fix: corregir bug en login
chore: actualizar dependencias
docs: actualizar README
test: agregar tests para productos
refactor: reorganizar código de handlers
```

### Branches

- `main`/`master`: Código de producción
- `develop`: Desarrollo activo (opcional)
- `feature/*`: Nuevas funcionalidades
- `bugfix/*`: Corrección de bugs
- `hotfix/*`: Correcciones urgentes

### Tags de Versión

Usa **Semantic Versioning**:
- `v1.0.0`: Major release (breaking changes)
- `v1.1.0`: Minor release (nuevas features)
- `v1.1.1`: Patch release (bug fixes)

---

## 📚 Recursos Adicionales

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [golangci-lint](https://golangci-lint.run/)
- [Trivy Security Scanner](https://trivy.dev/)

---

## 🎉 Estado del Pipeline

✅ **CI/CD Pipeline** configurado y funcional  
✅ **Testing automático** de backend y frontend  
✅ **Linting y calidad** de código  
✅ **Security scanning** con Trivy  
✅ **Docker images** auto-build y push  
✅ **Releases automáticos** con tags  
✅ **Cache optimizado** para builds rápidos  

**¡El robot obrero está listo para trabajar! 🤖**
