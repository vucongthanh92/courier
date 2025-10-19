package identity

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type identityQueryRepository struct {
	readDb *gorm.DB
}

func InitIdentityQueryRepository(readDb *database.GormReadDb) interfaces.IdentityQueryRepoI {
	return &identityQueryRepository{
		readDb: *readDb,
	}
}

func (repo *identityQueryRepository) GetIdentityByID(ctx context.Context, id uint64) (
	res entities.Identity, errRes *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "GetIdentityByID")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	// Query identity by ID
	err := run.Model(&entities.Identity{}).
		Select("*").
		Where("id = ?", id).Where("deleted_at is null").
		Take(&res).Error

	// Handle potential errors
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return res, resErr
	}

	return res, errRes
}
