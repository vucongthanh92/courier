package idempotency

import (
	"context"
	"errors"

	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"gorm.io/gorm"
)

type idempotencyCmdRepository struct {
	writeDB *gorm.DB
}

func InitIdempotencyCmdRepository(db *database.GormWriteDb) interfaces.IdempotencyCommandRepoI {
	return &idempotencyCmdRepository{writeDB: *db}
}

// Claim attempts to claim an idempotency key for a specific user and scope. If the key is successfully claimed, it returns the claimed entity and a boolean indicating success. If the key already exists, it returns the existing entity and a boolean indicating failure.
func (r *idempotencyCmdRepository) Claim(ctx context.Context, entity *entities.IdempotencyKey) (*entities.IdempotencyKey, bool, *errHandler.ErrorBuilder) {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	if err := run.Create(entity).Error; err == nil {
		return entity, true, nil
	} else if !isUnique(err) {
		return nil, false, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	var existing entities.IdempotencyKey
	if err := run.Where("scope = ? AND user_id = ? AND idempotency_key = ?",
		entity.Scope,
		entity.UserID,
		entity.IdempotencyKey,
	).First(&existing).Error; err != nil {
		return nil, false, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return &existing, false, nil
}

// Complete updates the status and response of an idempotency key record in the database. It takes the context, the ID of the idempotency key, the status to set, and the response body as parameters. If the update is successful, it returns nil; otherwise, it returns an error builder with details about the failure.
func (r *idempotencyCmdRepository) Complete(ctx context.Context, id uint64, status int16, response []byte) *errHandler.ErrorBuilder {
	err := transaction.RunnerFromCtx(ctx, r.writeDB).
		Model(&entities.IdempotencyKey{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"response_status": status,
			"response_body":   response,
		}).Error

	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

// isUnique checks if the provided error is a unique constraint violation error, indicating that a record with the same unique key already exists in the database. It returns true if the error is a unique constraint violation, and false otherwise.
func isUnique(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey) || errHandler.InitErrorBuilder(context.Background()).SetLogError(err).IsUniqueViolation()
}
