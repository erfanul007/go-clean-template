# Docker Operations Guide

Advanced Docker operations, troubleshooting, and deployment guide for the Go Clean Architecture Template. For basic setup, see [README.md](../README.md).

## 🐳 Docker-First Benefits

- **No Local Dependencies:** Only Docker and Task required
- **Consistent Environment:** Same Go version across all developers and CI/CD
- **Isolated Operations:** Clean container environment for each command
- **Cross-Platform:** Works identically on Windows, macOS, and Linux

## 🔧 Advanced Configuration

### Production Environment Variables

Critical variables to update for production:

```bash
# Security (REQUIRED)
JWT_SECRET=your-secure-random-secret-key-change-this-in-production
DB_PASSWORD=secure-database-password
REDIS_PASSWORD=secure-redis-password

# Environment Configuration
ENVIRONMENT=production
LOG_LEVEL=info
LOG_FORMAT=json

# Server Configuration
PORT=8080
HOST=0.0.0.0

# Database Configuration (if external)
DB_HOST=your-postgres-host
DB_PORT=5432
DB_USER=your-db-user
DB_NAME=your-db-name
DB_SSLMODE=require

# Redis Configuration (if external)
REDIS_HOST=your-redis-host
REDIS_PORT=6379

# Authentication
JWT_EXPIRATION=3600
```

**Note**: Static configurations (CORS, rate limiting, Swagger, metrics) are now in `config/config.yaml` and don't need environment variables.

### Data Management

```bash
# Backup database
docker run --rm -v go-clean-template_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz -C /data .

# Restore database
docker run --rm -v go-clean-template_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_backup.tar.gz -C /data

# Inspect volumes
docker run --rm -v go-clean-template_postgres_data:/data alpine ls -la /data
```

## 🔍 Debugging

### Container Access

```bash
# Enter containers
docker exec -it go-clean-template-api sh
docker exec -it go-clean-template-db psql -U postgres -d app_db
docker exec -it go-clean-template-redis redis-cli

# Monitor resources
docker stats go-clean-template-api
docker stats --no-stream

# Inspect containers
docker inspect go-clean-template-api

# View logs
docker logs -t go-clean-template-api
docker logs --since="1h" go-clean-template-db
task compose-logs  # All services
```

## 🐛 Troubleshooting

### Common Issues

```bash
# Port conflicts
netstat -an | findstr :8080  # Windows
lsof -i :8080               # macOS/Linux

# Database connectivity
docker exec go-clean-template-api nc -zv postgres 5432

# Clean rebuild
task compose-clean
task compose-rebuild

# Force rebuild without cache
docker compose -f deployments/docker-compose.yml build --no-cache

# Service status
task health
docker compose -f deployments/docker-compose.yml ps
```

### Performance

```bash
# Enable BuildKit
set DOCKER_BUILDKIT=1     # Windows
export DOCKER_BUILDKIT=1  # Linux/Mac

# Check optimized .dockerignore (54 lines, 41% smaller)
type .dockerignore   # Windows
cat .dockerignore    # Linux/Mac
```

**Build Context Optimization**: The `.dockerignore` file has been optimized to exclude unnecessary files while maintaining functionality:
- **Reduced size**: From 92 to 54 lines (41% reduction)
- **Better organization**: Grouped by category (Git, docs, IDE, OS, build artifacts, etc.)
- **Improved performance**: Faster Docker builds with smaller context

### Configuration System

The application uses a **hybrid configuration system** optimized for Docker deployments:

**Environment Variables** (`.env` → Docker environment):
- Server settings, database credentials, Redis connection
- Authentication secrets, logging preferences
- Environment-specific and sensitive configurations

**Static Configuration** (`config/config.yaml`):
- CORS policies, rate limiting, Swagger documentation
- Metrics configuration, server timeouts
- Application behavior that rarely changes

**Benefits for Docker**:
- **Security**: Secrets only in environment variables
- **Flexibility**: Easy environment overrides via Docker Compose
- **Maintainability**: Static configs in version control
- **Deployment**: Simple `.env` changes for different environments

## 🚀 Production Deployment

### Image Building
```bash
# Build production image
task docker-build

# Tag for registry
docker tag go-clean-template:latest your-registry.com/go-clean-template:v1.0.0

# Push to registry
docker push your-registry.com/go-clean-template:v1.0.0
```

### Health Monitoring
- Built-in health checks via `/health` endpoint
- Service dependency health conditions
- Use `task health` for status verification

## 🔄 CI/CD Integration

### GitHub Actions Example
```yaml
name: Docker Build and Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install Task
        uses: arduino/setup-task@v1
      - name: Test
        run: |
          task setup
          task dev &
          sleep 30
          task health
          task test-docker
          task compose-down
      - name: Build
        run: task docker-build
```

## 📚 Advanced Task Commands

### Docker Operations
```bash
task compose-up         # Start all services
task compose-down       # Stop all services
task compose-rebuild    # Rebuild and restart
task compose-clean      # Remove volumes and containers
task compose-logs       # View all service logs
task docker-build       # Build application image
task health             # Check health of all services
```

### Development Tools
```bash
task fmt                # Format code
task lint               # Lint code
task generate           # Run go generate
task build              # Build binary
task test               # Run unit tests
task test-docker        # Test in running environment
task deps-vendor        # Create vendor directory
```

## 📚 Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)
- [Task Documentation](https://taskfile.dev/)
- [Docker Security](https://docs.docker.com/engine/security/)