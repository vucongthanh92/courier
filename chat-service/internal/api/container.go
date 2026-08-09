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
	ChatEventConsumer         *worker.ChatEventConsumer
	AssistantResponseConsumer *worker.AssistantResponseConsumer
}

func NewApiContainer(
	http *chathttp.Server,
	grpc *chatgrpc.Server,
	cron *chatcron.Server,
	userEventConsumer *worker.UserEventConsumer,
	chatEventConsumer *worker.ChatEventConsumer,
	assistantResponseConsumer *worker.AssistantResponseConsumer,
) *ApiContainer {
	return &ApiContainer{
		HttpServer:                http,
		GrpcServer:                grpc,
		CronServer:                cron,
		UserEventConsumer:         userEventConsumer,
		ChatEventConsumer:         chatEventConsumer,
		AssistantResponseConsumer: assistantResponseConsumer,
	}
}
