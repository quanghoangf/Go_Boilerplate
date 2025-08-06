package interceptor

import (
	"context"
	"strings"
	
	"boilerplate/pkg/jwt"
	"boilerplate/pkg/log"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor provides gRPC authentication interceptor
type AuthInterceptor struct {
	jwt    *jwt.JWT
	logger *log.Logger
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(jwt *jwt.JWT, logger *log.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		jwt:    jwt,
		logger: logger,
	}
}

// UnaryInterceptor returns a server interceptor function to authenticate and authorize unary RPCs
func (a *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for Register and Login methods
		if a.isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}
		
		// Extract token from metadata
		token, err := a.extractToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		
		// Validate token and extract user ID
		claims, err := a.jwt.ParseToken(token)
		if err != nil {
			a.logger.Sugar().Errorf("Invalid token: %v", err)
			return nil, status.Error(codes.Unauthenticated, "Invalid token")
		}
		
		// Add user ID to context
		ctx = context.WithValue(ctx, "userID", claims.UserId)
		
		return handler(ctx, req)
	}
}

// extractToken extracts JWT token from gRPC metadata
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

// isPublicMethod checks if the method doesn't require authentication
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