# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Core Commands
- `make init` - Install required tools (wire, mockgen, swag, air, buf, protoc-gen-*)
- `make bootstrap` - Start development environment (docker-compose up, run migrations, start server)
- `make dev` - Start development server with live reload using Air
- `make build` - Build the server binary to `./bin/server`
- `make test` - Run tests with coverage report (generates `coverage.html`)
- `make mock` - Generate mocks for services and repositories
- `make swag` - Generate Swagger documentation
- `make air-init` - Initialize Air configuration (if needed)

### Protocol Buffers Commands
- `make proto-gen` - Generate Go code from Protocol Buffers using Buf
- `make proto-lint` - Lint Protocol Buffer files
- `make proto-breaking` - Check for breaking changes in proto files

### Database Commands
- `make migrate-up` - Apply all pending database migrations
- `make migrate-down` - Rollback the last migration
- `make migrate-version` - Check current migration version
- `make migrate-create NAME=<name>` - Create new migration files
- `make migrate-force VERSION=<version>` - Force migration to specific version (use with caution)
- `sqlc generate` - Generate Go code from SQL queries (configured in `sqlc.yaml`)

### Running the Application
- `make dev` - Start development server with live reload (recommended for development)
- `go run ./cmd/server` - Start the development server manually
- **HTTP server** runs on `localhost:8000` by default
- **gRPC server** runs on `localhost:8080` by default
- Swagger docs available at `http://localhost:8000/swagger/index.html`

### Development Workflow
- Use `make dev` for development - Air will automatically restart the server when you save changes
- Air watches Go files in `cmd/`, `internal/`, `pkg/`, and `api/` directories
- Configuration files (`.yml`, `.yaml`) are also monitored for changes
- Builds are logged to `build-errors.log` if compilation fails

## Architecture Overview

This is a Go web application boilerplate following **Clean Architecture** principles with dependency injection via Google Wire. It supports both HTTP REST and gRPC APIs running simultaneously.

### Project Structure
```
cmd/           - Application entry points (server, migration, task)
internal/      - Private application code
  ├── handler/ - HTTP handlers (Gin controllers)
  ├── grpc/    - gRPC service implementations and interceptors
  ├── service/ - Business logic layer
  ├── repository/ - Data access layer with GORM
  ├── model/   - Domain models
  ├── middleware/ - HTTP middleware (CORS, JWT, logging)
  ├── server/  - Server setup (HTTP, gRPC, jobs, migrations)
  └── db/      - Database schemas and SQLC generated code
pkg/           - Shared/public packages
api/           - API definitions
  ├── v1/      - REST API definitions and error handling
  └── grpc/    - Generated gRPC code from Protocol Buffers
proto/         - Protocol Buffer definitions
  └── user/v1/ - User service proto files
test/          - Test files with mocks
```

### Key Dependencies and Patterns
- **Web Framework**: Gin HTTP framework
- **RPC Framework**: gRPC with Protocol Buffers
- **ORM**: GORM with support for PostgreSQL, MySQL, SQLite
- **Database Query Builder**: SQLC for type-safe SQL queries
- **Dependency Injection**: Google Wire for compile-time DI
- **Authentication**: JWT tokens with custom middleware and gRPC interceptors
- **Logging**: Zap structured logging
- **Configuration**: Viper (YAML configs in `config/`)
- **Documentation**: Swagger/OpenAPI via swaggo
- **Protocol Buffers**: Buf for linting, breaking change detection, and code generation
- **Testing**: Testify with gomock for mocking

### Database Layer
- **Migrations**: Uses golang-migrate for database schema management (see `migrations/` directory)
- **Query Builder**: SQLC for type-safe SQL queries (see `internal/db/`)
- **ORM**: GORM for ORM operations
- **Transaction Support**: Repository pattern with transaction support
- **Database Support**: PostgreSQL (primary), MySQL, SQLite

### Configuration
- Environment-specific configs in `config/` directory
- Use `config/local.yml` for development
- Use `config/prod.yml` for production
- Pass config file with `-conf` flag: `go run ./cmd/server -conf config/local.yml`

### Wire Dependency Injection
- Wire configuration files in `cmd/*/wire/` directories
- Run `wire` command in these directories to regenerate `wire_gen.go`
- Each command (server, migration, task) has its own wire setup

### gRPC Implementation
- **Protocol Buffers**: Defined in `proto/` directory with comprehensive documentation
- **Code Generation**: Uses Buf for consistent and reliable code generation
- **Interceptors**: Authentication, logging, and recovery interceptors for robust request handling
- **Error Mapping**: Proper conversion from domain errors to gRPC status codes
- **Dual Protocol**: Both REST and gRPC share the same business logic layer

### Testing Strategy
- Unit tests in `test/server/` directory
- Mocks generated with mockgen in `test/mocks/`
- Tests cover handler, service, and repository layers
- Support for both HTTP and gRPC testing
- Run tests with coverage: `make test`