package server

import (
	userv1 "boilerplate/api/grpc/proto/user/v1"
	grpcHandler "boilerplate/internal/grpc"
	"boilerplate/internal/grpc/interceptor"
	"boilerplate/internal/service"
	"boilerplate/pkg/jwt"
	"boilerplate/pkg/log"
	grpcServer "boilerplate/pkg/server/grpc"

	"github.com/spf13/viper"
)

// NewGRPCServer creates and configures a new gRPC server
func NewGRPCServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	userService service.UserService,
) *grpcServer.Server {
	// Create interceptors
	authInterceptor := interceptor.NewAuthInterceptor(jwt, logger)
	loggingInterceptor := interceptor.NewLoggingInterceptor(logger)
	recoveryInterceptor := interceptor.NewRecoveryInterceptor(logger)

	// Create gRPC server with interceptors
	s := grpcServer.NewServer(
		logger,
		grpcServer.WithServerHost(conf.GetString("grpc.host")),
		grpcServer.WithServerPort(conf.GetInt("grpc.port")),
		grpcServer.WithUnaryInterceptors(
			recoveryInterceptor.UnaryInterceptor(),
			loggingInterceptor.UnaryInterceptor(),
			authInterceptor.UnaryInterceptor(),
		),
	)

	// Create and register service handlers
	userGRPCServer := grpcHandler.NewUserGRPCServer(userService, logger)
	userv1.RegisterUserServiceServer(s.Server, userGRPCServer)

	return s
}
