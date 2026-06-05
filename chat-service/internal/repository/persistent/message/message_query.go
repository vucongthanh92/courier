package message

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/database"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"gorm.io/gorm"
)

type repoQueryMessage struct {
	readDB *gorm.DB
}

func InitMessageQueryRepo(readDb *database.GormReadDb, writeDb *database.GormWriteDb) interfaces.MessageQueryRepoI {
	return &repoQueryMessage{
		readDB: *readDb,
	}
}

func (r *repoQueryMessage) GetMessageByClientMessageID(ctx context.Context, conversationID uint64, clientMessageID string) (*entities.Message, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.Message
	err := run.Model(&entities.Message{}).
		Where("conversation_id = ? AND client_message_id = ?", conversationID, clientMessageID).
		Take(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}

func (r *repoQueryMessage) ListMessages(ctx context.Context, conversationID uint64, req models.ListMessagesRequest) ([]entities.Message, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	query := run.Model(&entities.Message{}).
		Where("conversation_id = ? AND deleted_at IS NULL", conversationID).
		Order("created_at DESC, id DESC")
	if req.BeforeMessageID != nil {
		query = query.Where("id < ?", *req.BeforeMessageID)
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var res []entities.Message
	if err := query.Limit(limit).Find(&res).Error; err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return res, nil
}

func (r *repoQueryMessage) GetMessageByID(ctx context.Context, id uint64) (*entities.Message, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.Message
	err := run.Model(&entities.Message{}).
		Where("id = ?", id).
		Take(&res).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}
