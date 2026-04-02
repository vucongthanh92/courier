package token

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
)

type TokenUseCaseImpl struct {
	txn                   *transaction.ManagerTxn
	auditLogService       interfaces.AuditLogServiceI
	JwtSigner             interfaces.JWTSignerI
	refreshTokenWriteRepo interfaces.RefreshTokenCommandRepoI
}

func InitTokenUseCase(
	txn *transaction.ManagerTxn,
	auditLogService interfaces.AuditLogServiceI,
	JwtSigner interfaces.JWTSignerI,
	refreshTokenWriteRepo interfaces.RefreshTokenCommandRepoI,
) interfaces.TokenUseCaseI {
	return &TokenUseCaseImpl{
		txn:                   txn,
		auditLogService:       auditLogService,
		JwtSigner:             JwtSigner,
		refreshTokenWriteRepo: refreshTokenWriteRepo,
	}
}

// GenerateJwtToken implements interfaces.TokenUseCaseI
func (s *TokenUseCaseImpl) GenerateJwtToken(ctx context.Context, userEntity *entities.User) (
	*models.JwtTokenResponse, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "GenerateJwtToken")
	defer span.End()

	// Issue tokens
	now := time.Now()
	accessTTL := 30 * time.Minute
	refreshTTL := 90 * 24 * time.Hour
	accessToken, signErr := s.JwtSigner.SignAccessToken(*userEntity, now, accessTTL)
	if signErr != nil {
		return nil, signErr
	}

	// Store refresh token hash in DB for later verification.
	// We will use the same random string as the raw refresh token to return to client,
	// and only store the hash in DB for better security.
	refreshPlain := utils.RandString(64)
	refreshHash := utils.HashPwdBySha256(userEntity.Email, refreshPlain)
	rt := entities.RefreshToken{
		UserID:    userEntity.ID,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(refreshTTL),
		UserAgent: utils.StrPtr(utils.GetUserAgent(ctx)),
		IP:        utils.StrPtr(utils.GetClientIP(ctx)),
	}

	if err := s.refreshTokenWriteRepo.UpsertByUserAgent(ctx, &rt); err != nil {
		return nil, err
	}

	// Log the successful token generation with audit log (after transaction to ensure we have user ID)
	return &models.JwtTokenResponse{
		AccessToken:      accessToken,
		ExpiresIn:        int64(accessTTL.Seconds()),
		RefreshToken:     refreshPlain,
		RefreshExpiresIn: int64(refreshTTL.Seconds()),
		TokenType:        "Bearer",
	}, nil
}

// RenewJwtToken implements interfaces.TokenUseCaseI
func (s *TokenUseCaseImpl) RenewJwtToken(ctx context.Context, userEntity *entities.User) (
	*models.RenewTokenResponse, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "RenewJwtToken")
	defer span.End()

	// if all valid then generate new access token,
	// if refresh token is about to expire (e.g. less than 7 days), then also generate new refresh token,
	accessTTL := 30 * time.Minute
	now := time.Now()
	accessToken, signErr := s.JwtSigner.SignAccessToken(*userEntity, now, accessTTL)
	if signErr != nil {
		return nil, signErr
	}

	// Check if refresh token needs to be renewed (e.g. if it's going to expire in less than 7 days)
	return &models.RenewTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(accessTTL.Seconds()),
		TokenType:   "Bearer",
	}, nil
}
