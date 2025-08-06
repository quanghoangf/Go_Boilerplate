# gRPC Interceptors Documentation

This document describes the gRPC interceptors implemented in the Go Web API Boilerplate for cross-cutting concerns like authentication, logging, and error recovery.

## Overview

Interceptors in gRPC are similar to middleware in HTTP frameworks. They allow you to intercept and process RPC calls before they reach the actual service handler. Our implementation includes three key interceptors that are chained together to provide a robust request processing pipeline.

## Interceptor Chain

The interceptors are applied in the following order:

1. **Recovery Interceptor** (outermost) - Catches panics
2. **Logging Interceptor** - Logs request/response details  
3. **Authentication Interceptor** (innermost) - Validates JWT tokens

## Recovery Interceptor

**File**: `internal/grpc/interceptor/recovery.go`

### Purpose
Catches and handles panics that occur during gRPC request processing, preventing the server from crashing and providing graceful error responses to clients.

### Features
- Catches any panic in the request processing chain
- Logs panic details with full stack trace for debugging
- Returns a proper gRPC `INTERNAL` error to the client
- Prevents server crashes from unhandled panics

### Implementation Details

```go
type RecoveryInterceptor struct {
    logger *log.Logger
}

func (r *RecoveryInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
        defer func() {
            if panicVal := recover(); panicVal != nil {
                // Log panic with stack trace
                r.logger.Sugar().Errorf(
                    "gRPC panic recovered in %s: %v\nStack trace:\n%s",
                    info.FullMethod,
                    panicVal,
                    string(debug.Stack()),
                )
                
                // Return gRPC internal error
                err = status.Error(codes.Internal, "Internal server error")
            }
        }()
        
        return handler(ctx, req)
    }
}
```

### Error Response
When a panic is caught, clients receive:
- **Status Code**: `codes.Internal` (13)
- **Message**: "Internal server error"
- **Details**: None (for security reasons)

### Logging Output
```
ERROR gRPC panic recovered in /user.v1.UserService/GetProfile: runtime error: nil pointer dereference
Stack trace:
goroutine 123 [running]:
internal/grpc/interceptor.(*RecoveryInterceptor).UnaryInterceptor.func1.1()
    /app/internal/grpc/interceptor/recovery.go:32 +0x123
...
```

## Logging Interceptor

**File**: `internal/grpc/interceptor/logging.go`

### Purpose
Provides structured logging for all gRPC requests, including timing, status codes, and error information for monitoring and debugging purposes.

### Features
- Logs every gRPC request with timing information
- Records gRPC status codes (success and error)
- Integrates with Zap structured logging
- Provides request correlation for debugging
- Measures and logs request duration

### Implementation Details

```go
type LoggingInterceptor struct {
    logger *log.Logger
}

func (l *LoggingInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        // Call the handler
        resp, err := handler(ctx, req)
        
        // Calculate duration
        duration := time.Since(start)
        
        // Determine status code
        statusCode := codes.OK
        if err != nil {
            if st, ok := status.FromError(err); ok {
                statusCode = st.Code()
            } else {
                statusCode = codes.Internal
            }
        }
        
        // Log the request
        l.logger.Sugar().Infof(
            "gRPC %s %v %s %v",
            info.FullMethod,
            statusCode,
            duration,
            l.getErrorMessage(err),
        )
        
        return resp, err
    }
}
```

### Log Output Examples

**Successful Request:**
```
INFO gRPC /user.v1.UserService/Login OK 45.2ms
```

**Error Request:**
```
INFO gRPC /user.v1.UserService/GetProfile UNAUTHENTICATED 1.2ms Invalid token
```

**High Latency Request:**
```
INFO gRPC /user.v1.UserService/UpdateProfile OK 1.234s
```

### Metrics Integration
The logging interceptor can be extended to emit metrics:

