package grpc

import (
	"github.com/vucongthanh92/go-base-project/config"
	grpc "github.com/vucongthanh92/go-base-utils/grpc/server"
)

type Server struct {
	Cfg *config.AppConfig
}

func NewServer(cfg *config.AppConfig) *Server {
	return &Server{
		Cfg: cfg,
	}
}

func (s *Server) Run() {
	grpcServer, _ := grpc.NewServer(
		grpc.GrpcServerConfig(*s.Cfg.GRPC),
	)
	grpcServer.Run()
}
