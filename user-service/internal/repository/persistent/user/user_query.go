package user

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type userQueryRepository struct {
	readDb *gorm.DB
}

func InitUserQueryRepository(readDb *database.GormReadDb) interfaces.UserQueryRepoI {
	return &userQueryRepository{
		readDb: *readDb,
	}
}

func (repo *userQueryRepository) GetUserByID(ctx context.Context, id uint64) (res entities.User, errRes *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetUserByID")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	// Query user by ID
	err := run.Model(&entities.User{}).Select("*").
		Where("id = ?", id).Where("deleted_at is null").
		Take(&res).Error

	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return res, resErr
	}

	return res, errRes
}
