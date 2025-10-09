# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o jot ./cmd/jot

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 jot && \
    adduser -D -u 1000 -G jot jot

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/jot /usr/local/bin/jot

# Create directories for documentation
RUN mkdir -p /app/docs /app/dist && \
    chown -R jot:jot /app

# Switch to non-root user
USER jot

# Default configuration
COPY --chown=jot:jot jot.yaml.example /app/jot.yaml

# Expose port for serve command
EXPOSE 8080

# Volume for documentation source
VOLUME ["/app/docs"]

# Volume for output
VOLUME ["/app/dist"]

# Default command
ENTRYPOINT ["jot"]
CMD ["build"]