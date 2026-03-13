package user_uc

import (
	"context"
	"encoding/json"
	"fmt"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

// createUserCreatedOutboxEvent creates an outbox event for USER_CREATED
func (s *UserUseCaseImpl) createUserCreatedOutboxEvent(
	ctx context.Context,
	user *entities.User,
) *errHandler.ErrorBuilder {

	ctx, span := tracing.StartSpanFromContext(ctx, "createUserCreatedOutboxEvent")
	defer span.End()

	// Serialize user entity to JSON
	payload, err := json.Marshal(user)
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).
			SetLogError(err).
			SetStatus(500).
			SetError(models.ErrorDTO{
				Code:    "OUTBOX_SERIALIZATION_ERROR",
				Message: "Failed to serialize user entity",
			})
	}

	// Generate outbox event ID
	outboxID, err := utils.NewSnowflakeID()
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).
			SetLogError(err).
			SetStatus(500).
			SetError(models.ErrorDTO{
				Code:    "OUTBOX_ID_GENERATION_ERROR",
				Message: "Failed to generate outbox event ID",
			})
	}

	// Create outbox event
	outboxEvent := entities.Outbox{
		ID:            outboxID,
		AggregateType: "User",
		AggregateID:   fmt.Sprintf("%d", user.ID),
		EventType:     "USER_CREATED",
		Payload:       payload,
	}

	// Insert outbox event
	_, txnErr := s.outboxWriteRepo.InsertOutbox(ctx, outboxEvent)
	if txnErr != nil {
		return txnErr
	}

	return nil
}
