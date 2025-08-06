# Go Web API Boilerplate Documentation

Welcome to the comprehensive documentation for the Go Web API Boilerplate. This documentation covers all aspects of the boilerplate, from basic usage to advanced development topics.

## 📚 Documentation Structure

### Main Documentation
- **[README.md](../README.md)** - Project overview, features, and quick start guide
- **[CLAUDE.md](../CLAUDE.md)** - Development commands and architecture overview for Claude Code

### gRPC Documentation
- **[gRPC Overview](grpc/README.md)** - Complete gRPC API documentation
- **[Interceptors Guide](grpc/interceptors.md)** - Authentication, logging, and recovery interceptors
- **[Client Examples](grpc/client-examples.md)** - Multi-language client implementations
- **[Development Guide](grpc/development-guide.md)** - Development workflows, testing, and best practices

## 🚀 Quick Start

1. **Setup Development Environment**:
   ```bash
   git clone <repository-url>
   cd Go_Boilerplate
   make init
   ```

2. **Start Development Servers**:
   ```bash
   make dev
   ```
   This starts both HTTP (port 8000) and gRPC (port 8080) servers with live reload.

3. **Generate API Code**:
   ```bash
   make proto-gen  # Generate Protocol Buffer code
   make swag       # Generate Swagger documentation
   ```

## 🏗️ Architecture Overview

The boilerplate follows **Clean Architecture** principles with:

- **HTTP REST API** (Gin framework) on port 8000
- **gRPC API** (Protocol Buffers) on port 8080  
- **Shared Business Logic** between both protocols
- **Database Layer** with GORM and SQLC
- **JWT Authentication** for both REST and gRPC
- **Dependency Injection** via Google Wire

## 📋 Available Services

### User Service (Both REST & gRPC)
- **Register** - Create new user account
- **Login** - Authenticate and get JWT token
- **GetProfile** - Retrieve user profile (requires auth)
- **UpdateProfile** - Update user information (requires auth)

## 🔧 Development Commands

### Core Commands
```bash
make init          # Install all required tools
make dev           # Start with live reload
make build         # Build production binary
make test          # Run tests with coverage
```

### Protocol Buffers
```bash
make proto-gen     # Generate Go code from .proto files
make proto-lint    # Lint Protocol Buffer files
make proto-breaking # Check for breaking changes
```

### Code Generation
```bash
make mock          # Generate mocks for testing
make swag          # Generate Swagger documentation
```

## 🧪 Testing

### API Testing with curl
```bash
# REST API
curl -X POST http://localhost:8000/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

curl -X POST http://localhost:8000/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### gRPC Testing with grpcurl
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Register user
grpcurl -plaintext \
  -d '{"email":"test@example.com","password":"password123"}' \
  localhost:8080 user.v1.UserService/Register

# Login
grpcurl -plaintext \
  -d '{"email":"test@example.com","password":"password123"}' \
  localhost:8080 user.v1.UserService/Login
```

### Unit Testing
```bash
make test          # Run all tests with coverage
go test ./...      # Run tests without coverage
go test -v ./test/server/handler  # Run specific test package
```

## 🛠️ Technology Stack

### Core Technologies
- **Go 1.19+** - Programming language
- **Gin** - HTTP web framework
- **gRPC** - High-performance RPC framework
- **Protocol Buffers** - Interface definition language

### Database & ORM
- **GORM** - ORM with PostgreSQL, MySQL, SQLite support
- **SQLC** - Generate type-safe Go from SQL

### Development Tools
- **Air** - Live reload for development
- **Buf** - Protocol Buffer toolchain
- **Wire** - Compile-time dependency injection
- **Swagger** - API documentation
- **Mockgen** - Mock generation for testing

### Authentication & Security
- **JWT** - JSON Web Tokens for authentication
- **Bcrypt** - Password hashing
- **CORS** - Cross-origin resource sharing

## 📁 Project Structure

