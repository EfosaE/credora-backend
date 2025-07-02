FROM golang:1.24-bullseye AS builder

# Set working directory inside the container
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached if go.mod/go.sum don't change)
RUN go mod download

# Copy all source code
COPY . .

# Build the Go application from cmd folder
# Note: Building from ./cmd and outputting as 'server' (not 'main')
RUN go build -o server ./cmd

# Use a minimal base image for the final build
FROM debian:bullseye-slim


# Set working directory in the final container
WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /app/server .

# Expose the port the server listens on (change if needed)
EXPOSE 8080

# Run the server binary
CMD ["./server"]