package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vucongthanh92/go-base-utils/logger"

	interceptors "github.com/vucongthanh92/go-base-utils/grpc/interceptors"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	cfg                GrpcServerConfig
	log                logger.Logger
	GrpcServerInstance *grpc.Server
}

type GrpcServer interface {
	Run()
	Stop()
}

type GrpcServerConfig struct {
	Port              string
	Development       bool
	MaxConnectionIdle int
	Timeout           int
	MaxConnectionAge  int
	Time              int
}
type GrpcServerOption func(*Server)

func WithLogger(log logger.Logger) GrpcServerOption {
	return func(s *Server) {
		s.log = log
	}
}

func WithPort(port string) GrpcServerOption {
	return func(s *Server) {
		s.cfg.Port = port
	}
}

func WithDevelopment(development bool) GrpcServerOption {
	return func(s *Server) {
		s.cfg.Development = development
	}
}

func WithMaxConnectionIdle(maxConnectionIdle int) GrpcServerOption {
	return func(s *Server) {
		s.cfg.MaxConnectionIdle = maxConnectionIdle
	}
}

func WithTimeout(timeout int) GrpcServerOption {
	return func(s *Server) {
		s.cfg.Timeout = timeout
	}
}

func NewServer(cfg GrpcServerConfig, options ...GrpcServerOption) (GrpcServer, *grpc.Server) {
	customFunc := func(p interface{}) (err error) {
		logger.Fatal("grpc panic triggered", zap.Any("panic", p))
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Duration(cfg.MaxConnectionIdle) * time.Minute,
			Timeout:           time.Duration(cfg.Timeout) * time.Second,
			MaxConnectionAge:  time.Duration(cfg.MaxConnectionAge) * time.Minute,
			Time:              time.Duration(cfg.Time) * time.Minute,
		}),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_prometheus.UnaryServerInterceptor,
				grpc_recovery.UnaryServerInterceptor(),
				interceptors.Localizer(),
				interceptors.Logger(logger.GetDefaultLogger()),
				otelgrpc.UnaryServerInterceptor(),
				grpc_recovery.UnaryServerInterceptor(opts...),
			),
		),
		grpc.ChainStreamInterceptor(
			otelgrpc.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(opts...),
		),
	)

	grpc_prometheus.Register(grpcServer)

	if cfg.Development {
		reflection.Register(grpcServer)
	}
	for _, option := range options {
		option(&Server{cfg: cfg, GrpcServerInstance: grpcServer})
	}

	return &Server{cfg: cfg, GrpcServerInstance: grpcServer}, grpcServer
}

func (s *Server) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	lis, err := net.Listen("tcp", s.cfg.Port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
		panic(err)
	}

	go func() {
		logger.Info("GRPC server is listening at: ", zap.String("PORT", lis.Addr().String()))
		if err := s.GrpcServerInstance.Serve(lis); err != nil {
			logger.Fatal("failed to listen", zap.Error(err))
		}
	}()

	<-ctx.Done()

	_, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		fmt.Println("Close another connection")
		cancel()
	}()

	s.GrpcServerInstance.GracefulStop()
}

func (s *Server) Stop() {
	logger.Info("Stop GRPC server")
	s.GrpcServerInstance.GracefulStop()
}