```
├── cmd/                    # Application entry points
│   ├── server/            # Main HTTP+gRPC server
│   ├── migration/         # Database migrations
│   └── task/              # Background tasks
├── internal/              # Private application code
│   ├── handler/           # HTTP handlers (REST API)
│   ├── grpc/             # gRPC service implementations
│   │   └── interceptor/   # gRPC interceptors
│   ├── service/           # Business logic layer
│   ├── repository/        # Data access layer
│   ├── model/            # Domain models
│   ├── middleware/       # HTTP middleware
│   ├── server/           # Server setup
│   └── db/               # Database schemas & SQLC
├── pkg/                   # Shared/public packages
│   ├── jwt/              # JWT utilities
│   ├── log/              # Logging utilities
│   └── server/           # Server abstractions
├── api/                   # API definitions
│   ├── v1/               # REST API structs
│   └── grpc/             # Generated gRPC code
├── proto/                 # Protocol Buffer definitions
│   └── user/v1/          # User service proto files
├── config/               # Configuration files
│   ├── local.yml         # Development config
│   └── prod.yml          # Production config
├── test/                 # Test files
│   ├── mocks/            # Generated mocks
│   └── server/           # Integration tests
├── docs/                 # Documentation
│   └── grpc/             # gRPC-specific docs
└── deploy/               # Deployment configurations
```

## 🔄 Development Workflow

1. **Make Changes**: Edit source code, proto files, or configs
2. **Auto Reload**: Air automatically restarts server on Go file changes
3. **Generate Code**: Run `make proto-gen` after proto file changes
4. **Test**: Use `make test` for unit tests or API testing tools
5. **Lint**: Run `make proto-lint` to check proto files
6. **Document**: Update Swagger docs with `make swag`

## 🚦 API Endpoints

### REST API (Port 8000)
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST   | `/v1/register` | Register new user | No |
| POST   | `/v1/login` | User login | No |
| GET    | `/v1/user` | Get user profile | Yes |
| PUT    | `/v1/user` | Update user profile | Yes |
| GET    | `/swagger/index.html` | API documentation | No |

### gRPC API (Port 8080)
| Service | Method | Description | Auth Required |
|---------|--------|-------------|---------------|
| UserService | Register | Register new user | No |
| UserService | Login | User login | No |
| UserService | GetProfile | Get user profile | Yes |
| UserService | UpdateProfile | Update user profile | Yes |

## 🔐 Authentication

Both REST and gRPC APIs use JWT tokens for authentication:

### REST API
```bash
# Include in Authorization header
Authorization: Bearer <jwt-token>
```

### gRPC API
```bash
# Include in authorization metadata
authorization: Bearer <jwt-token>
```

## 📊 Monitoring & Observability

The boilerplate includes comprehensive logging and can be extended with:

- **Structured Logging** - JSON logs with Zap
- **Request Tracing** - Request ID correlation
- **Metrics Collection** - Ready for Prometheus integration  
- **Health Checks** - gRPC health check service
- **Error Tracking** - Comprehensive error mapping

## 🤝 Contributing

1. **Follow Architecture**: Maintain clean architecture principles
2. **Update Proto Files**: Use `make proto-gen` after changes
3. **Write Tests**: Include unit tests for new features
4. **Document APIs**: Update Swagger and proto documentation
5. **Lint Code**: Run `make proto-lint` before commits

## 📞 Support

- **Issues**: Report bugs and feature requests in the repository issues
- **Documentation**: All documentation is in the `docs/` directory
- **Examples**: Check `docs/grpc/client-examples.md` for usage examples

## 🔗 Related Links

- **[Go gRPC Documentation](https://grpc.io/docs/languages/go/)**
- **[Protocol Buffers Guide](https://developers.google.com/protocol-buffers)**
- **[Buf Documentation](https://buf.build/docs)**
- **[Gin Framework](https://gin-gonic.com/docs/)**
- **[GORM Documentation](https://gorm.io/docs/)**

---

Happy coding! 🚀