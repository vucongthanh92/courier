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

type repoQueryMember struct {
	readDB *gorm.DB
}

func InitMemberQueryRepo(readDb *database.GormReadDb, writeDb *database.GormWriteDb) interfaces.MemberQueryRepoI {
	return &repoQueryMember{
		readDB: *readDb,
	}
}

func (r *repoQueryMember) ListConversationMembers(ctx context.Context, conversationID uint64) ([]entities.ConversationMember, *errHandler.ErrorBuilder) {
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

func (r *repoQueryMember) GetConversationMember(ctx context.Context, conversationID, userID uint64) (*entities.ConversationMember, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var res entities.ConversationMember
	err := run.Model(&entities.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Take(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &res, nil
}
