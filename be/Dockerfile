# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install swag for generating Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/main.go -o docs

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Development stage - for hot reload with air
FROM golang:1.25-alpine AS development

WORKDIR /app

# Install dependencies needed for development and healthcheck
RUN apk add --no-cache curl

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

# Install swag for generating Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Source code will be mounted as volume
# Generate swagger docs on startup and run air
CMD ["sh", "-c", "swag init -g cmd/main.go -o docs 2>/dev/null || true && air -c .air.toml"]

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Config will be mounted as volume
CMD ["./main"]
