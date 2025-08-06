package interceptor

import (
	"context"
	"runtime/debug"

	"boilerplate/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor provides gRPC panic recovery interceptor
type RecoveryInterceptor struct {
	logger *log.Logger
}

// NewRecoveryInterceptor creates a new recovery interceptor
func NewRecoveryInterceptor(logger *log.Logger) *RecoveryInterceptor {
	return &RecoveryInterceptor{
		logger: logger,
	}
}

// UnaryInterceptor returns a server interceptor function to recover from panics in unary RPCs
func (r *RecoveryInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if panicVal := recover(); panicVal != nil {
				// Log the panic with stack trace
				r.logger.Sugar().Errorf(
					"gRPC panic recovered in %s: %v\nStack trace:\n%s",
					info.FullMethod,
					panicVal,
					string(debug.Stack()),
				)

				// Return internal server error
				err = status.Error(codes.Internal, "Internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
