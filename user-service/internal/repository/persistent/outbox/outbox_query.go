package outbox

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	"github.com/vucongthanh92/go-base-utils/helper/transaction"
	utilsModels "github.com/vucongthanh92/go-base-utils/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"
)

type outboxQueryRepository struct {
	readDb *gorm.DB
}

func InitOutboxQueryRepository(readDb *database.GormReadDb) interfaces.OutboxQueryRepoI {
	return &outboxQueryRepository{
		readDb: *readDb,
	}
}

func (repo *outboxQueryRepository) GetOutboxByID(ctx context.Context, id uint64) (
	*entities.Outbox, *utilsError.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "GetOutboxByID")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	// Fetch outbox event by ID
	var outbox entities.Outbox
	err := run.Model(entities.Outbox{}).Where("id = ?", id).First(&outbox).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utilsError.InitErrorBuilder(ctx).
				SetLogError(err).
				SetStatus(404).
				SetError(utilsModels.ErrorDTO{
					Code:    "OUTBOX_NOT_FOUND",
					Message: "Outbox event not found",
				})
		}
		resErr := utilsError.InitErrorBuilder(ctx).ValidateError(err)
		return nil, resErr
	}

	return &outbox, nil
}
