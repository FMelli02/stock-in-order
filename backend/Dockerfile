# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install build tools if needed (git for go modules)
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/api

# ---- Final stage ----
FROM alpine:latest
WORKDIR /app

# Copy only the compiled binary
COPY --from=builder /main ./main

# Expose the port used by the app
EXPOSE 8080

# Run the binary
CMD ["./main"]
