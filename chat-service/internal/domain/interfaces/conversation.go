package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type ConversationServiceI interface {
	CreateConversation(ctx context.Context, req *models.CreateConversationRequest) (*models.CreateConversationResponse, *errHandler.ErrorBuilder)
}

type ConversationQueryRepoI interface {
	GetDirectConversationByKey(ctx context.Context, directKey string) (*entities.Conversation, *errHandler.ErrorBuilder)
	GetConversationByID(ctx context.Context, id uint64) (*entities.Conversation, *errHandler.ErrorBuilder)
}

type ConversationCommandRepoI interface {
	CreateConversation(ctx context.Context, entity *entities.Conversation) *errHandler.ErrorBuilder
}
