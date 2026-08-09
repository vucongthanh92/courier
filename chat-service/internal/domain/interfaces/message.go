package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type MessageServiceI interface {
	CreateMessage(ctx context.Context, req *models.SendMessageRequest) (*models.MessageResponse, bool, *errHandler.ErrorBuilder)
	CreateSystemMessage(ctx context.Context, req *models.CreateSystemMessageRequest) (*models.MessageResponse, bool, *errHandler.ErrorBuilder)
	ListMessages(ctx context.Context, req *models.ListMessagesRequest) (*models.ListMessagesResponse, *errHandler.ErrorBuilder)
}

type AssistantEventPublisherI interface {
	PublishAssistantRequested(ctx context.Context, payload models.AssistantRequestedPayload) error
}

type ChatEventPublisherI interface {
	PublishConversationCreated(ctx context.Context, payload models.ConversationCreatedPayload) error
}

type MessageQueryRepoI interface {
	GetMessageByClientMessageID(ctx context.Context, conversationID uint64, clientMessageID string) (*entities.Message, *errHandler.ErrorBuilder)
	ListMessages(ctx context.Context, conversationID uint64, req models.ListMessagesRequest) ([]entities.Message, *errHandler.ErrorBuilder)
	GetMessageByID(ctx context.Context, id uint64) (*entities.Message, *errHandler.ErrorBuilder)
}

type MessageCmdRepoI interface {
	CreateMessage(ctx context.Context, entity *entities.Message) *errHandler.ErrorBuilder
}
