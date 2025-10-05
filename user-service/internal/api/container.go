package api

import (
	"github.com/vucongthanh92/courier/user-service/internal/api/cron"
	"github.com/vucongthanh92/courier/user-service/internal/api/grpc"
	"github.com/vucongthanh92/courier/user-service/internal/api/http"
)

type ApiContainer struct {
	HttpServer *http.Server
	GrpcServer *grpc.Server
	CronServer *cron.Server
}

func NewApiContainer(
	http *http.Server,
	grpc *grpc.Server,
	cron *cron.Server,
) *ApiContainer {
	return &ApiContainer{
		HttpServer: http,
		GrpcServer: grpc,
		CronServer: cron,
	}
}
