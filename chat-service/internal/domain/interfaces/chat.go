package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type ChatUseCaseI interface {
	CreateDirectConversation(ctx context.Context, actorUserID uint64, req models.CreateDirectConversationRequest) (*models.ConversationResponse, *errHandler.ErrorBuilder)
	CreateGroupConversation(ctx context.Context, actorUserID uint64, req models.CreateGroupConversationRequest) (*models.ConversationResponse, *errHandler.ErrorBuilder)
	SendTextMessage(ctx context.Context, actorUserID uint64, conversationID uint64, req models.SendMessageRequest) (*models.MessageResponse, *errHandler.ErrorBuilder)
	ListInbox(ctx context.Context, actorUserID uint64) ([]models.InboxConversationResponse, *errHandler.ErrorBuilder)
	ListMessages(ctx context.Context, actorUserID uint64, conversationID uint64, req models.ListMessagesRequest) ([]models.MessageResponse, *errHandler.ErrorBuilder)
	MarkConversationRead(ctx context.Context, actorUserID uint64, conversationID uint64, req models.MarkConversationReadRequest) (*models.ConversationMemberResponse, *errHandler.ErrorBuilder)
}

type ConversationRepositoryI interface {
	GetDirectConversationByKey(ctx context.Context, directKey string) (*entities.Conversation, *errHandler.ErrorBuilder)
	GetConversationByID(ctx context.Context, id uint64) (*entities.Conversation, *errHandler.ErrorBuilder)
	CreateConversation(ctx context.Context, entity *entities.Conversation) *errHandler.ErrorBuilder
	ListInbox(ctx context.Context, userID uint64) ([]models.InboxConversationResponse, *errHandler.ErrorBuilder)
}

type ConversationMemberRepositoryI interface {
	CreateMembers(ctx context.Context, entities []entities.ConversationMember) *errHandler.ErrorBuilder
	ListConversationMembers(ctx context.Context, conversationID uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder)
	GetConversationMember(ctx context.Context, conversationID, userID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder)
	UpdateReadState(ctx context.Context, conversationID, userID, lastReadMessageID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder)
}

type MessageRepositoryI interface {
	CreateMessage(ctx context.Context, entity *entities.Message) *errHandler.ErrorBuilder
	GetMessageByClientMessageID(ctx context.Context, conversationID uint64, clientMessageID string) (*entities.Message, *errHandler.ErrorBuilder)
	ListMessages(ctx context.Context, conversationID uint64, req models.ListMessagesRequest) ([]entities.Message, *errHandler.ErrorBuilder)
	GetMessageByID(ctx context.Context, id uint64) (*entities.Message, *errHandler.ErrorBuilder)
}

type UserRepositoryI interface {
	CountExistingUsers(ctx context.Context, userIDs []uint64) (int64, *errHandler.ErrorBuilder)
}
