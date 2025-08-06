# gRPC API Documentation

This document provides comprehensive information about the gRPC implementation in the Go Web API Boilerplate.

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)
- [Authentication](#authentication)
- [Error Handling](#error-handling)
- [Client Examples](#client-examples)
- [Development](#development)
- [Best Practices](#best-practices)

## Overview

The Go Web API Boilerplate includes a high-performance gRPC server that runs alongside the HTTP REST API. Both servers share the same business logic layer, ensuring consistency and maintainability.

### Key Features

- **Protocol Buffers**: Type-safe API definitions with comprehensive documentation
- **Dual Protocol Support**: REST and gRPC APIs running simultaneously
- **JWT Authentication**: Secure authentication via gRPC metadata
- **Comprehensive Interceptors**: Logging, recovery, and authentication
- **Error Mapping**: Proper conversion from domain errors to gRPC status codes
- **Modern Tooling**: Buf for protobuf management and code generation

### Server Configuration

- **gRPC Server**: `localhost:8080` (development) / `0.0.0.0:8080` (production)
- **HTTP Server**: `localhost:8000` (development) / `0.0.0.0:8000` (production)
- **Protocol**: HTTP/2 with binary protocol buffers
- **TLS**: Optional (configured via environment)

## Getting Started

### Prerequisites

```bash
# Install required tools
make init
```

### Generate Protocol Buffer Code

```bash
# Generate Go code from .proto files
make proto-gen

# Lint proto files
make proto-lint

# Check for breaking changes
make proto-breaking
```

### Start the Server

```bash
# Development with live reload
make dev

# Manual start
go run ./cmd/server -conf config/local.yml
```

## API Reference

### User Service

The `UserService` provides user authentication and profile management functionality.

#### Service Definition

```protobuf
service UserService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc GetProfile(GetProfileRequest) returns (GetProfileResponse);
  rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse);
}
```

### Methods

#### Register

Creates a new user account.

**Request:**
```protobuf
message RegisterRequest {
  string email = 1;     // User's email address (required)
  string password = 2;  // User's password (required, minimum 6 characters)
}
```

**Response:**
```protobuf
message RegisterResponse {
  string message = 1;   // Success message
}
```

**Example:**
```go
resp, err := client.Register(ctx, &userv1.RegisterRequest{
    Email:    "user@example.com",
    Password: "securepassword",
})
```

#### Login

Authenticates a user and returns an access token.

**Request:**
```protobuf
message LoginRequest {
  string email = 1;     // User's email address (required)
  string password = 2;  // User's password (required)
}
```

**Response:**
```protobuf
message LoginResponse {
  string access_token = 1;  // JWT access token for API authentication
  int64 expires_in = 2;     // Token expiration time in seconds
}
```

**Example:**
```go
resp, err := client.Login(ctx, &userv1.LoginRequest{
    Email:    "user@example.com",
    Password: "securepassword",
})
token := resp.AccessToken
```

#### GetProfile

Retrieves the authenticated user's profile information.

**Request:**
```protobuf
message GetProfileRequest {} // Empty - user ID comes from JWT token
```

**Response:**
```protobuf
message GetProfileResponse {
  string user_id = 1;   // Unique user identifier
  string nickname = 2;  // User's display name
  string email = 3;     // User's email address
}
```

**Example:**
```go
// Requires authentication metadata
resp, err := client.GetProfile(ctx, &userv1.GetProfileRequest{})
```

#### UpdateProfile

Updates the authenticated user's profile information.

**Request:**
```protobuf
message UpdateProfileRequest {
  string nickname = 1;  // User's display name
  string email = 2;     // User's email address (required)
}
```

**Response:**
```protobuf
message UpdateProfileResponse {
  string message = 1;                    // Success message
  GetProfileResponse profile = 2;        // Updated user profile
}
```

**Example:**
```go
resp, err := client.UpdateProfile(ctx, &userv1.UpdateProfileRequest{
    Nickname: "New Nickname",
    Email:    "newemail@example.com",
})
```

## Authentication

gRPC endpoints (except `Register` and `Login`) require JWT authentication via metadata.

### Adding Authentication Metadata

```go
import (
    "google.golang.org/grpc/metadata"
)

// Add Bearer token to context
md := metadata.Pairs("authorization", "Bearer "+token)
ctx := metadata.NewOutgoingContext(context.Background(), md)

// Make authenticated request
resp, err := client.GetProfile(ctx, &userv1.GetProfileRequest{})
```

### Public Endpoints

The following endpoints do not require authentication:
- `UserService.Register`
- `UserService.Login`

All other endpoints require a valid JWT token in the `authorization` metadata.

## Error Handling

gRPC errors follow the standard gRPC status codes and are mapped from domain errors:

### Common Error Codes

| gRPC Code | Description | Common Causes |
|-----------|-------------|---------------|
| `OK` | Success | Request completed successfully |
| `UNAUTHENTICATED` | Authentication failed | Invalid/missing JWT token, wrong credentials |
| `ALREADY_EXISTS` | Resource already exists | Email already registered |
| `INTERNAL` | Internal server error | Server-side errors, panics |
| `INVALID_ARGUMENT` | Invalid request data | Malformed request, validation errors |

### Error Response Example

```go
if err != nil {
    if st, ok := status.FromError(err); ok {
        switch st.Code() {
        case codes.Unauthenticated:
            // Handle authentication error
        case codes.AlreadyExists:
            // Handle resource already exists
        case codes.Internal:
            // Handle server error
        default:
            // Handle other errors
        }
    }
}
```

## Client Examples

### Basic Go Client

```go
package main

import (
    "context"
    "log"
    "time"

    userv1 "your-module/api/grpc/proto/user/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := userv1.NewUserServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()

    // Register a new user
    registerResp, err := client.Register(ctx, &userv1.RegisterRequest{
        Email:    "test@example.com",
        Password: "password123",
    })
    if err != nil {
        log.Fatalf("Register failed: %v", err)
    }
    log.Printf("Register response: %s", registerResp.Message)

    // Login
    loginResp, err := client.Login(ctx, &userv1.LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    })
    if err != nil {
        log.Fatalf("Login failed: %v", err)
    }
    token := loginResp.AccessToken
    log.Printf("Login successful, token expires in: %d seconds", loginResp.ExpiresIn)

    // Get profile (authenticated)
    md := metadata.Pairs("authorization", "Bearer "+token)
    authCtx := metadata.NewOutgoingContext(ctx, md)

    profileResp, err := client.GetProfile(authCtx, &userv1.GetProfileRequest{})
    if err != nil {
        log.Fatalf("GetProfile failed: %v", err)
    }
    log.Printf("Profile: ID=%s, Nickname=%s, Email=%s", 
        profileResp.UserId, profileResp.Nickname, profileResp.Email)
}
```

### Python Client Example

```python
import grpc
from proto.user.v1 import user_pb2, user_pb2_grpc

def main():
    channel = grpc.insecure_channel('localhost:8080')
    client = user_pb2_grpc.UserServiceStub(channel)

    # Register
    register_request = user_pb2.RegisterRequest(
        email="test@example.com",
        password="password123"
    )
    register_response = client.Register(register_request)
    print(f"Register: {register_response.message}")

    # Login
    login_request = user_pb2.LoginRequest(
        email="test@example.com",
        password="password123"
    )
    login_response = client.Login(login_request)
    token = login_response.access_token
    
    # Get profile with authentication
    metadata = [('authorization', f'Bearer {token}')]
    profile_response = client.GetProfile(
        user_pb2.GetProfileRequest(),
        metadata=metadata
    )
    print(f"Profile: {profile_response.user_id}")

if __name__ == '__main__':
    main()
```

## Development

### Protocol Buffer Development

1. **Edit Proto Files**: Modify files in `proto/user/v1/`
2. **Generate Code**: Run `make proto-gen`
3. **Lint**: Run `make proto-lint` to check for issues
4. **Breaking Changes**: Run `make proto-breaking` before merging

### Testing gRPC Services

Use tools like `grpcurl` for testing:

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:8080 list

# Call Register method
grpcurl -plaintext -d '{"email":"test@example.com","password":"password123"}' \
  localhost:8080 user.v1.UserService/Register

# Call Login method  
grpcurl -plaintext -d '{"email":"test@example.com","password":"password123"}' \
  localhost:8080 user.v1.UserService/Login

# Call authenticated method
grpcurl -plaintext -H 'authorization: Bearer YOUR_TOKEN' \
  -d '{}' localhost:8080 user.v1.UserService/GetProfile
```

### Adding New Services

1. **Define Service**: Add to `proto/user/v1/user.proto`
2. **Generate Code**: Run `make proto-gen`
3. **Implement Handler**: Add to `internal/grpc/user.go`
4. **Register Service**: Update `internal/server/grpc.go`
5. **Update Wire**: Modify `cmd/server/wire/wire.go` if needed

## Best Practices

### Protocol Buffers

- Use semantic versioning for proto packages
- Add comprehensive field documentation
- Use appropriate field numbers (1-15 for frequent fields)
- Avoid breaking changes in field names/numbers
- Use `oneof` for mutually exclusive fields

### Error Handling

- Map domain errors to appropriate gRPC status codes
- Provide meaningful error messages
- Use structured logging for debugging
- Handle panics gracefully with recovery interceptor

### Security

- Always validate input in handlers
- Use JWT tokens for authentication
- Implement proper authorization checks
- Log security-relevant events
- Use TLS in production

### Performance

- Keep message sizes reasonable
- Use streaming for large datasets
- Implement connection pooling on client side
- Monitor gRPC metrics
- Use compression when appropriate

### Monitoring

- Log all gRPC requests with timing
- Monitor error rates and status codes
- Track authentication failures
- Use distributed tracing
- Set up alerts for anomalies

---

For more information, see the main [README.md](../../README.md) and [CLAUDE.md](../../CLAUDE.md) files.