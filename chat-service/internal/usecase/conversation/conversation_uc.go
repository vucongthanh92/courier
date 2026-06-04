package conversation

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
)

type ConversationUseCaseImpl struct {
	conversationReadRepo interfaces.ConversationQueryRepoI
}

func InitConversationUsecase(
	conversationReadRepo interfaces.ConversationQueryRepoI,
) interfaces.ConversationServiceI {
	return &ConversationUseCaseImpl{
		conversationReadRepo: conversationReadRepo,
	}
}

func (s *ConversationUseCaseImpl) CreateConversation(ctx context.Context) {
}
