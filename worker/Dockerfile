# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar el binario
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /root/

# Copiar el binario desde el builder
COPY --from=builder /worker .

# Exponer puerto (aunque el worker no usa HTTP, lo dejamos por consistencia)
EXPOSE 8081

# Ejecutar el worker
CMD ["./worker"]
