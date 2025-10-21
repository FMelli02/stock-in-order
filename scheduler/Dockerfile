# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar dependencias de compilación
RUN apk add --no-cache git

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar el binario
RUN CGO_ENABLED=0 GOOS=linux go build -o /scheduler ./cmd/api

# Stage 2: Run
FROM alpine:latest

WORKDIR /root/

# Copiar el binario compilado
COPY --from=builder /scheduler .

# Ejecutar el scheduler
CMD ["./scheduler"]
