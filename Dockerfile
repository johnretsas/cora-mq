# Start from the official Go base image
FROM golang:1.22-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first for dependency resolution
COPY go.mod ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application code to the container
COPY . .

# Build the Go application as a statically linked binary
RUN go build -o /queue-service .

# Final image
FROM alpine:latest

# Create a non-root user for running the application
RUN addgroup -S queue && adduser -S queue -G queue

# Install curl
RUN apk add --no-cache curl

# Copy the statically compiled Go binary from the builder image
COPY --from=builder /queue-service /queue-service

# Set the binary as the entry point of the container
ENTRYPOINT ["/queue-service"]

# Expose the port on which the service will run
EXPOSE 8080

# Run as non-root user
USER queue
