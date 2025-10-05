package product

import (
	"github.com/vucongthanh92/courier/user-service/database"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type productCommandRepository struct {
	writeDb *gorm.DB
}

func NewProductCommandRepository(writeDb *database.GormWriteDb) interfaces.ProductCommandRepoI {
	return &productCommandRepository{
		writeDb: *writeDb,
	}
}
