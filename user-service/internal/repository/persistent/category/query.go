package category

import (
	"context"

	"github.com/vucongthanh92/go-base-project/database"
	errHandler "github.com/vucongthanh92/go-base-project/helper/error_handler"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/go-base-project/internal/domain/entities"
	"github.com/vucongthanh92/go-base-project/internal/domain/interfaces"
)

type categoryQueryRepository struct {
	readDb *gorm.DB
}

func NewCategoryQueryRepository(readDb *database.GormReadDb) interfaces.CategoryQueryRepoI {
	return &categoryQueryRepository{
		readDb: *readDb,
	}
}

func (repo *categoryQueryRepository) GetCategoryByID(ctx context.Context, id uint64) (res entities.Category, errRes *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetCategoryByID")
	defer span.End()

	err := repo.readDb.WithContext(ctx).Model(&entities.Category{}).
		Select("id, name, description, status, created_at, updated_at").
		Where("id = ?", id).Where("deleted_at is null").
		Take(&res).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return res, resErr
	}

	return res, errRes
}

func (repo *categoryQueryRepository) GetCategoryList(ctx context.Context) (res []entities.Category, errRes *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetCategoryList")
	defer span.End()

	err := repo.readDb.WithContext(ctx).Model(&entities.Category{}).
		Select("id, name, description, status, created_at, updated_at").
		Find(&res).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return res, resErr
	}

	return res, nil
}
