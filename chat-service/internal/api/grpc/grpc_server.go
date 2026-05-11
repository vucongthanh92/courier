package grpc

import (
	"github.com/vucongthanh92/courier/chat-service/config"
	basegrpc "github.com/vucongthanh92/go-base-utils/grpc/server"
)

type Server struct {
	cfg *config.AppConfig
}

func NewServer(cfg *config.AppConfig) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Run() {
	grpcServer, _ := basegrpc.NewServer(
		basegrpc.GrpcServerConfig(*s.cfg.GRPC),
	)
	grpcServer.Run()
}
