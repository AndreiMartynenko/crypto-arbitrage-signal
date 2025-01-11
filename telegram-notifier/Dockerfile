# Build stage
FROM golang:1.20 AS builder

WORKDIR /app

# Copy Go modules from the root directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code from the `telegram-notifier` directory
COPY telegram-notifier/ .

# Build the application for Linux
RUN GOOS=linux GOARCH=amd64 go build -o telegram-notifier

# Runtime stage
FROM alpine:3.17

WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/telegram-notifier .

# Make the binary executable
RUN chmod +x telegram-notifier

# Expose the service on port 8004
EXPOSE 8004

# Healthcheck endpoint
HEALTHCHECK --interval=10s --timeout=5s --retries=3 CMD curl -f http://localhost:8004/health || exit 1

# Run the application
CMD ["./telegram-notifier"]
