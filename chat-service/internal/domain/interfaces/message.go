package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type MessageServiceI interface {
}

type MessageQueryRepoI interface {
	GetMessageByClientMessageID(ctx context.Context, conversationID uint64, clientMessageID string) (*entities.Message, *errHandler.ErrorBuilder)
	ListMessages(ctx context.Context, conversationID uint64, req models.ListMessagesRequest) ([]entities.Message, *errHandler.ErrorBuilder)
	GetMessageByID(ctx context.Context, id uint64) (*entities.Message, *errHandler.ErrorBuilder)
}

type MessageCmdRepoI interface {
	CreateMessage(ctx context.Context, entity *entities.Message) *errHandler.ErrorBuilder
}
