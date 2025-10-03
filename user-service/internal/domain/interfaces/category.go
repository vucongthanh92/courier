package interfaces

import (
	"context"
	"time"

	errHandler "github.com/vucongthanh92/go-base-project/helper/error_handler"
	"github.com/vucongthanh92/go-base-project/internal/domain/entities"
	"github.com/vucongthanh92/go-base-project/internal/domain/models"
)

// repository interface
type CategoryQueryRepoI interface {
	GetCategoryList(ctx context.Context) (res []entities.Category, errRes *errHandler.ErrorBuilder)
	GetCategoryByID(ctx context.Context, id uint64) (res entities.Category, errRes *errHandler.ErrorBuilder)
}

type CategoryCommandRepoI interface {
	InsertCategory(ctx context.Context, entity entities.Category) (entities.Category, *errHandler.ErrorBuilder)
	UpdateCategory(ctx context.Context, entity entities.Category) (entities.Category, *errHandler.ErrorBuilder)
	SoftDeleteCategoryByID(ctx context.Context, id uint64, updatedAt time.Time) *errHandler.ErrorBuilder
}

// service interface
type CategoryServiceI interface {
	CreateCategory(ctx context.Context, req models.CreateCategoryReq) (entities.Category, *errHandler.ErrorBuilder)
	UpdateCategory(ctx context.Context, req models.UpdateCategoryReq) (res entities.Category, resErr *errHandler.ErrorBuilder)
	DeleteCategoryByID(ctx context.Context, categoryID uint64, updatedAt time.Time) *errHandler.ErrorBuilder
	GetCategoryList(ctx context.Context) ([]entities.Category, *errHandler.ErrorBuilder)
	GetCategoryByID(ctx context.Context, categoryID int64) (entities.Category, *errHandler.ErrorBuilder)
}
