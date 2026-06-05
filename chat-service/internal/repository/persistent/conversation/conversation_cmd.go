package conversation

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/database"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"gorm.io/gorm"
)

type repoCommandConversation struct {
	writeDB *gorm.DB
}

func InitConversationCommandRepo(readDb *database.GormReadDb, writeDb *database.GormWriteDb) interfaces.ConversationCommandRepoI {
	return &repoCommandConversation{
		writeDB: *writeDb,
	}
}

func (r *repoCommandConversation) CreateConversation(ctx context.Context, entity *entities.Conversation) *errHandler.ErrorBuilder {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	if err := run.Model(&entities.Conversation{}).Create(entity).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}
