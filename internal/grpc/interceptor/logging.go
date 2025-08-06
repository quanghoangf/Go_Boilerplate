package interceptor

import (
	"context"
	"time"
	
	"boilerplate/pkg/log"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor provides gRPC logging interceptor
type LoggingInterceptor struct {
	logger *log.Logger
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(logger *log.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logger,
	}
}

// UnaryInterceptor returns a server interceptor function to log unary RPCs
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

// getErrorMessage returns error message or empty string if no error
func (l *LoggingInterceptor) getErrorMessage(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}