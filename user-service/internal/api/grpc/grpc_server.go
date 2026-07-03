package grpc

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
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
	jwkpb.RegisterJWKServiceServer(srv, NewJWKServer(s.JwkRepo, s.JwkCache))
	grpcServer.Run()
}

type JWKServer struct {
	jwkpb.UnimplementedJWKServiceServer
	jwkRepo  interfaces.JWKQueryRepoI
	jwkCache cacheRepo.JWKCacheRepo
	cacheTTL time.Duration
}

func NewJWKServer(jwkRepo interfaces.JWKQueryRepoI, jwkCache cacheRepo.JWKCacheRepo) *JWKServer {
	return &JWKServer{
		jwkRepo:  jwkRepo,
		jwkCache: jwkCache,
		cacheTTL: 15 * time.Minute,
	}
}

func (s *JWKServer) GetPublicKey(ctx context.Context, req *jwkpb.GetPublicKeyRequest) (*jwkpb.GetPublicKeyResponse, error) {
	kid := req.Kid
	if kid == "" {
		active, commonErr := s.jwkRepo.GetActiveKey(ctx)
		if commonErr != nil {
			return nil, commonErr.LogError
		}
		kid = active.Kid
	}

	var entry *models.JWKCacheEntry
	if s.jwkCache != nil {
		entry, _ = s.jwkCache.GetByKid(ctx, kid)
	}
	if entry != nil {
		return &jwkpb.GetPublicKeyResponse{
			Kid:       entry.Kid,
			PublicPem: entry.PublicPEM,
			Alg:       entry.Alg,
		}, nil
	}

	key, commonErr := s.jwkRepo.GetKeyByKid(ctx, kid)
	if commonErr != nil {
		return nil, commonErr.LogError
	}
	if s.jwkCache != nil {
		_ = s.jwkCache.SetByKid(ctx, models.JWKCacheEntry{Kid: key.Kid, PublicPEM: key.PublicPEM, Alg: key.Alg}, s.cacheTTL)
	}
	return &jwkpb.GetPublicKeyResponse{
		Kid:       key.Kid,
		PublicPem: key.PublicPEM,
		Alg:       key.Alg,
	}, nil
}
