package authcredential

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type authCredentialCmdRepository struct {
	writeDb *gorm.DB
}

func InitAuthCredentialCmdRepository(writeDb *database.GormWriteDb) interfaces.AuthCredentialCommandRepoI {
	return &authCredentialCmdRepository{
		writeDb: *writeDb,
	}
}

func (repo *authCredentialCmdRepository) InsertAuthCredential(ctx context.Context, entity *entities.AuthCredential) *errHandler.ErrorBuilder {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertAuthCredential")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert auth credential record
	err := run.Model(entities.AuthCredential{}).Create(entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return resErr
	}

	return nil
}
