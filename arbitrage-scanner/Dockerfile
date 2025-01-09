# Use the official Golang image as a base
FROM golang:1.20 AS build

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to the working directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
# Ensure this copies all source files in the arbitrage-scanner directory
COPY . .  

# Build the Go application
RUN go build -o arbitrage-scanner .

# Use a minimal image for the final build
FROM alpine:3.17
WORKDIR /app
COPY --from=build /app/arbitrage-scanner .
CMD ["./arbitrage-scanner"]
