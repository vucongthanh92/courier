package supplier

import (
	"github.com/vucongthanh92/go-base-project/database"
	"gorm.io/gorm"

	"github.com/vucongthanh92/go-base-project/internal/domain/interfaces"
)

type supplierCommandRepository struct {
	writeDB *gorm.DB
}

func NewSupplierCommandRepository(writeDB *database.GormWriteDb) interfaces.SupplierCommandRepoI {
	return &supplierCommandRepository{
		writeDB: *writeDB,
	}
}
