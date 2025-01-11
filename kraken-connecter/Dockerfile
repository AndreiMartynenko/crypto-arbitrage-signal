# Use the official Golang image as a base
FROM golang:1.20-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files from the project root
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code from the binance-connector folder
COPY kraken-connector/ .

# Build the application
RUN go build -o kraken-connector .

# Use a minimal image for the final build
FROM alpine:3.17
WORKDIR /app
COPY --from=build /app/kraken-connector .
EXPOSE 8001
CMD ["./kraken-connector"]
