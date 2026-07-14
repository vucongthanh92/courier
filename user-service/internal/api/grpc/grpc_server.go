package grpc

import (
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	cacheRepo "github.com/vucongthanh92/courier/user-service/internal/repository/external/redis"
	jwkpb "github.com/vucongthanh92/courier/user-service/pkg/grpc/gen"
	grpc "github.com/vucongthanh92/go-base-utils/grpc/server"
)

type Server struct {
	Cfg      *config.AppConfig
	JwkRepo  interfaces.JWKQueryRepoI
	JwkCache cacheRepo.JWKCacheRepo
}

func NewServer(cfg *config.AppConfig, jwkRepo interfaces.JWKQueryRepoI, jwkCache cacheRepo.JWKCacheRepo) *Server {
	return &Server{
		Cfg:      cfg,
		JwkRepo:  jwkRepo,
		JwkCache: jwkCache,
	}
}

func (s *Server) Run() {
	grpcServer, srv := grpc.NewServer(
		grpc.GrpcServerConfig(*s.Cfg.GRPC),
	)

	// Register gRPC services, including the JWKService
	jwkpb.RegisterJWKServiceServer(srv, NewJWKServer(s.JwkRepo, s.JwkCache))

	// Start the gRPC server
	grpcServer.Run()
}
