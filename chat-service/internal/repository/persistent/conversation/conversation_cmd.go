package conversation

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/database"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *repoCommandConversation) CreateProcessedEvent(ctx context.Context, eventID, eventType string) (bool, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	entity := entities.ProcessedEvent{
		EventID:   eventID,
		EventType: eventType,
	}
	res := run.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity)
	if res.Error != nil {
		return false, errHandler.InitErrorBuilder(ctx).ValidateError(res.Error)
	}
	return res.RowsAffected > 0, nil
}
