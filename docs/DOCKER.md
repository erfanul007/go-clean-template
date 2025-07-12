# Docker Operations Guide

Advanced Docker operations, troubleshooting, and deployment guide for the Go Clean Architecture Template. For basic setup, see [README.md](../README.md).

## 🐳 Docker-First Benefits

- **No Local Dependencies:** Only Docker and Task required
- **Consistent Environment:** Same Go version (1.24) across all developers and CI/CD
- **Isolated Operations:** Clean container environment for each command
- **Cross-Platform:** Works identically on Windows, macOS, and Linux
- **Multi-Stage Builds:** Optimized for both development and production

## 🔧 Configuration

### Config-First Approach

This project uses a **config-first approach** where `docker-compose.yml` contains minimal environment variables, with most configuration in `config.yaml` or `.env` files.

**Docker Compose Variables (Networking Only):**
```bash
DB_HOST=postgres
REDIS_HOST=redis
```

**Configuration Hierarchy:**
1. **Environment Variables** → Runtime overrides
2. **config.yaml** → Application defaults
3. **Code Defaults** → Fallback values

### Container Names
- **API:** `go-clean-template-api`
- **Database:** `go-clean-template-db`
- **Redis:** `go-clean-template-redis`

## 🔍 Debugging

### Container Access

```bash
# Enter containers
docker exec -it go-clean-template-api sh
docker exec -it go-clean-template-db psql -U postgres -d app_db
docker exec -it go-clean-template-redis redis-cli

# Monitor resources
docker stats
docker stats --no-stream

# View logs
task logs                    # All services
task compose-logs           # Alternative
docker compose -f deployments/docker-compose.yml logs api
docker compose -f deployments/docker-compose.yml logs postgres
```

### Configuration Debugging

```bash
# Validate docker-compose configuration
docker compose -f deployments/docker-compose.yml config

# Check environment setup
task setup

# Verify service dependencies
task health
```

## 🐛 Troubleshooting

### Common Issues

```bash
# Service status
task health
docker compose -f deployments/docker-compose.yml ps

# Database connectivity
docker exec go-clean-template-api nc -zv postgres 5432

# Clean rebuild
task compose-clean          # Remove volumes and containers
task compose-rebuild        # Rebuild and restart

# Force rebuild without cache
docker compose -f deployments/docker-compose.yml build --no-cache
```

### Performance

```bash
# Enable BuildKit (enabled by default in newer Docker versions)
set DOCKER_BUILDKIT=1     # Windows
export DOCKER_BUILDKIT=1  # Linux/Mac

# Check optimized .dockerignore
type .dockerignore   # Windows
cat .dockerignore    # Linux/Mac
```

**Build Context Optimization**: The `.dockerignore` file excludes unnecessary files for faster builds and smaller context.

## 🏗️ Multi-Stage Docker Architecture

The `build/Dockerfile` uses a **4-stage build process**:

**Stage 1: Dependencies** (`deps`):
- Base Go 1.24 Alpine image with system dependencies
- Non-root user setup for security
- Go module download and verification

**Stage 2: Development** (`development`):
- Includes Air for live reload
- Full source code and development tools
- Optimized for fast iteration

**Stage 3: Builder** (`builder`):
- Compiles optimized production binary
- Static linking with security flags
- Minimal attack surface

**Stage 4: Production** (`production`):
- Minimal Alpine base
- Only binary and config files
- Non-root execution
- Built-in health checks

**Benefits:**
- **Development**: Fast rebuilds with layer caching
- **Production**: Minimal image size and attack surface
- **Security**: Non-root execution in both stages
- **Performance**: Optimized binaries with static linking

## 🚀 Production Deployment

### Image Building
```bash
# Build production image (targets 'production' stage)
task docker-build

# Build specific stage
docker build -f build/Dockerfile --target production -t go-clean-template:prod .
docker build -f build/Dockerfile --target development -t go-clean-template:dev .

# Tag for registry
docker tag go-clean-template:latest your-registry.com/go-clean-template:v1.0.0

# Push to registry
docker push your-registry.com/go-clean-template:v1.0.0
```

### Health Monitoring
- **Built-in health checks**: Both development and production images include health checks
- **Endpoints**: `/health`, `/live`, `/ready` for comprehensive monitoring
- **Verification**: Use `task health` for status verification

## 📦 Data Management

```bash
# Backup database volume
docker run --rm -v go-clean-template_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz -C /data .

# Restore database volume
docker run --rm -v go-clean-template_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_backup.tar.gz -C /data

# Inspect volumes
docker run --rm -v go-clean-template_postgres_data:/data alpine ls -la /data
```

## 📚 Docker Operations

### Task Commands
```bash
# Development
task start              # Setup and start development
task dev                # Start with live reload
task restart            # Restart development environment

# Docker Compose
task compose-up         # Start all services
task compose-down       # Stop all services
task compose-rebuild    # Rebuild and restart
task compose-clean      # Remove volumes and containers
task compose-logs       # View all service logs

# Code Quality
task check              # Run all quality checks
task fmt                # Format code
task lint               # Lint code
task test               # Run tests

# Build & Dependencies
task docker-build       # Build production image
task deps               # Download dependencies
task deps-update        # Update dependencies

# Project Management
task setup              # One-time project setup
task clean              # Clean everything
task health             # Check service health
```

### Direct Docker Compose
```bash
# Basic operations
docker compose -f deployments/docker-compose.yml up -d
docker compose -f deployments/docker-compose.yml down
docker compose -f deployments/docker-compose.yml logs -f
```