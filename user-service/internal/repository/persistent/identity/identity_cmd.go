package identity

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type identityCmdRepository struct {
	writeDb *gorm.DB
}

// InitIdentityCmdRepository initializes a new instance of identityCmdRepository with the provided write database connection.
func InitIdentityCmdRepository(writeDb *database.GormWriteDb) interfaces.IdentityCommandRepoI {
	return &identityCmdRepository{
		writeDb: *writeDb,
	}
}

// InserIdentity inserts a new identity record into the database and returns the created entity along with any potential error.
func (repo *identityCmdRepository) InserIdentity(ctx context.Context, entity *entities.Identity) *errHandler.ErrorBuilder {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InserIdentity")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert identity record
	err := run.Model(entities.Identity{}).Create(entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return resErr
	}

	return nil
}
