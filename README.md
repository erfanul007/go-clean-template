# Go Clean Architecture Template

A robust Go API template built following Clean Architecture principles and Domain-Driven Design (DDD) patterns. This project uses **Docker-first development** for consistency across all environments.

## 🚀 Quick Start

### Prerequisites
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Task (Taskfile)** - [Installation guide](https://taskfile.dev/installation/)

**Note:** No local Go installation required!

### Setup & Run
```bash
git clone <repository-url>
cd go-clean-template
task start  # One-command setup and start
```

**Access Points:**
- API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health

## 🐳 Services

| Service | Port | Description |
|---------|------|-----------|
| **API** | 8080 | Go application with live reload |
| **PostgreSQL** | 5432 | Primary database |
| **Redis** | 6380 (host) | Caching layer |

## 📋 Essential Commands

```bash
# Development
task start              # Setup and start development
task dev                # Start with live reload
task check              # Run quality checks (format, lint, test)
task health             # Check service health
task clean              # Clean everything

# Testing
task test               # Run tests
task test-coverage      # Run with coverage

# Dependencies
task deps               # Download and tidy dependencies
task deps-update        # Update dependencies
```

For advanced Docker operations, see [DOCKER.md](docs/DOCKER.md).

## 🏗️ Tech Stack

**Current:**
- **Go 1.24** with Chi Router
- **PostgreSQL 15** + **Redis 7** (configured)
- **Viper** (config) + **Zap** (logging)
- **Swagger/OpenAPI** documentation
- **Docker & Docker Compose**
- Health monitoring endpoints

**Planned:**
- Ent ORM, JWT Authentication, Business Logic Implementation

## 📁 Project Structure

```
go-clean-template/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── application/            # Use cases and business logic
│   ├── domain/                 # Domain entities and business rules
│   ├── infrastructure/         # External concerns (database, config, logging)
│   ├── presentation/           # HTTP handlers, routes, API documentation
│   └── shared/                 # Shared utilities (errors, response)
├── config/                     # Configuration files
├── docs/                       # API documentation
├── build/                      # Dockerfiles
├── deployments/               # Docker Compose & scripts
└── Taskfile.yml               # Task automation
```

## 🏛️ Clean Architecture Layers

This project implements Clean Architecture with clear separation of concerns:

### 🎯 Domain (`internal/domain/`)
**Core business logic** - Entities, value objects, business rules. No external dependencies.
```
domain/
├── example/         # Domain-specific entities
├── shared/          # Common domain concepts
│   ├── events/      # Domain events
│   ├── values/      # Shared value objects
│   └── interfaces/  # Domain contracts
└── services/        # Domain services
```

### 🔄 Application (`internal/application/`)
**Use cases** - Orchestrates domain objects, depends only on domain layer.
```
application/
├── example/         # Application services
│   ├── commands/    # Command handlers
│   ├── queries/     # Query handlers
│   ├── dto/         # Application DTOs
│   └── interfaces/  # Repository contracts
├── common/          # Shared application logic
└── services/        # Application services
```

### 🌐 Presentation (`internal/presentation/`)
**HTTP interface** - Handlers, routes, middleware, DTOs, Swagger documentation.
```
presentation/http/
├── handlers/    # HTTP request handlers
├── middleware/  # CORS, auth, logging
├── dto/         # Request/response structures
├── routes.go    # API endpoints
└── server.go    # HTTP server setup
```

### 🔧 Infrastructure (`internal/infrastructure/`)
**External concerns** - Database, config, logging, authentication implementations.
```
infrastructure/
├── auth/        # JWT, password hashing
├── config/      # Environment, YAML config
├── logger/      # Structured logging
└── persistence/ # Database, repositories
```

### 🤝 Shared (`internal/shared/`)
**Common utilities** - Error handling, response formatting, validation.

### 🎯 Key Principles
- **Dependency Rule**: Outer layers depend on inner layers
- **Framework Independence**: Business logic isolated from frameworks
- **Testability**: Each layer independently testable
- **Single Responsibility**: Each layer has one clear purpose

### 🔄 Request Flow
```
HTTP → Presentation → Application → Domain
  ↑         ↓            ↓         ↓
Response ← Infrastructure ← Infrastructure
```

## 🔍 Health Endpoints

| Endpoint | Purpose |
|----------|----------|
| `/health` | Detailed health with dependencies |
| `/heartbeat` | Simple heartbeat |
| `/live` | Container liveness probe |
| `/ready` | Container readiness probe |
| `/system` | System information |

## 📚 API Documentation

**Swagger UI:** http://localhost:8080/swagger/index.html

**Currently Available Endpoints:**
- Health monitoring (`/health`, `/live`, `/ready`)
- Heartbeat (`/heartbeat`)

**Planned:** Full RESTful API implementation following clean architecture patterns.

## 🔧 Configuration

Uses hybrid configuration:
- Environment variables (`.env` file)
- YAML configuration (`config/config.yaml`)
- Docker environment (docker-compose.yml)

`task setup` automatically copies `.env.example` to `.env` and downloads dependencies.

## 🧪 Testing

```bash
task test               # Run tests
task test-coverage      # Run with coverage
```

All tests run in Docker for consistency across environments.

## 🚀 Deployment

```bash
task setup              # Copy .env.example to .env
# Edit .env with production values
task start              # Start all services
task health             # Verify services
```

For production, update `.env` with secure values (JWT_SECRET, passwords, etc.).

## 🤝 Contributing

1. Install Docker and Task
2. Run `task setup` and `task dev`
3. Make changes (live reload enabled)
4. Run `task check` before committing
5. Submit pull request

## 📚 Additional Documentation

- **[DOCKER.md](docs/DOCKER.md)** - Advanced Docker operations, troubleshooting, and deployment

## 👨‍💻 Author

**Md. Erfanul Islam Bhuiyan** - Software Engineer

[![GitHub](https://img.shields.io/badge/GitHub-erfanul007-181717?style=flat&logo=github)](https://github.com/erfanul007) [![LinkedIn](https://img.shields.io/badge/LinkedIn-erfanul007-0077B5?style=flat&logo=linkedin)](https://www.linkedin.com/in/erfanul007/)
