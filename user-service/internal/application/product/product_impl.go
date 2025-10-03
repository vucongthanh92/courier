package product

import (
	"context"

	errHandler "github.com/vucongthanh92/go-base-project/helper/error_handler"
	"github.com/vucongthanh92/go-base-project/internal/domain/entities"
	"github.com/vucongthanh92/go-base-project/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-project/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type ProductImpl struct {
	productReadRepo interfaces.ProductQueryRepoI
}

func NewProductService(productReadRepo interfaces.ProductQueryRepoI) interfaces.ProductService {
	return &ProductImpl{
		productReadRepo: productReadRepo,
	}
}

func (s *ProductImpl) CreateProduct(ctx context.Context) error {
	return nil
}

func (s *ProductImpl) GetProductsByFilter(ctx context.Context, req models.ProductListFilter) (
	response []entities.Product, totalRows int64, resErr *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "GetProductsByFilter")
	defer func() {
		span.End()
	}()

	response, totalRows, resErr = s.productReadRepo.GetProductByFilter(ctx, req)
	return response, totalRows, resErr
}

func (s *ProductImpl) GetProductByID(ctx context.Context) error {
	return nil
}

func (s *ProductImpl) UpdateProductByID(ctx context.Context) error {
	return nil
}

func (s *ProductImpl) DeleteProductByID(ctx context.Context) error {
	return nil
}
