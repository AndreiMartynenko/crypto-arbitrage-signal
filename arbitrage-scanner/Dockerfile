# Use the official Golang image as a base for building the binary
FROM golang:1.20 AS build

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum from the root build context
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code of arbitrage-scanner
COPY arbitrage-scanner/ .

# Debug: Check contents of /app
RUN ls -la /app

# Build the Go application with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o arbitrage-scanner .

# Use a minimal image (Alpine) for the runtime
FROM alpine:3.17
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/arbitrage-scanner .

# Ensure the binary is executable
RUN chmod +x arbitrage-scanner

# Debug: Check contents of /app in the runtime image
RUN ls -la /app

# Run the application
CMD ["./arbitrage-scanner"]
