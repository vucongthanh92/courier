package member

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/database"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"gorm.io/gorm"
)

type repository struct {
	readDB  *gorm.DB
	writeDB *gorm.DB
}

func InitConversationMemberRepository(readDb *database.GormReadDb, writeDb *database.GormWriteDb) interfaces.ConversationMemberRepositoryI {
	return &repository{readDB: *readDb, writeDB: *writeDb}
}

func (r *repository) CreateMembers(ctx context.Context, entities []entities.ConversationMember) *errHandler.ErrorBuilder {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	if err := run.Model(&entities.ConversationMember{}).Create(&entities).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *repository) ListConversationMembers(ctx context.Context, conversationID uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res []entities.ConversationMember
	if err := run.Model(&entities.ConversationMember{}).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&res).Error; err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return res, nil
}

func (r *repository) GetConversationMember(ctx context.Context, conversationID, userID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.ConversationMember
	err := run.Model(&entities.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Take(&res).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}

func (r *repository) UpdateReadState(ctx context.Context, conversationID, userID, lastReadMessageID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	res := entities.ConversationMember{}
	err := run.Model(&entities.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Updates(map[string]any{
			"last_read_message_id": lastReadMessageID,
			"last_read_at":         gorm.Expr("NOW()"),
			"updated_at":           gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	err = run.Model(&entities.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Take(&res).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}
