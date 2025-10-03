package supplier

import (
	"context"

	"github.com/vucongthanh92/go-base-project/internal/domain/interfaces"
)

type SupplierImpl struct {
	supplierReadRepo interfaces.SupplierQueryRepoI
}

func NewSupplierService(supplierReadRepo interfaces.SupplierQueryRepoI) SupllierService {
	return &SupplierImpl{
		supplierReadRepo: supplierReadRepo,
	}
}

func (s *SupplierImpl) CreateSupplier(ctx context.Context) error {
	return nil
}

func (s *SupplierImpl) GetSupplierByID(ctx context.Context) error {
	return nil
}

func (s *SupplierImpl) UpdateSupplierByID(ctx context.Context) error {
	return nil
}

func (s *SupplierImpl) DeleteSupplierByID(ctx context.Context) error {
	return nil
}
