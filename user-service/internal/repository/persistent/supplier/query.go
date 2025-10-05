package supplier

import (
	"github.com/vucongthanh92/courier/user-service/database"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type supplierQueryRepository struct {
	readDb *gorm.DB
}

func NewSupplierQueryRepository(readDb *database.GormReadDb) interfaces.SupplierQueryRepoI {
	return &supplierQueryRepository{
		readDb: *readDb,
	}
}
