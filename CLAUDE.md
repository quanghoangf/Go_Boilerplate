# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Core Commands
- `make init` - Install required tools (wire, mockgen, swag)
- `make bootstrap` - Start development environment (docker-compose up, run migrations, start server)
- `make build` - Build the server binary to `./bin/server`
- `make test` - Run tests with coverage report (generates `coverage.html`)
- `make mock` - Generate mocks for services and repositories
- `make swag` - Generate Swagger documentation

### Database Commands
- `go run ./cmd/migration` - Run database migrations
- `sqlc generate` - Generate Go code from SQL queries (configured in `sqlc.yaml`)

### Running the Application
- `go run ./cmd/server` - Start the development server
- Server runs on `localhost:8000` by default
- Swagger docs available at `http://localhost:8000/swagger/index.html`

## Architecture Overview

This is a Go web application boilerplate following **Clean Architecture** principles with dependency injection via Google Wire.

### Project Structure
```
cmd/           - Application entry points (server, migration, task)
internal/      - Private application code
  ├── handler/ - HTTP handlers (Gin controllers)
  ├── service/ - Business logic layer
  ├── repository/ - Data access layer with GORM
  ├── model/   - Domain models
  ├── middleware/ - HTTP middleware (CORS, JWT, logging)
  ├── server/  - Server setup (HTTP, jobs, migrations)
  └── db/      - Database schemas and SQLC generated code
pkg/           - Shared/public packages
api/v1/        - API definitions and error handling
test/          - Test files with mocks
```

### Key Dependencies and Patterns
- **Web Framework**: Gin HTTP framework
- **ORM**: GORM with support for PostgreSQL, MySQL, SQLite
- **Database Query Builder**: SQLC for type-safe SQL queries
- **Dependency Injection**: Google Wire for compile-time DI
- **Authentication**: JWT tokens with custom middleware
- **Logging**: Zap structured logging
- **Configuration**: Viper (YAML configs in `config/`)
- **Documentation**: Swagger/OpenAPI via swaggo
- **Testing**: Testify with gomock for mocking

### Database Layer
- Uses SQLC for type-safe SQL queries (see `internal/db/`)
- GORM for ORM operations and migrations
- Transaction support via repository pattern
- Database configuration supports multiple drivers (postgres/mysql/sqlite)

### Configuration
- Environment-specific configs in `config/` directory
- Use `config/local.yml` for development
- Use `config/prod.yml` for production
- Pass config file with `-conf` flag: `go run ./cmd/server -conf config/local.yml`

### Wire Dependency Injection
- Wire configuration files in `cmd/*/wire/` directories
- Run `wire` command in these directories to regenerate `wire_gen.go`
- Each command (server, migration, task) has its own wire setup

### Testing Strategy
- Unit tests in `test/server/` directory
- Mocks generated with mockgen in `test/mocks/`
- Tests cover handler, service, and repository layers
- Run tests with coverage: `make test`