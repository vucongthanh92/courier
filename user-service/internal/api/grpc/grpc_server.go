package grpc

import (
	userstatuspb "github.com/vucongthanh92/courier/shared/grpc/user-service/user_status/gen"
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	cacheRepo "github.com/vucongthanh92/courier/user-service/internal/repository/external/redis"
	jwkpb "github.com/vucongthanh92/courier/user-service/pkg/grpc/gen"
	grpc "github.com/vucongthanh92/go-base-utils/grpc/server"
)

type Server struct {
	Cfg      *config.AppConfig
	JwkRepo  interfaces.JWKQueryRepoI
	UserRepo interfaces.UserQueryRepoI
	JwkCache cacheRepo.JWKCacheRepo
}

func NewServer(cfg *config.AppConfig, jwkRepo interfaces.JWKQueryRepoI, userRepo interfaces.UserQueryRepoI, jwkCache cacheRepo.JWKCacheRepo) *Server {
	return &Server{
		Cfg:      cfg,
		JwkRepo:  jwkRepo,
		UserRepo: userRepo,
		JwkCache: jwkCache,
	}
}

func (s *Server) Run() {
	grpcServer, srv := grpc.NewServer(
		grpc.GrpcServerConfig(*s.Cfg.GRPC),
	)

	grpcUsecase := NewGrpcUsecase(s.JwkRepo, s.JwkCache, s.UserRepo)

	// Register all gRPC services exposed by user-service.
	jwkpb.RegisterJWKServiceServer(srv, grpcUsecase)
	userstatuspb.RegisterUserStatusServiceServer(srv, grpcUsecase)

	// Start the gRPC server
	grpcServer.Run()
}
