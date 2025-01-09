# Use the official Golang image as a base
FROM golang:1.20-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files from the root of your project
COPY ../../go.mod ../../go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the `binance-connector` source code
COPY . .

# Build the Go application
RUN go build -o binance-connector .

# Use a minimal base image to run the app
FROM alpine:3.17

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled binary from the build stage
COPY --from=build /app/binance-connector .

# Expose the HTTP port
EXPOSE 8001

# Run the application
CMD ["./binance-connector"]
