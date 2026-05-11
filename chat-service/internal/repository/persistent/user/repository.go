package user

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/database"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"gorm.io/gorm"
)

type repository struct {
	readDB *gorm.DB
}

func InitUserRepository(readDb *database.GormReadDb) interfaces.UserRepositoryI {
	return &repository{readDB: *readDb}
}

func (r *repository) CountExistingUsers(ctx context.Context, userIDs []uint64) (int64, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	var count int64
	if len(userIDs) == 0 {
		return 0, nil
	}
	err := run.Table(`"user-service".users`).
		Where("id IN ?", userIDs).
		Count(&count).Error
	if err != nil {
		return 0, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return count, nil
}
