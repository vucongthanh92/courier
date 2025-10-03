package category

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jinzhu/copier"
	"github.com/vucongthanh92/go-base-project/helper/constants"
	errHandler "github.com/vucongthanh92/go-base-project/helper/error_handler"
	"github.com/vucongthanh92/go-base-project/internal/domain/entities"
	"github.com/vucongthanh92/go-base-project/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-project/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type CategoryImpl struct {
	categoryReadRepo  interfaces.CategoryQueryRepoI
	categoryWriteRepo interfaces.CategoryCommandRepoI
	productReadRepo   interfaces.ProductQueryRepoI
}

func NewCategoryService(
	categoryReadRepo interfaces.CategoryQueryRepoI,
	categoryWriteRepo interfaces.CategoryCommandRepoI,
	productReadRepo interfaces.ProductQueryRepoI,
) interfaces.CategoryServiceI {
	return &CategoryImpl{
		categoryReadRepo:  categoryReadRepo,
		categoryWriteRepo: categoryWriteRepo,
		productReadRepo:   productReadRepo,
	}
}

func (s *CategoryImpl) CreateCategory(ctx context.Context, req models.CreateCategoryReq) (
	entities.Category, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "GetCategoryList")
	defer span.End()

	categoryEntity := entities.Category{}
	copier.Copy(&categoryEntity, &req)

	res, resErr := s.categoryWriteRepo.InsertCategory(ctx, categoryEntity)
	if resErr != nil {
		return res, resErr
	}

	return res, nil
}

func (s *CategoryImpl) UpdateCategory(ctx context.Context, req models.UpdateCategoryReq) (
	res entities.Category, resErr *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "UpdateCategory")
	defer span.End()

	categoryEntity, resErr := s.categoryReadRepo.GetCategoryByID(ctx, req.ID)
	if resErr != nil {
		return res, resErr
	}

	copier.Copy(&categoryEntity, &req)

	res, resErr = s.categoryWriteRepo.UpdateCategory(ctx, categoryEntity)
	if resErr != nil {
		return res, resErr
	}

	return res, nil
}

func (s *CategoryImpl) DeleteCategoryByID(ctx context.Context, categoryID uint64, updatedAt time.Time) *errHandler.ErrorBuilder {

	ctx, span := tracing.StartSpanFromContext(ctx, "DeleteCategoryByID")
	defer span.End()

	_, resErr := s.categoryReadRepo.GetCategoryByID(ctx, categoryID)
	if resErr != nil {
		return resErr
	}

	totalProduct, resErr := s.productReadRepo.CountProductByCategoryID(ctx, categoryID)
	if resErr != nil {
		return resErr
	}

	if totalProduct > 0 {
		resErr := errHandler.InitErrorBuilder(ctx).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Message: "category has products",
				Code:    constants.InvalidValue,
			}).
			SetIsMultipleError(false).
			SetIsSystemError(false)
		return resErr
	}

	resErr = s.categoryWriteRepo.SoftDeleteCategoryByID(ctx, categoryID, updatedAt)
	if resErr != nil {
		return resErr
	}

	return nil
}

func (s *CategoryImpl) GetCategoryList(ctx context.Context) ([]entities.Category, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "GetCategoryList")
	defer span.End()

	res, errRes := s.categoryReadRepo.GetCategoryList(ctx)
	return res, errRes
}

func (s *CategoryImpl) GetCategoryByID(ctx context.Context, categoryID int64) (entities.Category, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "GetCategoryList")
	defer span.End()

	res, errRes := s.categoryReadRepo.GetCategoryByID(ctx, uint64(categoryID))
	return res, errRes
}
