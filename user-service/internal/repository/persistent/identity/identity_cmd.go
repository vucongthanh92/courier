package product

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type identityCmdRepository struct {
	writeDb *gorm.DB
}

func InitIdentityCmdRepository(writeDb *database.GormWriteDb) interfaces.IdentityCommandRepoI {
	return &identityCmdRepository{
		writeDb: *writeDb,
	}
}

func (repo *identityCmdRepository) InserIdentity(ctx context.Context, entity entities.Identity) (
	entities.Identity, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "InserIdentity")
	defer span.End()

	err := repo.writeDb.WithContext(ctx).Model(entities.Identity{}).
		Create(&entity).Error

	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return entity, resErr
	}

	return entity, nil
}
