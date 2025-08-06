package grpc

import (
	"context"
	"time"
	
	userv1 "boilerplate/api/grpc/proto/user/v1"
	v1 "boilerplate/api/v1"
	"boilerplate/internal/service"
	"boilerplate/pkg/log"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserGRPCServer implements the gRPC UserService
type UserGRPCServer struct {
	userv1.UnimplementedUserServiceServer
	userService service.UserService
	logger      *log.Logger
}

// NewUserGRPCServer creates a new gRPC user server
func NewUserGRPCServer(
	userService service.UserService,
	logger *log.Logger,
) *UserGRPCServer {
	return &UserGRPCServer{
		userService: userService,
		logger:      logger,
	}
}

// Register creates a new user account
func (s *UserGRPCServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	// Convert gRPC request to service request
	serviceReq := &v1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	
	// Call service layer
	err := s.userService.Register(ctx, serviceReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	
	return &userv1.RegisterResponse{
		Message: "User registered successfully",
	}, nil
}

// Login authenticates a user and returns an access token
func (s *UserGRPCServer) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	// Convert gRPC request to service request
	serviceReq := &v1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	
	// Call service layer
	token, err := s.userService.Login(ctx, serviceReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	
	return &userv1.LoginResponse{
		AccessToken: token,
		ExpiresIn:   int64(24 * 90 * time.Hour.Seconds()), // 90 days in seconds
	}, nil
}

// GetProfile retrieves the user's profile information
func (s *UserGRPCServer) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "User ID not found in context")
	}
	
	// Call service layer
	profile, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, s.mapError(err)
	}
	
	return &userv1.GetProfileResponse{
		UserId:   profile.UserId,
		Nickname: profile.Nickname,
		Email:    "", // Email will be added when we have it in the service response
	}, nil
}

// UpdateProfile updates the user's profile information
func (s *UserGRPCServer) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "User ID not found in context")
	}
	
	// Convert gRPC request to service request
	serviceReq := &v1.UpdateProfileRequest{
		Nickname: req.Nickname,
		Email:    req.Email,
	}
	
	// Call service layer
	err := s.userService.UpdateProfile(ctx, userID, serviceReq)
	if err != nil {
		return nil, s.mapError(err)
	}
	
	// Get updated profile
	profile, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, s.mapError(err)
	}
	
	return &userv1.UpdateProfileResponse{
		Message: "Profile updated successfully",
		Profile: &userv1.GetProfileResponse{
			UserId:   profile.UserId,
			Nickname: profile.Nickname,
			Email:    req.Email,
		},
	}, nil
}

// mapError converts service errors to gRPC status errors
func (s *UserGRPCServer) mapError(err error) error {
	switch err {
	case v1.ErrEmailAlreadyUse:
		return status.Error(codes.AlreadyExists, "Email already in use")
	case v1.ErrUnauthorized:
		return status.Error(codes.Unauthenticated, "Invalid credentials")
	case v1.ErrInternalServerError:
		return status.Error(codes.Internal, "Internal server error")
	default:
		s.logger.Sugar().Errorf("Unmapped error: %v", err)
		return status.Error(codes.Internal, "Internal server error")
	}
}