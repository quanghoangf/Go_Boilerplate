# gRPC Development Guide

This guide covers development workflows, best practices, and advanced topics for working with gRPC in the Go Web API Boilerplate.

## Table of Contents

- [Development Workflow](#development-workflow)
- [Protocol Buffer Development](#protocol-buffer-development)
- [Code Generation](#code-generation)
- [Testing Strategies](#testing-strategies)
- [Performance Optimization](#performance-optimization)
- [Monitoring and Observability](#monitoring-and-observability)
- [Troubleshooting](#troubleshooting)
- [Advanced Topics](#advanced-topics)

## Development Workflow

### Initial Setup

```bash
# Clone the repository
git clone <repository-url>
cd Go_Boilerplate

# Install all required tools
make init

# Generate Protocol Buffer code
make proto-gen

# Start development servers
make dev
```

### Daily Development

1. **Make Changes**: Edit proto files in `proto/user/v1/`
2. **Generate Code**: Run `make proto-gen` after proto changes
3. **Implement Handlers**: Update `internal/grpc/user.go`
4. **Test Changes**: Use `grpcurl` or write unit tests
5. **Check Compatibility**: Run `make proto-breaking` before commits

### Live Reload Setup

Air automatically watches and restarts the server on changes, including:
- Go source files in `cmd/`, `internal/`, `pkg/`
- Configuration files (`.yml`, `.yaml`)
- **Note**: Proto file changes require manual `make proto-gen`

To auto-regenerate proto files on change, add this to `.air.toml`:

```toml
[build]
  cmd = "make proto-gen && go build -o ./tmp/main ./cmd/server"
```

## Protocol Buffer Development

### File Structure

```
proto/
└── user/
    └── v1/
        └── user.proto
```

### Proto File Best Practices

#### 1. Package Naming
```protobuf
syntax = "proto3";

package user.v1;  // Use semantic versioning

option go_package = "boilerplate/api/grpc/user/v1;userv1";
```

#### 2. Field Numbering
```protobuf
message UserRequest {
  // Fields 1-15: Use for frequently set fields (1 byte encoding)
  string email = 1;
  string password = 2;
  
  // Fields 16+: Use for less frequently set fields (2+ byte encoding)
  string optional_field = 16;
}
```

#### 3. Field Evolution
```protobuf
message User {
  string id = 1;
  string email = 2;
  string name = 3;
  
  // Adding new fields (backward compatible)
  string nickname = 4;           // OK: new optional field
  repeated string tags = 5;      // OK: new repeated field
  
  // NEVER do this:
  // int32 email = 2;            // BAD: changed field type
  // string phone = 1;           // BAD: reused field number
}
```

#### 4. Documentation
```protobuf
// UserService provides user authentication and profile management
service UserService {
  // Register creates a new user account
  // 
  // Returns:
  //   - ALREADY_EXISTS: Email is already registered
  //   - INVALID_ARGUMENT: Invalid email format or weak password
  rpc Register(RegisterRequest) returns (RegisterResponse);
}

// RegisterRequest contains user registration data
message RegisterRequest {
  // User's email address (required)
  // Must be a valid email format
  string email = 1;
  
  // User's password (required, minimum 6 characters)
  // Should contain letters, numbers, and special characters
  string password = 2;
}
```

#### 5. Validation Patterns
```protobuf
message CreateUserRequest {
  // Use clear naming for required vs optional fields
  string email = 1;              // required in business logic
  string password = 2;           // required in business logic
  string nickname = 3;           // optional
  
  // Use oneof for mutually exclusive fields
  oneof contact_method {
    string phone = 10;
    string telegram = 11;
  }
}
```

### Service Design Patterns

#### RESTful gRPC Design
```protobuf
service UserService {
  // Collection operations
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc CreateUser(CreateUserRequest) returns (User);
  
  // Resource operations
  rpc GetUser(GetUserRequest) returns (User);
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty);
  
  // Custom operations
  rpc ActivateUser(ActivateUserRequest) returns (User);
}
```

#### Request/Response Patterns
```protobuf
// Standard CRUD operations
message GetUserRequest {
  string user_id = 1;
}

message ListUsersRequest {
  int32 page_size = 1;      // Pagination
  string page_token = 2;
  string filter = 3;        // Filtering
}

message ListUsersResponse {
  repeated User users = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

// Batch operations
message BatchGetUsersRequest {
  repeated string user_ids = 1;
}

message BatchGetUsersResponse {
  repeated User users = 1;
}
```

## Code Generation

### Buf Configuration

#### buf.yaml
```yaml
version: v1
breaking:
  use:
    - FILE
  ignore:
    - internal/legacy  # Ignore legacy files from breaking change detection
lint:
  use:
    - DEFAULT
  except:
    - UNARY_RPC       # Allow unary RPCs without streaming
  ignore:
    - internal/legacy
```

#### buf.gen.yaml
```yaml
version: v2
plugins:
  # Go code generation
  - local: protoc-gen-go
    out: api/grpc
    opt:
      - paths=source_relative
  
  # gRPC Go code generation
  - local: protoc-gen-go-grpc
    out: api/grpc
    opt:
      - paths=source_relative
  
  # Validation (optional)
  - local: protoc-gen-validate
    out: api/grpc
    opt:
      - paths=source_relative
      - lang=go
```

### Custom Code Generation

#### Makefile Extensions
```makefile
# Custom proto generation with validation
.PHONY: proto-gen-validate
proto-gen-validate:
	buf generate --template buf.gen.validate.yaml

# Generate mocks for gRPC services
.PHONY: proto-mock
proto-mock:
	mockgen -source=api/grpc/proto/user/v1/user_grpc.pb.go \
		-destination=test/mocks/grpc/user_service_mock.go

# Clean generated files
.PHONY: proto-clean
proto-clean:
	rm -rf api/grpc/proto/
```

### Multi-Language Generation

```yaml
# buf.gen.multi.yaml
version: v2
plugins:
  # Go
  - local: protoc-gen-go
    out: api/grpc/go
    opt: [paths=source_relative]
  
  # Python
  - local: protoc-gen-python
    out: api/grpc/python
  
  # TypeScript
  - remote: buf.build/connectrpc/es:v1.4.0
    out: api/grpc/typescript
    opt: [target=ts]
```

## Testing Strategies

### Unit Testing

#### Testing gRPC Handlers
```go
func TestUserGRPCServer_Register(t *testing.T) {
    tests := []struct {
        name           string
        request        *userv1.RegisterRequest
        mockSetup      func(*service.MockUserService)
        expectedError  codes.Code
        expectedResp   *userv1.RegisterResponse
    }{
        {
            name: "successful registration",
            request: &userv1.RegisterRequest{
                Email:    "test@example.com",
                Password: "password123",
            },
            mockSetup: func(m *service.MockUserService) {
                m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil)
            },
            expectedResp: &userv1.RegisterResponse{
                Message: "User registered successfully",
            },
        },
        {
            name: "email already exists",
            request: &userv1.RegisterRequest{
                Email:    "existing@example.com",
                Password: "password123",
            },
            mockSetup: func(m *service.MockUserService) {
                m.EXPECT().Register(gomock.Any(), gomock.Any()).
                    Return(v1.ErrEmailAlreadyUse)
            },
            expectedError: codes.AlreadyExists,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockService := service.NewMockUserService(ctrl)
            tt.mockSetup(mockService)

            server := grpc.NewUserGRPCServer(mockService, logger)
            
            resp, err := server.Register(context.Background(), tt.request)

            if tt.expectedError != codes.OK {
                assert.Error(t, err)
                st, ok := status.FromError(err)
                assert.True(t, ok)
                assert.Equal(t, tt.expectedError, st.Code())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResp.Message, resp.Message)
            }
        })
    }
}
```

#### Testing with bufconn
```go
func TestUserGRPCServer_Integration(t *testing.T) {
    // Create buffer connection
    lis := bufconn.Listen(1024 * 1024)
    
    // Setup test server
    s := grpc.NewServer()
    userGRPCServer := setupTestUserGRPCServer() // Your test setup
    userv1.RegisterUserServiceServer(s, userGRPCServer)
    
    go func() {
        if err := s.Serve(lis); err != nil {
            t.Errorf("Server exited with error: %v", err)
        }
    }()
    defer s.GracefulStop()

    // Create client
    conn, err := grpc.DialContext(
        context.Background(),
        "bufnet",
        grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
            return lis.Dial()
        }),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    require.NoError(t, err)
    defer conn.Close()

    client := userv1.NewUserServiceClient(conn)

    // Test full workflow
    ctx := context.Background()
    
    // Register
    _, err = client.Register(ctx, &userv1.RegisterRequest{
        Email:    "test@example.com",
        Password: "password123",
    })
    assert.NoError(t, err)

    // Login
    loginResp, err := client.Login(ctx, &userv1.LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, loginResp.AccessToken)
}
```

### Integration Testing

#### Full Server Testing
```go
func TestGRPCServer_FullStack(t *testing.T) {
    // Start test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Start gRPC server
    server := startTestGRPCServer(t, db)
    defer server.Stop()

    // Create client
    conn, err := grpc.Dial("localhost:8080", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)
    defer conn.Close()

    client := userv1.NewUserServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Test complete user workflow
    email := fmt.Sprintf("test-%d@example.com", time.Now().Unix())
    
    // Register user
    _, err = client.Register(ctx, &userv1.RegisterRequest{
        Email:    email,
        Password: "password123",
    })
    assert.NoError(t, err)

    // Verify user in database
    var count int
    db.Model(&model.User{}).Where("email = ?", email).Count(&count)
    assert.Equal(t, 1, count)
}
```

### Load Testing

#### Basic Load Test
```go
func TestGRPCServer_LoadTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }

    conn, err := grpc.Dial("localhost:8080", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)
    defer conn.Close()

    client := userv1.NewUserServiceClient(conn)

    // Setup: Login to get token
    loginResp, err := client.Login(context.Background(), &userv1.LoginRequest{
        Email:    "load-test@example.com",
        Password: "password123",
    })
    require.NoError(t, err)

    // Load test parameters
    concurrency := 50
    requests := 1000
    requestsPerWorker := requests / concurrency

    var wg sync.WaitGroup
    var totalDuration int64
    var errors int64

    start := time.Now()

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            md := metadata.Pairs("authorization", "Bearer "+loginResp.AccessToken)
            
            for j := 0; j < requestsPerWorker; j++ {
                ctx := metadata.NewOutgoingContext(context.Background(), md)
                ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
                
                reqStart := time.Now()
                _, err := client.GetProfile(ctx, &userv1.GetProfileRequest{})
                reqDuration := time.Since(reqStart)
                cancel()

                atomic.AddInt64(&totalDuration, reqDuration.Nanoseconds())
                if err != nil {
                    atomic.AddInt64(&errors, 1)
                }
            }
        }()
    }

    wg.Wait()
    totalTestDuration := time.Since(start)

    // Calculate metrics
    avgLatency := time.Duration(atomic.LoadInt64(&totalDuration) / int64(requests))
    rps := float64(requests) / totalTestDuration.Seconds()
    errorRate := float64(atomic.LoadInt64(&errors)) / float64(requests)

    t.Logf("Load test results:")
    t.Logf("  Requests: %d", requests)
    t.Logf("  Concurrency: %d", concurrency)
    t.Logf("  Total duration: %v", totalTestDuration)
    t.Logf("  RPS: %.2f", rps)
    t.Logf("  Average latency: %v", avgLatency)
    t.Logf("  Error rate: %.2f%%", errorRate*100)

    // Assertions
    assert.Less(t, errorRate, 0.01, "Error rate should be less than 1%")
    assert.Less(t, avgLatency, 100*time.Millisecond, "Average latency should be less than 100ms")
    assert.Greater(t, rps, 100.0, "RPS should be greater than 100")
}
```

## Performance Optimization

### Connection Pooling

#### Client-Side Connection Pool
```go
type ConnectionPool struct {
    connections []*grpc.ClientConn
    clients     []userv1.UserServiceClient
    current     int64
    mu          sync.RWMutex
}

func NewConnectionPool(address string, size int) (*ConnectionPool, error) {
    pool := &ConnectionPool{
        connections: make([]*grpc.ClientConn, size),
        clients:     make([]userv1.UserServiceClient, size),
    }

    for i := 0; i < size; i++ {
        conn, err := grpc.Dial(address,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
            grpc.WithKeepaliveParams(keepalive.ClientParameters{
                Time:                10 * time.Second,
                Timeout:             time.Second,
                PermitWithoutStream: true,
            }),
            grpc.WithDefaultCallOptions(
                grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
                grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
            ),
        )
        if err != nil {
            // Cleanup on error
            for j := 0; j < i; j++ {
                pool.connections[j].Close()
            }
            return nil, err
        }

        pool.connections[i] = conn
        pool.clients[i] = userv1.NewUserServiceClient(conn)
    }

    return pool, nil
}

func (p *ConnectionPool) GetClient() userv1.UserServiceClient {
    index := atomic.AddInt64(&p.current, 1) % int64(len(p.clients))
    return p.clients[index]
}
```

#### Server-Side Optimization
```go
func NewOptimizedGRPCServer() *grpc.Server {
    return grpc.NewServer(
        // Connection limits
        grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
        grpc.MaxSendMsgSize(4*1024*1024), // 4MB
        grpc.MaxConcurrentStreams(1000),
        
        // Keepalive settings
        grpc.KeepaliveParams(keepalive.ServerParameters{
            MaxConnectionIdle:     15 * time.Second,
            MaxConnectionAge:      30 * time.Second,
            MaxConnectionAgeGrace: 5 * time.Second,
            Time:                  5 * time.Second,
            Timeout:               1 * time.Second,
        }),
        grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
            MinTime:             5 * time.Second,
            PermitWithoutStream: false,
        }),
    )
}
```

### Message Optimization

#### Efficient Message Design
```protobuf
// Good: Minimize message size
message UserSummary {
  string id = 1;
  string name = 2;
  int64 last_login = 3;  // Use int64 for timestamps
}

// Better: Use field presence for optionals
message UpdateUserRequest {
  string user_id = 1;
  
  // Use optional fields for partial updates
  optional string name = 2;
  optional string email = 3;
  optional bool active = 4;
}
```

#### Compression
```go
// Enable compression
conn, err := grpc.Dial(address,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithCompressor(grpc.NewGZIPCompressor()),
    grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
)
```

## Monitoring and Observability

### Metrics Collection

#### Prometheus Metrics
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    grpcRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests",
        },
        []string{"method", "status"},
    )

    grpcRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "grpc_request_duration_seconds",
            Help: "Duration of gRPC requests",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~32s
        },
        []string{"method"},
    )
)

func MetricsInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        resp, err := handler(ctx, req)
        
        duration := time.Since(start)
        method := info.FullMethod
        status := "success"
        if err != nil {
            status = "error"
        }

        grpcRequestsTotal.WithLabelValues(method, status).Inc()
        grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())

        return resp, err
    }
}
```

#### OpenTelemetry Integration
```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func NewTracedGRPCServer() *grpc.Server {
    return grpc.NewServer(
        grpc.StatsHandler(otelgrpc.NewServerHandler()),
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )
}
```

### Health Checks

#### gRPC Health Check Service
```go
import (
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)

