package api

import (
	grpcapi "github.com/vucongthanh92/courier/payment-gateway/internal/api/grpc"
	httpapi "github.com/vucongthanh92/courier/payment-gateway/internal/api/http"
)

// ApiContainer is the startup boundary assembled exclusively by Wire.
type ApiContainer struct {
	HttpServer *httpapi.Server
	GrpcServer *grpcapi.Server
}

func NewApiContainer(httpServer *httpapi.Server, grpcServer *grpcapi.Server) *ApiContainer {
	return &ApiContainer{HttpServer: httpServer, GrpcServer: grpcServer}
}
