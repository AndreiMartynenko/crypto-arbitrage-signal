# Use an official Go image
FROM golang:1.20-alpine AS build

# Create an app directory
WORKDIR /app

# Copy module files and download dependencies first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o binance-connector .

# Final stage: minimal runtime image
FROM alpine:3.17
WORKDIR /app

# Copy compiled binary from the build stage
COPY --from=build /app/binance-connector /app/binance-connector

# Set environment variables (example)
ENV EXCHANGE_API_KEY=your_binance_key
ENV EXCHANGE_API_SECRET=your_binance_secret
ENV PAIRS=BTC/USDT,ETH/USDT
ENV POLL_INTERVAL=5s

# Expose a port if the connector serves an HTTP endpoint
EXPOSE 8001

# Default command
CMD ["./binance-connector"]
