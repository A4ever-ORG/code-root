# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o telegram-store-hub .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/telegram-store-hub .

# Copy configuration file
COPY --from=builder /app/.env.example .env.example

# Make binary executable
RUN chmod +x ./telegram-store-hub

# Expose port (if needed for health checks)
EXPOSE 8080

# Run the application
CMD ["./telegram-store-hub"]