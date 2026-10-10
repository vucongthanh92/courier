package outbox

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"

	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	baseUtils "github.com/vucongthanh92/go-base-utils/helper/utils"
)

type OutboxUseCaseImpl struct {
	auditLogService interfaces.AuditLogServiceI
	outboxWriteRepo interfaces.OutboxCommandRepoI
	outboxQueryRepo interfaces.OutboxQueryRepoI
}

func InitOutboxUsecase(
	auditLogService interfaces.AuditLogServiceI,
	outboxWriteRepo interfaces.OutboxCommandRepoI,
	outboxQueryRepo interfaces.OutboxQueryRepoI,
) interfaces.OutboxServiceI {
	return &OutboxUseCaseImpl{
		auditLogService: auditLogService,
		outboxWriteRepo: outboxWriteRepo,
		outboxQueryRepo: outboxQueryRepo,
	}
}

// helper to publish outbox event
// you can move this to a common place if needed by other usecases
func (s *OutboxUseCaseImpl) CreateOutbox(
	ctx context.Context, req models.CreateOutboxRequest) *utilsError.ErrorBuilder {

	outboxID, _ := baseUtils.NewSnowflakeID()

	// create outbox event
	event := entities.Outbox{
		ID:            outboxID,
		AggregateType: req.AggregateType,
		AggregateID:   req.AggregateID,
		EventType:     req.EventType,
		Payload:       req.Payload,
	}

	// insert outbox event
	_, err := s.outboxWriteRepo.InsertOutbox(ctx, event)
	return err
}
