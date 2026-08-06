package api

import (
	chatcron "github.com/vucongthanh92/courier/chat-service/internal/api/cron"
	chatgrpc "github.com/vucongthanh92/courier/chat-service/internal/api/grpc"
	chathttp "github.com/vucongthanh92/courier/chat-service/internal/api/http"
	"github.com/vucongthanh92/courier/chat-service/internal/worker"
)

type ApiContainer struct {
	HttpServer                *chathttp.Server
	GrpcServer                *chatgrpc.Server
	CronServer                *chatcron.Server
	UserEventConsumer         *worker.UserEventConsumer
	AssistantResponseConsumer *worker.AssistantResponseConsumer
}

func NewApiContainer(
	http *chathttp.Server,
	grpc *chatgrpc.Server,
	cron *chatcron.Server,
	userEventConsumer *worker.UserEventConsumer,
	assistantResponseConsumer *worker.AssistantResponseConsumer,
) *ApiContainer {
	return &ApiContainer{
		HttpServer:                http,
		GrpcServer:                grpc,
		CronServer:                cron,
		UserEventConsumer:         userEventConsumer,
		AssistantResponseConsumer: assistantResponseConsumer,
	}
}
