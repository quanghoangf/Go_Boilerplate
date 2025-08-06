# Go Web API Boilerplate

A production-ready Go web API boilerplate built with clean architecture principles. This boilerplate provides a solid foundation for building scalable web APIs with JWT authentication, database integration, comprehensive testing, and excellent developer experience.

- [ ] add `go install github.com/air-verse/air@latest`
- [x] add slqc
- [ ] add socical login
- [ ] add queue, message queues (Kafka)
- [ ] add sent mail
- [ ] add file uploads
- [ ] add websocket
- [ ] add notifications systems
- [ ] add pagination, filter
- [ ] add graphql
- [ ] add grpc
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

## Getting Started

1. **Install dependencies**:
   ```bash
   make init
   ```

2. **Start development environment**:
   ```bash
   make bootstrap
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Generate API documentation**:
   ```bash
   make swag
   ```

## License

This project is released under the MIT License. For more information, see the [LICENSE](LICENSE) file.
