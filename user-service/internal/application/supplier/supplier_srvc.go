package supplier

import "context"

type SupllierService interface {
	CreateSupplier(ctx context.Context) error
	GetSupplierByID(ctx context.Context) error
	UpdateSupplierByID(ctx context.Context) error
	DeleteSupplierByID(ctx context.Context) error
}
