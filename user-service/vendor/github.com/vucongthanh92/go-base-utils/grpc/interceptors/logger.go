package interceptors

import (
	"context"
	"time"

	"github.com/vucongthanh92/go-base-utils/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Logger(logger logger.Logger) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		md, _ := metadata.FromIncomingContext(ctx)
		reply, err := handler(ctx, req)

		logger.Info("GRPC",
			zap.String("Method", info.FullMethod),
			zap.Duration("Time", time.Since(start)),
			zap.Any("Metadata", md),
			zap.Error(err),
		)

		return reply, err
	}
}

func ClientLogger(logger logger.Logger) func(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		md, _ := metadata.FromIncomingContext(ctx)

		logger.Info("GRPC",
			zap.String("Method", method),
			zap.Any("Request", req),
			zap.Any("Reply", reply),
			zap.Duration("Time", time.Since(start)),
			zap.Any("Metadata", md),
			zap.Error(err),
		)

		return err
	}
}
