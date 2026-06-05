package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
)

type MemberServiceI interface {
}

type MemberCmdRepoI interface {
	CreateMembers(ctx context.Context, entities []entities.ConversationMember) *errHandler.ErrorBuilder
	UpdateReadState(ctx context.Context, conversationID, userID, lastReadMessageID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder)
}

type MemberQueryRepoI interface {
	ListConversationMembers(ctx context.Context, conversationID uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder)
	GetConversationMember(ctx context.Context, conversationID, userID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder)
}
