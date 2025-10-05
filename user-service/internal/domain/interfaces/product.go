package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type ProductQueryRepoI interface {
	GetProductByFilter(ctx context.Context, filter models.ProductListFilter) (response []entities.Product, totalRows int64, resErr *errHandler.ErrorBuilder)
	CountProductByCategoryID(ctx context.Context, categoryID uint64) (total int64, resErr *errHandler.ErrorBuilder)
}

type ProductCommandRepoI interface {
}

type ProductService interface {
	CreateProduct(ctx context.Context) error
	GetProductsByFilter(ctx context.Context, req models.ProductListFilter) (response []entities.Product, totalRows int64, resErr *errHandler.ErrorBuilder)
	GetProductByID(ctx context.Context) error
	UpdateProductByID(ctx context.Context) error
	DeleteProductByID(ctx context.Context) error
}
