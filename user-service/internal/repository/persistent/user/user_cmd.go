package category

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
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

func (repo *userCmdRepository) InsertUser(ctx context.Context, entity entities.User) (
	entities.User, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "InsertUser")
	defer span.End()

	err := repo.writeDB.WithContext(ctx).Model(entities.User{}).
		Create(&entity).Error

	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return entity, resErr
	}

	return entity, nil
}
