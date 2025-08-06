# gRPC Client Examples

This document provides comprehensive examples for connecting to and using the gRPC API in various programming languages and scenarios.

## Table of Contents

- [Go Client](#go-client)
- [Python Client](#python-client)
- [Node.js Client](#nodejs-client)
- [Java Client](#java-client)
- [Command Line Tools](#command-line-tools)
- [Authentication Examples](#authentication-examples)
- [Error Handling](#error-handling)
- [Testing Examples](#testing-examples)

## Go Client

### Basic Client Implementation

```go
package main

import (
    "context"
    "log"
    "time"

    userv1 "boilerplate/api/grpc/proto/user/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
)

type UserClient struct {
    client userv1.UserServiceClient
    conn   *grpc.ClientConn
}

func NewUserClient(address string) (*UserClient, error) {
    conn, err := grpc.Dial(address, 
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        return nil, err
    }

    return &UserClient{
        client: userv1.NewUserServiceClient(conn),
        conn:   conn,
    }, nil
}

func (c *UserClient) Close() error {
    return c.conn.Close()
}

func (c *UserClient) Register(ctx context.Context, email, password string) error {
    req := &userv1.RegisterRequest{
        Email:    email,
        Password: password,
    }

    resp, err := c.client.Register(ctx, req)
    if err != nil {
        return err
    }

    log.Printf("Registration successful: %s", resp.Message)
    return nil
}

func (c *UserClient) Login(ctx context.Context, email, password string) (string, error) {
    req := &userv1.LoginRequest{
        Email:    email,
        Password: password,
    }

    resp, err := c.client.Login(ctx, req)
    if err != nil {
        return "", err
    }

    log.Printf("Login successful, token expires in %d seconds", resp.ExpiresIn)
    return resp.AccessToken, nil
}

func (c *UserClient) GetProfile(ctx context.Context, token string) (*userv1.GetProfileResponse, error) {
    md := metadata.Pairs("authorization", "Bearer "+token)
    authCtx := metadata.NewOutgoingContext(ctx, md)

    return c.client.GetProfile(authCtx, &userv1.GetProfileRequest{})
}

func (c *UserClient) UpdateProfile(ctx context.Context, token, nickname, email string) (*userv1.UpdateProfileResponse, error) {
    md := metadata.Pairs("authorization", "Bearer "+token)
    authCtx := metadata.NewOutgoingContext(ctx, md)

    req := &userv1.UpdateProfileRequest{
        Nickname: nickname,
        Email:    email,
    }

    return c.client.UpdateProfile(authCtx, req)
}

func main() {
    client, err := NewUserClient("localhost:8080")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()

    // Register
    if err := client.Register(ctx, "user@example.com", "password123"); err != nil {
        if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
            log.Println("User already exists, proceeding to login")
        } else {
            log.Fatalf("Registration failed: %v", err)
        }
    }

    // Login
    token, err := client.Login(ctx, "user@example.com", "password123")
    if err != nil {
        log.Fatalf("Login failed: %v", err)
    }

    // Get profile
    profile, err := client.GetProfile(ctx, token)
    if err != nil {
        log.Fatalf("Get profile failed: %v", err)
    }
    log.Printf("Profile: ID=%s, Nickname=%s, Email=%s", 
        profile.UserId, profile.Nickname, profile.Email)

    // Update profile
    updateResp, err := client.UpdateProfile(ctx, token, "New Nickname", "newemail@example.com")
    if err != nil {
        log.Fatalf("Update profile failed: %v", err)
    }
    log.Printf("Update successful: %s", updateResp.Message)
}
```

### Connection Pool Example

```go
package main

import (
    "context"
    "sync"
    "time"

    userv1 "boilerplate/api/grpc/proto/user/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/keepalive"
)

type ConnectionPool struct {
    connections []*grpc.ClientConn
    clients     []userv1.UserServiceClient
    current     int
    mutex       sync.RWMutex
}

func NewConnectionPool(address string, poolSize int) (*ConnectionPool, error) {
    pool := &ConnectionPool{
        connections: make([]*grpc.ClientConn, poolSize),
        clients:     make([]userv1.UserServiceClient, poolSize),
    }

    for i := 0; i < poolSize; i++ {
        conn, err := grpc.Dial(address,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
            grpc.WithKeepaliveParams(keepalive.ClientParameters{
                Time:                10 * time.Second,
                Timeout:             time.Second,
                PermitWithoutStream: true,
            }),
        )
        if err != nil {
            // Cleanup existing connections
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
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    client := p.clients[p.current]
    p.current = (p.current + 1) % len(p.clients)
    return client
}

func (p *ConnectionPool) Close() {
    for _, conn := range p.connections {
        conn.Close()
    }
}
```

## Python Client

### Basic Python Client

```python
import grpc
from concurrent import futures
import asyncio
from typing import Optional

# Import generated protobuf files
from proto.user.v1 import user_pb2, user_pb2_grpc

class UserGRPCClient:
    def __init__(self, address: str = "localhost:8080"):
        self.channel = grpc.insecure_channel(address)
        self.client = user_pb2_grpc.UserServiceStub(self.channel)
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
    
    def close(self):
        self.channel.close()
    
    def register(self, email: str, password: str) -> str:
        """Register a new user"""
        request = user_pb2.RegisterRequest(
            email=email,
            password=password
        )
        
        try:
            response = self.client.Register(request)
            return response.message
        except grpc.RpcError as e:
            print(f"Registration failed: {e.code()} - {e.details()}")
            raise
    
    def login(self, email: str, password: str) -> tuple[str, int]:
        """Login and return access token and expiry"""
        request = user_pb2.LoginRequest(
            email=email,
            password=password
        )
        
        try:
            response = self.client.Login(request)
            return response.access_token, response.expires_in
        except grpc.RpcError as e:
            print(f"Login failed: {e.code()} - {e.details()}")
            raise
    
    def get_profile(self, token: str) -> user_pb2.GetProfileResponse:
        """Get user profile"""
        metadata = [('authorization', f'Bearer {token}')]
        request = user_pb2.GetProfileRequest()
        
        try:
            response = self.client.GetProfile(request, metadata=metadata)
            return response
        except grpc.RpcError as e:
            print(f"Get profile failed: {e.code()} - {e.details()}")
            raise
    
    def update_profile(self, token: str, nickname: str, email: str) -> user_pb2.UpdateProfileResponse:
        """Update user profile"""
        metadata = [('authorization', f'Bearer {token}')]
        request = user_pb2.UpdateProfileRequest(
            nickname=nickname,
            email=email
        )
        
        try:
            response = self.client.UpdateProfile(request, metadata=metadata)
            return response
        except grpc.RpcError as e:
            print(f"Update profile failed: {e.code()} - {e.details()}")
            raise

def main():
    with UserGRPCClient() as client:
        try:
            # Register
            message = client.register("user@example.com", "password123")
            print(f"Registration: {message}")
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.ALREADY_EXISTS:
                print("User already exists, proceeding to login")
            else:
                raise
        
        # Login
        token, expires_in = client.login("user@example.com", "password123")
        print(f"Login successful, token expires in {expires_in} seconds")
        
        # Get profile
        profile = client.get_profile(token)
        print(f"Profile: ID={profile.user_id}, Nickname={profile.nickname}, Email={profile.email}")
        
        # Update profile
        update_response = client.update_profile(token, "New Nickname", "newemail@example.com")
        print(f"Update successful: {update_response.message}")

if __name__ == "__main__":
    main()
```

### Async Python Client

```python
import grpc
import asyncio
from proto.user.v1 import user_pb2, user_pb2_grpc

class AsyncUserGRPCClient:
    def __init__(self, address: str = "localhost:8080"):
        self.channel = grpc.aio.insecure_channel(address)
        self.client = user_pb2_grpc.UserServiceStub(self.channel)
    
    async def __aenter__(self):
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
    
    async def close(self):
        await self.channel.close()
    
    async def register(self, email: str, password: str) -> str:
        request = user_pb2.RegisterRequest(email=email, password=password)
        response = await self.client.Register(request)
        return response.message
    
    async def login(self, email: str, password: str) -> tuple[str, int]:
        request = user_pb2.LoginRequest(email=email, password=password)
        response = await self.client.Login(request)
        return response.access_token, response.expires_in
    
    async def get_profile(self, token: str) -> user_pb2.GetProfileResponse:
        metadata = [('authorization', f'Bearer {token}')]
        request = user_pb2.GetProfileRequest()
        return await self.client.GetProfile(request, metadata=metadata)

async def async_main():
    async with AsyncUserGRPCClient() as client:
        # Login
        token, _ = await client.login("user@example.com", "password123")
        
        # Get profile
        profile = await client.get_profile(token)
        print(f"Async Profile: {profile.user_id}")

if __name__ == "__main__":
    asyncio.run(async_main())
```

## Node.js Client

### JavaScript/TypeScript Client

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

class UserGRPCClient {
    constructor(address = 'localhost:8080') {
        // Load protobuf
        const packageDefinition = protoLoader.loadSync('proto/user/v1/user.proto', {
            keepCase: true,
            longs: String,
            enums: String,
            defaults: true,
            oneofs: true
        });
        
        const userProto = grpc.loadPackageDefinition(packageDefinition).user.v1;
        this.client = new userProto.UserService(address, grpc.credentials.createInsecure());
    }

    register(email, password) {
        return new Promise((resolve, reject) => {
            this.client.Register({ email, password }, (error, response) => {
                if (error) {
                    reject(error);
                } else {
                    resolve(response.message);
                }
            });
        });
    }

    login(email, password) {
        return new Promise((resolve, reject) => {
            this.client.Login({ email, password }, (error, response) => {
                if (error) {
                    reject(error);
                } else {
                    resolve({
                        accessToken: response.access_token,
                        expiresIn: response.expires_in
                    });
                }
            });
        });
    }

    getProfile(token) {
        return new Promise((resolve, reject) => {
            const metadata = new grpc.Metadata();
            metadata.add('authorization', `Bearer ${token}`);

            this.client.GetProfile({}, metadata, (error, response) => {
                if (error) {
                    reject(error);
                } else {
                    resolve({
                        userId: response.user_id,
                        nickname: response.nickname,
                        email: response.email
                    });
                }
            });
        });
    }

    updateProfile(token, nickname, email) {
        return new Promise((resolve, reject) => {
            const metadata = new grpc.Metadata();
            metadata.add('authorization', `Bearer ${token}`);

            this.client.UpdateProfile(
                { nickname, email },
                metadata,
                (error, response) => {
                    if (error) {
                        reject(error);
                    } else {
                        resolve({
                            message: response.message,
                            profile: {
                                userId: response.profile.user_id,
                                nickname: response.profile.nickname,
                                email: response.profile.email
                            }
                        });
                    }
                }
            );
        });
    }

    close() {
        this.client.close();
    }
}

async function main() {
    const client = new UserGRPCClient();

    try {
        // Register
        try {
            const message = await client.register('user@example.com', 'password123');
            console.log(`Registration: ${message}`);
        } catch (error) {
            if (error.code === grpc.status.ALREADY_EXISTS) {
                console.log('User already exists, proceeding to login');
            } else {
                throw error;
            }
        }

        // Login
        const { accessToken, expiresIn } = await client.login('user@example.com', 'password123');
        console.log(`Login successful, token expires in ${expiresIn} seconds`);

        // Get profile
        const profile = await client.getProfile(accessToken);
        console.log(`Profile: ID=${profile.userId}, Nickname=${profile.nickname}, Email=${profile.email}`);

        // Update profile
        const updateResponse = await client.updateProfile(accessToken, 'New Nickname', 'newemail@example.com');
        console.log(`Update successful: ${updateResponse.message}`);

    } catch (error) {
        console.error('Error:', error.message);
    } finally {
        client.close();
    }
}

main().catch(console.error);
```

## Java Client

### Spring Boot Integration

```java
package com.example.grpcclient;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Metadata;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.MetadataUtils;
import org.springframework.stereotype.Service;

import java.util.concurrent.TimeUnit;

@Service
public class UserGRPCClient {
    
    private final ManagedChannel channel;
    private final UserServiceGrpc.UserServiceBlockingStub blockingStub;
    private final UserServiceGrpc.UserServiceStub asyncStub;

    public UserGRPCClient() {
        this("localhost", 8080);
    }

    public UserGRPCClient(String host, int port) {
        this.channel = ManagedChannelBuilder.forAddress(host, port)
                .usePlaintext()
                .build();
        this.blockingStub = UserServiceGrpc.newBlockingStub(channel);
        this.asyncStub = UserServiceGrpc.newStub(channel);
    }

    public String register(String email, String password) throws StatusRuntimeException {
        UserOuterClass.RegisterRequest request = UserOuterClass.RegisterRequest.newBuilder()
                .setEmail(email)
                .setPassword(password)
                .build();

        UserOuterClass.RegisterResponse response = blockingStub.register(request);
        return response.getMessage();
    }

    public LoginResult login(String email, String password) throws StatusRuntimeException {
        UserOuterClass.LoginRequest request = UserOuterClass.LoginRequest.newBuilder()
                .setEmail(email)
                .setPassword(password)
                .build();

        UserOuterClass.LoginResponse response = blockingStub.login(request);
        return new LoginResult(response.getAccessToken(), response.getExpiresIn());
    }

    public UserProfile getProfile(String token) throws StatusRuntimeException {
        Metadata metadata = new Metadata();
        metadata.put(Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER), 
                     "Bearer " + token);

        UserServiceGrpc.UserServiceBlockingStub stubWithAuth = 
                MetadataUtils.attachHeaders(blockingStub, metadata);

        UserOuterClass.GetProfileRequest request = UserOuterClass.GetProfileRequest.newBuilder()
                .build();

        UserOuterClass.GetProfileResponse response = stubWithAuth.getProfile(request);
        return new UserProfile(response.getUserId(), response.getNickname(), response.getEmail());
    }

    public UpdateProfileResult updateProfile(String token, String nickname, String email) 
            throws StatusRuntimeException {
        Metadata metadata = new Metadata();
        metadata.put(Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER), 
                     "Bearer " + token);

        UserServiceGrpc.UserServiceBlockingStub stubWithAuth = 
                MetadataUtils.attachHeaders(blockingStub, metadata);

        UserOuterClass.UpdateProfileRequest request = UserOuterClass.UpdateProfileRequest.newBuilder()
                .setNickname(nickname)
                .setEmail(email)
                .build();

        UserOuterClass.UpdateProfileResponse response = stubWithAuth.updateProfile(request);
        UserProfile profile = new UserProfile(
                response.getProfile().getUserId(),
                response.getProfile().getNickname(),
                response.getProfile().getEmail()
        );
        return new UpdateProfileResult(response.getMessage(), profile);
    }

    public void shutdown() throws InterruptedException {
        channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
    }

    // Data classes
    public static class LoginResult {
        private final String accessToken;
        private final long expiresIn;

        public LoginResult(String accessToken, long expiresIn) {
            this.accessToken = accessToken;
            this.expiresIn = expiresIn;
        }

        // Getters...
    }

    public static class UserProfile {
        private final String userId;
        private final String nickname;
        private final String email;

        public UserProfile(String userId, String nickname, String email) {
            this.userId = userId;
            this.nickname = nickname;
            this.email = email;
        }

        // Getters...
    }

    public static class UpdateProfileResult {
        private final String message;
        private final UserProfile profile;

        public UpdateProfileResult(String message, UserProfile profile) {
            this.message = message;
            this.profile = profile;
        }

        // Getters...
    }
}
```

## Command Line Tools

### Using grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:8080 list

# List methods for UserService
grpcurl -plaintext localhost:8080 list user.v1.UserService

# Describe the UserService
grpcurl -plaintext localhost:8080 describe user.v1.UserService

# Register a new user
grpcurl -plaintext \
  -d '{"email":"test@example.com","password":"password123"}' \
  localhost:8080 user.v1.UserService/Register

# Login
grpcurl -plaintext \
  -d '{"email":"test@example.com","password":"password123"}' \
  localhost:8080 user.v1.UserService/Login

# Get profile (with authentication)
grpcurl -plaintext \
  -H 'authorization: Bearer YOUR_JWT_TOKEN' \
  -d '{}' \
  localhost:8080 user.v1.UserService/GetProfile

# Update profile
grpcurl -plaintext \
  -H 'authorization: Bearer YOUR_JWT_TOKEN' \
  -d '{"nickname":"New Name","email":"new@example.com"}' \
  localhost:8080 user.v1.UserService/UpdateProfile
```

### Using grpcui

```bash
# Install grpcui
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

# Start web interface
grpcui -plaintext localhost:8080

# Open browser to http://localhost:8080 for interactive testing
```

## Authentication Examples

### Token Management in Go

```go
type AuthenticatedClient struct {
    client userv1.UserServiceClient
    token  string
    mutex  sync.RWMutex
}

func (c *AuthenticatedClient) getAuthContext(ctx context.Context) context.Context {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    md := metadata.Pairs("authorization", "Bearer "+c.token)
    return metadata.NewOutgoingContext(ctx, md)
}

func (c *AuthenticatedClient) setToken(token string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.token = token
}

func (c *AuthenticatedClient) GetProfile(ctx context.Context) (*userv1.GetProfileResponse, error) {
    authCtx := c.getAuthContext(ctx)
    return c.client.GetProfile(authCtx, &userv1.GetProfileRequest{})
}
```

### Token Refresh Logic

```go
func (c *AuthenticatedClient) loginWithRetry(ctx context.Context, email, password string) error {
    for attempts := 0; attempts < 3; attempts++ {
        resp, err := c.client.Login(ctx, &userv1.LoginRequest{
            Email:    email,
            Password: password,
        })
        if err != nil {
            if attempts == 2 {
                return err
            }
            time.Sleep(time.Second * time.Duration(attempts+1))
            continue
        }
        
        c.setToken(resp.AccessToken)
        return nil
    }
    return nil
}
```

## Error Handling

### Comprehensive Error Handling

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func handleGRPCError(err error) {
    if err == nil {
        return
    }

    st, ok := status.FromError(err)
    if !ok {
        log.Printf("Non-gRPC error: %v", err)
        return
    }

    switch st.Code() {
    case codes.OK:
        // Success
    case codes.Canceled:
        log.Println("Request was cancelled")
    case codes.Unknown:
        log.Println("Unknown error occurred")
    case codes.InvalidArgument:
        log.Printf("Invalid argument: %s", st.Message())
    case codes.DeadlineExceeded:
        log.Println("Request timeout")
    case codes.NotFound:
        log.Printf("Resource not found: %s", st.Message())
    case codes.AlreadyExists:
        log.Printf("Resource already exists: %s", st.Message())
    case codes.PermissionDenied:
        log.Printf("Permission denied: %s", st.Message())
    case codes.Unauthenticated:
        log.Printf("Authentication failed: %s", st.Message())
    case codes.ResourceExhausted:
        log.Println("Resource exhausted (rate limited)")
    case codes.FailedPrecondition:
        log.Printf("Failed precondition: %s", st.Message())
    case codes.Aborted:
        log.Println("Request was aborted")
    case codes.OutOfRange:
        log.Printf("Out of range: %s", st.Message())
    case codes.Unimplemented:
        log.Printf("Method not implemented: %s", st.Message())
    case codes.Internal:
        log.Printf("Internal server error: %s", st.Message())
    case codes.Unavailable:
        log.Println("Service unavailable")
    case codes.DataLoss:
        log.Println("Data loss detected")
    default:
        log.Printf("Unknown gRPC error code: %v", st.Code())
    }
}
```

## Testing Examples

### Unit Testing gRPC Clients

```go
func TestUserClient_Login(t *testing.T) {
    // Start test server
    lis := bufconn.Listen(1024 * 1024)
    s := grpc.NewServer()
    
    // Register mock service
    mockService := &mockUserService{}
    userv1.RegisterUserServiceServer(s, mockService)
    
    go func() {
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Server exited with error: %v", err)
        }
    }()
    defer s.GracefulStop()

    // Create client connection
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

    // Test login
    resp, err := client.Login(context.Background(), &userv1.LoginRequest{
        Email:    "test@example.com",
        Password: "password",
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.AccessToken)
    assert.Greater(t, resp.ExpiresIn, int64(0))
}
```

### Load Testing

```go
func TestUserClient_LoadTest(t *testing.T) {
    client, err := NewUserClient("localhost:8080")
    require.NoError(t, err)
    defer client.Close()

    // Login to get token
    token, err := client.Login(context.Background(), "test@example.com", "password123")
    require.NoError(t, err)

    concurrency := 10
    requestsPerWorker := 100
    
    var wg sync.WaitGroup
    errors := make(chan error, concurrency*requestsPerWorker)

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < requestsPerWorker; j++ {
                ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
                _, err := client.GetProfile(ctx, token)
                cancel()
                if err != nil {
                    errors <- err
                }
            }
        }()
    }

    wg.Wait()
    close(errors)

    errorCount := 0
    for err := range errors {
        t.Logf("Request error: %v", err)
        errorCount++
    }

    successRate := float64(concurrency*requestsPerWorker-errorCount) / float64(concurrency*requestsPerWorker)
    t.Logf("Success rate: %.2f%% (%d/%d)", successRate*100, concurrency*requestsPerWorker-errorCount, concurrency*requestsPerWorker)
    
    assert.Greater(t, successRate, 0.95) // Expect 95%+ success rate
}
```

---

For more examples and advanced usage patterns, see the main [gRPC README](README.md).