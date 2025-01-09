# Use an official Golang runtime as the base image
FROM golang:1.20

# Set the working directory
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o arbitrage-scanner .

# Expose the port (if you add an HTTP API in the future)
EXPOSE 8002

# Run the application
CMD ["./arbitrage-scanner"]
