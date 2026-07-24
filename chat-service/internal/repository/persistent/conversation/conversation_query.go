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

type repoQueryConversation struct {
	readDB *gorm.DB
}

func InitConversationQueryRepo(readDb *database.GormReadDb, writeDb *database.GormWriteDb) interfaces.ConversationQueryRepoI {
	return &repoQueryConversation{
		readDB: *readDb,
	}
}

func (r *repoQueryConversation) GetDirectConversationByKey(ctx context.Context, directKey string) (*entities.Conversation, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.Conversation
	err := run.Model(&entities.Conversation{}).
		Where("direct_key = ? AND deleted_at IS NULL", directKey).
		Take(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}

func (r *repoQueryConversation) GetConversationByID(ctx context.Context, id uint64) (*entities.Conversation, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.Conversation
	err := run.Model(&entities.Conversation{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}
