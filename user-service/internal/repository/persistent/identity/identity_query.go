package identity

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	"github.com/vucongthanh92/go-base-utils/helper/transaction"
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
	res entities.Identity, errRes *utilsError.ErrorBuilder) {

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
		resErr := utilsError.InitErrorBuilder(ctx).ValidateError(err)
		return res, resErr
	}

	return res, errRes
}

// GetByProviderUID retrieves an identity based on the provider and provider UID.
func (repo *identityQueryRepository) GetByProviderUID(ctx context.Context, provider, providerUID string) (
	*entities.Identity, *utilsError.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "GetByProviderUID")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	var identity entities.Identity
	err := run.Model(&entities.Identity{}).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		Where("deleted_at is null").
		Take(&identity).Error
	if err != nil {
		return nil, utilsError.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &identity, nil
}
