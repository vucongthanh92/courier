package api

import (
	chatcron "github.com/vucongthanh92/courier/chat-service/internal/api/cron"
	chatgrpc "github.com/vucongthanh92/courier/chat-service/internal/api/grpc"
	chathttp "github.com/vucongthanh92/courier/chat-service/internal/api/http"
)

type ApiContainer struct {
	HttpServer *chathttp.Server
	GrpcServer *chatgrpc.Server
	CronServer *chatcron.Server
}

func NewApiContainer(
	http *chathttp.Server,
	grpc *chatgrpc.Server,
	cron *chatcron.Server,
) *ApiContainer {
	return &ApiContainer{
		HttpServer: http,
		GrpcServer: grpc,
		CronServer: cron,
	}
}
