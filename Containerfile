FROM registry.access.redhat.com/hi/go:latest-builder AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 go build -o openshell-proxy ./cmd/openshell-proxy

# Use a minimal base image to reduce the size of the final image
FROM registry.access.redhat.com/hi/static:latest

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/openshell-proxy .

# Expose application port
EXPOSE 8080

# Run the application
CMD ["./openshell-proxy"]