```go
// Example metrics integration (not implemented by default)
func (l *LoggingInterceptor) emitMetrics(method string, code codes.Code, duration time.Duration) {
    // Prometheus metrics
    grpcRequestTotal.WithLabelValues(method, code.String()).Inc()
    grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}
```

## Authentication Interceptor

**File**: `internal/grpc/interceptor/auth.go`

### Purpose
Validates JWT tokens for protected gRPC endpoints and extracts user information for use in request handlers.

### Features
- JWT token validation using the `authorization` metadata
- Selective authentication (public vs. protected endpoints)
- User ID extraction and context propagation
- Proper gRPC authentication error responses
- Support for Bearer token format

### Implementation Details

```go
type AuthInterceptor struct {
    jwt    *jwt.JWT
    logger *log.Logger
}

func (a *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Skip authentication for public methods
        if a.isPublicMethod(info.FullMethod) {
            return handler(ctx, req)
        }
        
        // Extract and validate token
        token, err := a.extractToken(ctx)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, err.Error())
        }
        
        claims, err := a.jwt.ParseToken(token)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "Invalid token")
        }
        
        // Add user ID to context
        ctx = context.WithValue(ctx, "userID", claims.UserId)
        
        return handler(ctx, req)
    }
}
```

### Token Extraction

The interceptor expects the JWT token in the `authorization` metadata with the `Bearer ` prefix:

```go
func (a *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", status.Error(codes.Unauthenticated, "metadata not found")
    }
    
    values := md.Get("authorization")
    if len(values) == 0 {
        return "", status.Error(codes.Unauthenticated, "authorization header not found")
    }
    
    authHeader := values[0]
    const bearerPrefix = "Bearer "
    if !strings.HasPrefix(authHeader, bearerPrefix) {
        return "", status.Error(codes.Unauthenticated, "invalid authorization header format")
    }
    
    return strings.TrimPrefix(authHeader, bearerPrefix), nil
}
```

### Public Endpoints

The following methods bypass authentication:

```go
func (a *AuthInterceptor) isPublicMethod(method string) bool {
    publicMethods := []string{
        "/user.v1.UserService/Register",
        "/user.v1.UserService/Login",
    }
    
    for _, publicMethod := range publicMethods {
        if method == publicMethod {
            return true
        }
    }
    
    return false
}
```

### Context Usage in Handlers

Once authenticated, handlers can access the user ID from context:

```go
func (s *UserGRPCServer) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
    userID, ok := ctx.Value("userID").(string)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "User ID not found in context")
    }
    
    // Use userID for business logic...
}
```

### Authentication Errors

| Error Case | Status Code | Message |
|------------|-------------|---------|
| Missing metadata | `UNAUTHENTICATED` | "metadata not found" |
| Missing authorization header | `UNAUTHENTICATED` | "authorization header not found" |
| Invalid header format | `UNAUTHENTICATED` | "invalid authorization header format" |
| Invalid/expired token | `UNAUTHENTICATED` | "Invalid token" |

## Interceptor Configuration

**File**: `internal/server/grpc.go`

The interceptors are configured and chained in the gRPC server setup:

```go
func NewGRPCServer(logger *log.Logger, conf *viper.Viper, jwt *jwt.JWT, userService service.UserService) *grpcServer.Server {
    // Create interceptors
    authInterceptor := interceptor.NewAuthInterceptor(jwt, logger)
    loggingInterceptor := interceptor.NewLoggingInterceptor(logger)
    recoveryInterceptor := interceptor.NewRecoveryInterceptor(logger)
    
    // Create gRPC server with chained interceptors
    s := grpcServer.NewServer(
        logger,
        grpcServer.WithServerHost(conf.GetString("grpc.host")),
        grpcServer.WithServerPort(conf.GetInt("grpc.port")),
        grpcServer.WithUnaryInterceptors(
            recoveryInterceptor.UnaryInterceptor(),  // Outermost
            loggingInterceptor.UnaryInterceptor(),   // Middle
            authInterceptor.UnaryInterceptor(),      // Innermost
        ),
    )
    
    return s
}
```

