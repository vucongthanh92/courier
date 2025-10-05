package supplier

import (
	"github.com/vucongthanh92/courier/user-service/database"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type supplierCommandRepository struct {
	writeDB *gorm.DB
}

func NewSupplierCommandRepository(writeDB *database.GormWriteDb) interfaces.SupplierCommandRepoI {
	return &supplierCommandRepository{
		writeDB: *writeDB,
	}
}
