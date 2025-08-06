package grpc

import (
	"context"
	"fmt"
	"boilerplate/pkg/log"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	*grpc.Server
	host               string
	port               int
	logger             *log.Logger
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

type Option func(s *Server)

func NewServer(logger *log.Logger, opts ...Option) *Server {
	s := &Server{
		logger: logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	
	// Create server options with interceptors
	var serverOpts []grpc.ServerOption
	if len(s.unaryInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(s.unaryInterceptors...))
	}
	if len(s.streamInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.ChainStreamInterceptor(s.streamInterceptors...))
	}
	
	s.Server = grpc.NewServer(serverOpts...)
	return s
}
func WithServerHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}
func WithServerPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}

func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(s *Server) {
		s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
	}
}

func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(s *Server) {
		s.streamInterceptors = append(s.streamInterceptors, interceptors...)
	}
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		s.logger.Sugar().Fatalf("Failed to listen: %v", err)
	}
	if err = s.Server.Serve(lis); err != nil {
		s.logger.Sugar().Fatalf("Failed to serve: %v", err)
	}
	return nil

}
func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	s.Server.GracefulStop()

	s.logger.Info("Server exiting")

	return nil
}