func RegisterHealthService(s *grpc.Server) {
    healthServer := health.NewServer()
    grpc_health_v1.RegisterHealthServer(s, healthServer)
    
    // Set service status
    healthServer.SetServingStatus("user.v1.UserService", grpc_health_v1.HealthCheckResponse_SERVING)
}
```

#### Custom Health Checks
```go
type HealthChecker struct {
    db     *gorm.DB
    logger *log.Logger
}

func (h *HealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
    // Check database connectivity
    if err := h.db.Exec("SELECT 1").Error; err != nil {
        h.logger.Sugar().Errorf("Database health check failed: %v", err)
        return &grpc_health_v1.HealthCheckResponse{
            Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
        }, nil
    }

    return &grpc_health_v1.HealthCheckResponse{
        Status: grpc_health_v1.HealthCheckResponse_SERVING,
    }, nil
}
```

### Logging Best Practices

#### Structured Logging
```go
func LoggingInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        // Extract metadata for correlation
        md, _ := metadata.FromIncomingContext(ctx)
        requestID := extractRequestID(md)
        userID := extractUserID(ctx)

        resp, err := handler(ctx, req)
        
        duration := time.Since(start)
        
        fields := []interface{}{
            "method", info.FullMethod,
            "duration", duration,
            "request_id", requestID,
        }
        
        if userID != "" {
            fields = append(fields, "user_id", userID)
        }

        if err != nil {
            st, _ := status.FromError(err)
            fields = append(fields, "error", err.Error(), "grpc_code", st.Code())
            logger.Sugar().Errorw("gRPC request failed", fields...)
        } else {
            logger.Sugar().Infow("gRPC request completed", fields...)
        }

        return resp, err
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Connection Issues
```bash
# Check if server is running
grpcurl -plaintext localhost:8080 list

# Test with health check
grpcurl -plaintext localhost:8080 grpc.health.v1.Health/Check

# Debug connection with verbose output
grpcurl -plaintext -v localhost:8080 list
```

