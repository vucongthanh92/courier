package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

type idempotencyUsecase struct {
	commandRepo interfaces.IdempotencyCommandRepoI
}

func InitIdempotencyUsecase(commandRepo interfaces.IdempotencyCommandRepoI) interfaces.IdempotencyServiceI {
	return &idempotencyUsecase{
		commandRepo: commandRepo,
	}
}

// Claim reserves the request key or returns the previously completed response.
// A key cannot be reused with a different request payload.
func (u *idempotencyUsecase) Claim(ctx context.Context, req models.IdempotencyClaimRequest) (
	*models.IdempotencyClaimResult, *errHandler.ErrorBuilder) {

	// Generate a hash of the request payload to ensure that the same key cannot be reused with a different request.
	requestBody, err := json.Marshal(req.Request)
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	requestSum := sha256.Sum256(requestBody)

	// Generate a unique ID for the idempotency key record using a Snowflake ID generator.
	recordID, err := utils.NewSnowflakeID()
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	// Create a new idempotency key record with the provided scope, user ID, key, request hash, and expiration time.
	record := &entities.IdempotencyKey{
		ID:             recordID,
		Scope:          req.Scope,
		UserID:         &req.UserID,
		IdempotencyKey: req.Key,
		RequestHash:    hex.EncodeToString(requestSum[:]),
		ExpiresAt:      time.Now().UTC().Add(req.ExpiresIn),
	}

	stored, created, commonErr := u.commandRepo.Claim(ctx, record)
	if commonErr != nil {
		return nil, commonErr
	}
	if !created && (stored.RequestHash != record.RequestHash || stored.ResponseStatus == nil) {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusConflict).
			SetError(models.ErrorDTO{
				Code:    "idempotency_conflict",
				Message: "Idempotency-Key is already used with a different or incomplete request",
			})
	}

	return &models.IdempotencyClaimResult{
		Record:       &models.IdempotencyKeyDTO{ID: stored.ID},
		Created:      created,
		ResponseBody: stored.ResponseBody,
	}, nil
}

// Complete updates the status and response of an idempotency key record in the database.
// It takes the context, the ID of the idempotency key, the status to set, and the response body as parameters.
// If the update is successful, it returns nil; otherwise, it returns an error builder with details about the failure.
func (u *idempotencyUsecase) Complete(ctx context.Context, recordID uint64, status int16, response any) *errHandler.ErrorBuilder {
	responseBody, err := json.Marshal(response)
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return u.commandRepo.Complete(ctx, recordID, status, responseBody)
}
