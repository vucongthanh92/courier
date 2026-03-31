package identity

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
func (repo *identityCmdRepository) InserIdentity(ctx context.Context, entity entities.Identity) (
	entities.Identity, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InserIdentity")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert identity record
	err := run.Model(entities.Identity{}).Create(&entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return entity, resErr
	}

	return entity, nil
}

// UpsertIdentity performs an upsert operation for the given identity entity,
// inserting it if it doesn't exist or updating it if it does, based on the provider and provider UID.
func (repo *identityCmdRepository) UpsertIdentity(ctx context.Context, entity entities.Identity) (
	entities.Identity, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "UpsertIdentity")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Perform upsert operation using GORM's OnConflict clause
	err := run.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "provider"}, {Name: "provider_uid"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "email_at_auth", "scopes", "expires_at", "access_token_enc", "refresh_token_enc"}),
	}).Create(&entity).Error
	if err != nil {
		return entity, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return entity, nil
}