#### 2. Authentication Issues
```bash
# Test with invalid token
grpcurl -plaintext \
  -H 'authorization: Bearer invalid-token' \
  -d '{}' \
  localhost:8080 user.v1.UserService/GetProfile

# Check token format
grpcurl -plaintext \
  -H 'authorization: invalid-format' \
  -d '{}' \
  localhost:8080 user.v1.UserService/GetProfile
```

#### 3. Message Size Issues
```go
// Increase message size limits
conn, err := grpc.Dial(address,
    grpc.WithDefaultCallOptions(
        grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
        grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
    ),
)
```

### Debugging Tools

#### Enable gRPC Logging
```bash
# Enable all gRPC logs
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info

# Run your application
go run ./cmd/server
```

#### Use grpc_cli
```bash
# Install grpc_cli (C++ implementation)
# More features than grpcurl

# List services
grpc_cli ls localhost:8080

# Call with debugging
grpc_cli call localhost:8080 user.v1.UserService.Login \
  "email:'test@example.com' password:'password123'" \
  --enable_log_reporter
```

## Advanced Topics

### Custom Interceptors

#### Rate Limiting Interceptor
```go
import (
    "golang.org/x/time/rate"
)

type RateLimitInterceptor struct {
    limiter *rate.Limiter
    logger  *log.Logger
}

func NewRateLimitInterceptor(rps int, logger *log.Logger) *RateLimitInterceptor {
    return &RateLimitInterceptor{
        limiter: rate.NewLimiter(rate.Limit(rps), rps*2), // Allow burst
        logger:  logger,
    }
}

func (r *RateLimitInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        if !r.limiter.Allow() {
            r.logger.Sugar().Warnf("Rate limit exceeded for method %s", info.FullMethod)
            return nil, status.Error(codes.ResourceExhausted, "Rate limit exceeded")
        }
        return handler(ctx, req)
    }
}
```

