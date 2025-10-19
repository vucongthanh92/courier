package outbox

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

type outboxCmdRepository struct {
	writeDb *gorm.DB
}

func InitOutboxCmdRepository(writeDb *database.GormWriteDb) interfaces.OutboxCommandRepoI {
	return &outboxCmdRepository{
		writeDb: *writeDb,
	}
}

func (repo *outboxCmdRepository) InsertOutbox(ctx context.Context, entity entities.Outbox) (
	entities.Outbox, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertOutbox")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert outbox record
	err := run.Model(entities.Outbox{}).Create(&entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return entity, resErr
	}

	return entity, nil
}
