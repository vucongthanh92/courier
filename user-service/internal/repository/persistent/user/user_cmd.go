package user

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type userCmdRepository struct {
	writeDB *gorm.DB
}

func InitUserCmdRepository(writeDB *database.GormWriteDb) interfaces.UserCommandRepoI {
	return &userCmdRepository{
		writeDB: *writeDB,
	}
}

func (repo *userCmdRepository) InsertUser(ctx context.Context, entity *entities.User) *errHandler.ErrorBuilder {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertUser")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDB)

	// Insert user record
	err := run.Model(entities.User{}).Create(entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return resErr
	}

	return nil
}