#### Circuit Breaker (Client-Side)
```go
import (
    "github.com/sony/gobreaker"
)

type CircuitBreakerInterceptor struct {
    cb *gobreaker.CircuitBreaker
}

func NewCircuitBreakerInterceptor() *CircuitBreakerInterceptor {
    settings := gobreaker.Settings{
        Name:        "grpc-client",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures > 3
        },
    }

    return &CircuitBreakerInterceptor{
        cb: gobreaker.NewCircuitBreaker(settings),
    }
}

func (c *CircuitBreakerInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        _, err := c.cb.Execute(func() (interface{}, error) {
            return nil, invoker(ctx, method, req, reply, cc, opts...)
        })
        return err
    }
}
```

### Streaming RPCs

#### Server Streaming
```protobuf
service UserService {
  // Stream user updates
  rpc WatchUsers(WatchUsersRequest) returns (stream User);
}
```

```go
func (s *UserGRPCServer) WatchUsers(req *userv1.WatchUsersRequest, stream userv1.UserService_WatchUsersServer) error {
    // Implementation for server streaming
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-stream.Context().Done():
            return stream.Context().Err()
        case <-ticker.C:
            // Send update
            user := &userv1.User{
                UserId:   "123",
                Nickname: "Updated Name",
                Email:    "updated@example.com",
            }
            if err := stream.Send(user); err != nil {
                return err
            }
        }
    }
}
```

#### Bidirectional Streaming
```protobuf
service ChatService {
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}
```

```go
func (s *ChatGRPCServer) Chat(stream chatv1.ChatService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }

        // Process message and send response
        response := &chatv1.ChatMessage{
            User:    "server",
            Content: fmt.Sprintf("Echo: %s", msg.Content),
        }

        if err := stream.Send(response); err != nil {
            return err
        }
    }
}
```

### Security Enhancements

#### TLS Configuration
```go
func NewSecureGRPCServer(certFile, keyFile string) (*grpc.Server, error) {
    creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
    if err != nil {
        return nil, err
    }

    return grpc.NewServer(grpc.Creds(creds)), nil
}
```

#### mTLS (Mutual TLS)
```go
func NewMutualTLSServer(certFile, keyFile, caFile string) (*grpc.Server, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }

    caCert, err := ioutil.ReadFile(caFile)
    if err != nil {
        return nil, err
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    caCertPool,
    }

    creds := credentials.NewTLS(tlsConfig)
    return grpc.NewServer(grpc.Creds(creds)), nil
}
```

---

This guide covers the essential aspects of gRPC development in our boilerplate. For more specific examples and use cases, refer to the other documentation files in this directory.