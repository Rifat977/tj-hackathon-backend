# Multi-stage build for Go Fiber application
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates, tzdata, and tini for proper signal handling
RUN apk --no-cache add ca-certificates tzdata tini wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy assets directory if it exists
COPY --from=builder /app/assets ./assets
# Copy views directory if it exists
COPY --from=builder /app/views ./views

# Create necessary directories and set permissions
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Enhanced health check with better timeout and interval
HEALTHCHECK --interval=20s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --timeout=3 http://localhost:3000/api/health || exit 1

# Use tini as init system for proper signal handling and zombie reaping
ENTRYPOINT ["/sbin/tini", "--"]

# Run the application
CMD ["./main"] 