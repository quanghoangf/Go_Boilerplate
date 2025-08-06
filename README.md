# Go Web API Boilerplate

A production-ready Go web API boilerplate built with clean architecture principles. This boilerplate provides a solid foundation for building scalable web APIs with JWT authentication, database integration, comprehensive testing, and excellent developer experience.

- [x] add `go install github.com/air-verse/air@latest`
- [x] add slqc
- [x] add grpc
- [ ] add socical login
- [ ] add queue, message queues (Kafka)
- [ ] add sent mail
- [ ] add file uploads
- [ ] add websocket
- [ ] add notifications systems
- [ ] add pagination, filter
- [ ] add graphql
- [ ] add hyerachy
- [ ] add global exeptions
- [ ] add validate decorators
- [ ] add workers
- [ ] add emit, listen events
- [ ] add infracstructors (...)

## Features

- **Clean Architecture**: Well-structured codebase following clean architecture principles
- **JWT Authentication**: Secure authentication with JWT tokens
- **Database Integration**: Support for PostgreSQL, MySQL, and SQLite with GORM
- **Type-safe Queries**: SQLC for generating type-safe Go code from SQL
- **Dependency Injection**: Google Wire for compile-time dependency injection
- **API Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Comprehensive Testing**: Unit tests with mocks and coverage reports
- **Logging**: Structured logging with Zap
- **Configuration**: Environment-based configuration with Viper
- **Middleware**: CORS, logging, and authentication middleware
- **Docker Support**: Containerization and docker-compose setup
- **Live Reload Development**: Air integration for automatic server restart during development
- **gRPC Support**: High-performance gRPC server with Protocol Buffers, interceptors, and comprehensive error handling
- **Dual Protocol**: Run both HTTP REST and gRPC services simultaneously

## Getting Started

1. **Install dependencies**:
   ```bash
   make init
   ```

2. **Start development environment**:
   ```bash
   make bootstrap
   ```

3. **Start development server with live reload**:
   ```bash
   make dev
   ```

4. **Run tests**:
   ```bash
   make test
   ```

5. **Generate Protocol Buffers**:
   ```bash
   make proto-gen
   ```

6. **Generate API documentation**:
   ```bash
   make swag
   ```

## API Endpoints

The application provides both REST and gRPC endpoints:

### REST API (Port 8000)
- `POST /v1/register` - User registration
- `POST /v1/login` - User login
- `GET /v1/user` - Get user profile (requires auth)
- `PUT /v1/user` - Update user profile (requires auth)
- `GET /swagger/index.html` - Swagger documentation

### gRPC API (Port 8080)
- `UserService.Register` - User registration
- `UserService.Login` - User login  
- `UserService.GetProfile` - Get user profile (requires auth)
- `UserService.UpdateProfile` - Update user profile (requires auth)

### Authentication
- REST: Use `Authorization: Bearer <token>` header
- gRPC: Use `authorization: Bearer <token>` metadata

## License

This project is released under the MIT License. For more information, see the [LICENSE](LICENSE) file.