## Testing Interceptors

### Unit Testing

```go
func TestRecoveryInterceptor(t *testing.T) {
    logger := log.NewLog(config)
    interceptor := NewRecoveryInterceptor(logger)
    
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        panic("test panic")
    }
    
    _, err := interceptor.UnaryInterceptor()(
        context.Background(), 
        nil, 
        &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, 
        handler,
    )
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "Internal server error")
}
```

### Integration Testing

```go
func TestAuthInterceptor_Integration(t *testing.T) {
    // Setup test server with interceptors
    server := setupTestGRPCServer()
    defer server.Stop()
    
    // Create client connection
    conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
    require.NoError(t, err)
    defer conn.Close()
    
    client := userv1.NewUserServiceClient(conn)
    
    // Test unauthenticated request
    _, err = client.GetProfile(context.Background(), &userv1.GetProfileRequest{})
    assert.Error(t, err)
    assert.Equal(t, codes.Unauthenticated, status.Code(err))
    
    // Test authenticated request
    md := metadata.Pairs("authorization", "Bearer "+validToken)
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    _, err = client.GetProfile(ctx, &userv1.GetProfileRequest{})
    assert.NoError(t, err)
}
```

## Performance Considerations

### Logging Impact
- Structured logging adds ~1-2ms per request
- Use appropriate log levels in production
- Consider sampling for high-traffic endpoints

### Authentication Overhead  
- JWT parsing adds ~0.5ms per request
- Token caching can reduce this overhead
- Consider connection-level authentication for trusted clients

### Recovery Overhead
- Minimal impact (~0.1ms) when no panics occur
- Stack trace generation is expensive during panics
- Consider panic frequency monitoring

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Recovery Interceptor**
   - Panic frequency (should be near zero)
   - Recovery response times
   - Error rate after recovery

2. **Logging Interceptor**
   - Request latency percentiles (p50, p95, p99)
   - Status code distribution
   - Error rates by method

3. **Authentication Interceptor**
   - Authentication failure rate
   - Token validation latency
   - Blocked request frequency

### Recommended Alerts

```yaml
# High panic rate
- alert: HighGRPCPanicRate
  expr: rate(grpc_panics_recovered_total[5m]) > 0.01
  
# High authentication failure rate  
- alert: HighGRPCAuthFailureRate
  expr: rate(grpc_requests_total{code="UNAUTHENTICATED"}[5m]) > 0.1

# High latency
- alert: HighGRPCLatency
  expr: histogram_quantile(0.95, grpc_request_duration_seconds) > 1.0
```

## Extending Interceptors

### Adding Custom Interceptors

1. **Create New Interceptor**:
```go
type RateLimitInterceptor struct {
    limiter *rate.Limiter
}

func (r *RateLimitInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        if !r.limiter.Allow() {
            return nil, status.Error(codes.ResourceExhausted, "Rate limit exceeded")
        }
        return handler(ctx, req)
    }
}
```

2. **Add to Chain**:
```go
s := grpcServer.NewServer(
    logger,
    grpcServer.WithUnaryInterceptors(
        recoveryInterceptor.UnaryInterceptor(),
        rateLimitInterceptor.UnaryInterceptor(), // Add here
        loggingInterceptor.UnaryInterceptor(),
        authInterceptor.UnaryInterceptor(),
    ),
)
```

### Stream Interceptors

For streaming RPCs, implement `grpc.StreamServerInterceptor`:

```go
func (l *LoggingInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
    return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        start := time.Now()
        err := handler(srv, ss)
        duration := time.Since(start)
        
        l.logger.Sugar().Infof("gRPC Stream %s %v", info.FullMethod, duration)
        return err
    }
}
```

---

For more information about gRPC implementation, see the main [gRPC README](README.md).