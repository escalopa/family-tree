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

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Copy config file
COPY --from=builder /app/config.yaml ./config.yaml

CMD ["./main"]
