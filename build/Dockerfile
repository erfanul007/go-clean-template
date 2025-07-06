# Multi-stage build for better caching and smaller images

# Stage 1: Base image with dependencies
FROM golang:1.24-alpine AS base

# Install system dependencies
RUN apk add --no-cache git ca-certificates tzdata curl

# Create non-root user
RUN adduser -D -s /bin/sh -u 1000 devuser

# Set working directory
WORKDIR /app

# Change ownership
RUN chown devuser:devuser /app

# Stage 2: Dependencies layer (cached separately)
FROM base AS deps

# Switch to non-root user for dependency management
USER devuser

# Copy go mod files first for better layer caching
COPY --chown=devuser:devuser go.mod go.sum ./

# Download dependencies (this layer will be cached unless go.mod/go.sum changes)
RUN go mod download && go mod verify

# Stage 3: Development image with Air
FROM deps AS development

# Install Air for live reload (as devuser to avoid permission issues)
RUN go install github.com/air-verse/air@latest

# Copy source code (this layer changes frequently)
COPY --chown=devuser:devuser . .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Default command for development with live reload
CMD ["air", "-c", ".air.toml"]

# Stage 4: Production build
FROM deps AS builder

# Copy source code
COPY --chown=devuser:devuser . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -buildvcs=false \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o go-clean-template \
    ./cmd/api

# Stage 5: Production image
FROM alpine:latest AS production

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh -u 1000 appuser

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder --chown=appuser:appuser /app/go-clean-template .

# Copy config files if needed
COPY --from=builder --chown=appuser:appuser /app/config ./config

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./go-clean-template"]