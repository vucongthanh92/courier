package product

import (
	"context"

	"github.com/vucongthanh92/go-base-project/database"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	errHandler "github.com/vucongthanh92/go-base-project/helper/error_handler"
	"github.com/vucongthanh92/go-base-project/internal/domain/entities"
	"github.com/vucongthanh92/go-base-project/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-project/internal/domain/models"
)

type productQueryRepository struct {
	readDb *gorm.DB
}

func NewProductQueryRepository(readDb *database.GormReadDb) interfaces.ProductQueryRepoI {
	return &productQueryRepository{
		readDb: *readDb,
	}
}

func (repo *productQueryRepository) GetProductByFilter(ctx context.Context, filter models.ProductListFilter) (
	response []entities.Product, totalRows int64, resErr *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "GetProductByFilter")
	defer span.End()

	err := repo.readDb.WithContext(ctx).Model(&entities.Product{}).Select("*").Count(&totalRows).
		Limit(filter.Limit).Offset(filter.Offset).Find(&response).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return response, totalRows, resErr
	}

	return response, totalRows, nil
}

func (repo *productQueryRepository) CountProductByCategoryID(ctx context.Context, categoryID uint64) (
	total int64, resErr *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "CountProductByCategoryID")
	defer span.End()

	err := repo.readDb.WithContext(ctx).Model(&entities.Product{}).
		Where("category_id = ?", categoryID).Where("deleted_at is null").
		Count(&total).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return total, resErr
	}

	return total, nil
}
