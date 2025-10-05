package category

import (
	"context"
	"errors"
	"time"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type categoryCommandRepository struct {
	writeDB *gorm.DB
}

func NewCategoryCommandRepository(writeDB *database.GormWriteDb) interfaces.CategoryCommandRepoI {
	return &categoryCommandRepository{
		writeDB: *writeDB,
	}
}

func (repo *categoryCommandRepository) InsertCategory(ctx context.Context, entity entities.Category) (
	entities.Category, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "InsertCategory")
	defer span.End()

	err := repo.writeDB.WithContext(ctx).Model(entities.Category{}).
		Create(&entity).Error

	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return entity, resErr
	}

	return entity, nil
}

func (repo *categoryCommandRepository) UpdateCategory(ctx context.Context, entity entities.Category) (
	entities.Category, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "UpdateCategory")
	defer span.End()

	res := repo.writeDB.WithContext(ctx).Model(entities.Category{}).
		Where("id = ?", entity.ID).Where("updated_at = ?", entity.UpdatedAt).
		Update("name", entity.Name)

	if res.Error != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(res.Error)
		return entity, resErr
	}

	if res.RowsAffected == 0 {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(errors.New(constants.STATUS_CONFLICT))
		return entity, resErr
	}

	return entity, nil
}

func (repo *categoryCommandRepository) SoftDeleteCategoryByID(ctx context.Context, id uint64, updatedAt time.Time) *errHandler.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "UpdateCategory")
	defer span.End()

	res := repo.writeDB.Debug().WithContext(ctx).Model(&entities.Category{}).
		Where("id = ?", id).Where("updated_at = ?", updatedAt).
		Update("deleted_at", time.Now())

	if res.Error != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(res.Error)
		return resErr
	}

	if res.RowsAffected == 0 {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(errors.New(constants.STATUS_CONFLICT))
		return resErr
	}

	return nil
}
