# Multi-stage build for better caching and smaller images

# Stage 1: Dependencies layer
FROM golang:1.24-alpine AS deps

# Install system dependencies and create user in one layer
RUN apk add --no-cache git ca-certificates tzdata curl su-exec && \
    adduser -D -s /bin/sh -u 1000 devuser

WORKDIR /app
USER devuser

# Copy go mod files first for better layer caching
COPY --chown=devuser:devuser go.mod go.sum ./
RUN go mod download && go mod verify

# Stage 2: Development image with Air
FROM deps AS development

# Install Air for live reload
RUN go install github.com/air-verse/air@latest

# Copy source code
COPY --chown=devuser:devuser . .

# Create logs directory and set permissions
RUN mkdir -p /app/logs && \
    chown -R devuser:devuser /app/logs && \
    chmod -R 755 /app/logs

# Switch to devuser
USER devuser

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["air", "-c", ".air.toml"]

# Stage 3: Production build
FROM deps AS builder

COPY --chown=devuser:devuser . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -buildvcs=false \
    -ldflags='-w -s -extldflags "-static"' \
    -o go-clean-template \
    ./cmd/api

# Stage 4: Production image
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh -u 1000 appuser

WORKDIR /app
USER appuser

COPY --from=builder --chown=appuser:appuser /app/go-clean-template .
COPY --from=builder --chown=appuser:appuser /app/config ./config

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./go-clean-template"]