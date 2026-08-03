package token

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
)

type TokenUseCaseImpl struct {
	txn              *transaction.ManagerTxn
	auditLogService  interfaces.AuditLogServiceI
	JwtSigner        interfaces.JWTSignerI
	rfTokenWriteRepo interfaces.RefreshTokenCommandRepoI
	tokenDeny        interfaces.TokenDenylistI
}

func InitTokenUseCase(
	txn *transaction.ManagerTxn,
	auditLogService interfaces.AuditLogServiceI,
	JwtSigner interfaces.JWTSignerI,
	rfTokenWriteRepo interfaces.RefreshTokenCommandRepoI,
	tokenDeny interfaces.TokenDenylistI,
) interfaces.TokenUseCaseI {
	return &TokenUseCaseImpl{
		txn:              txn,
		auditLogService:  auditLogService,
		JwtSigner:        JwtSigner,
		rfTokenWriteRepo: rfTokenWriteRepo,
		tokenDeny:        tokenDeny,
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

	if err := s.rfTokenWriteRepo.UpsertByUserAgent(ctx, &rt); err != nil {
		return nil, err
	}

	// Log the successful token generation with audit log (after transaction to ensure we have user ID)
	return &models.JwtTokenResponse{
		AccessToken:      accessToken,
		ExpiresIn:        int64(accessTTL.Seconds()),
		RefreshToken:     refreshPlain,
		RefreshExpiresIn: int64(refreshTTL.Seconds()),
		TokenType:        "Bearer",
		User: &models.AuthenticatedUserResponse{
			ID:          userEntity.ID,
			DisplayName: userEntity.DisplayName,
			AvatarURL:   userEntity.AvatarURL,
		},
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

// RevokeJwtToken implements interfaces.TokenUseCaseI
func (s *TokenUseCaseImpl) RevokeJwtToken(ctx context.Context, claims jwt.MapClaims) *errHandler.ErrorBuilder {

	// tracing for logout usecase, we want to trace the whole flow of logout process, from checking user context,
	ctx, span := tracing.StartSpanFromContext(ctx, "Logout")
	defer span.End()

	jti := claims["jti"].(string)
	exp := int64(claims["exp"].(float64))
	userID := utils.ParseUserID(claims["sub"])

	// calculate TTL for the token, and block it in token denylist with the same TTL,
	// so it will be automatically removed from denylist when expired
	ttl := time.Until(time.Unix(exp, 0))
	if ttl > 0 {
		if err := s.tokenDeny.Block(ctx, jti, ttl); err != nil {
			return errHandler.InitErrorBuilder(ctx).SetLogError(err).SetStatus(500)
		}
	}

	// revoke all refresh tokens of the user to force logout from all devices,
	// we can optimize this by only revoking the current refresh token if we have jti stored in refresh token table
	errCommon := s.rfTokenWriteRepo.RevokeByUser(ctx, userID, time.Now())
	if errCommon != nil {
		return errCommon
	}

	return nil
}
