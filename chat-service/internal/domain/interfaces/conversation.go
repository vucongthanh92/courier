package interfaces

import (
	"context"
)

type ConversationServiceI interface {
	CreateConversation(ctx context.Context)
}

type ConversationQueryRepoI interface {
}
