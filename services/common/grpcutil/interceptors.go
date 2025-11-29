package grpcutil

import (
	"context"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC interceptor for logging and tracing
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Start tracing span
		ctx, span := telemetry.StartSpan(ctx, info.FullMethod)
		defer span.End()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		logger.WithContext(ctx).Info("gRPC request",
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration,
			"error", err,
		)

		return resp, err
	}
}

// UnaryClientInterceptor returns a gRPC client interceptor for logging
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()

		// Start tracing span
		ctx, span := telemetry.StartSpan(ctx, method)
		defer span.End()

		// Call the method
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Log the call
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		logger.WithContext(ctx).Info("gRPC client call",
			"method", method,
			"code", code.String(),
			"duration", duration,
			"error", err,
		)

		return err
	}
}
