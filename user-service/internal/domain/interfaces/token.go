package interfaces

import (
	"context"

	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type TokenUseCaseI interface {
	GenerateJwtToken(ctx context.Context, userEntity *entities.User) (*models.JwtTokenResponse, *errHandler.ErrorBuilder)
	RenewJwtToken(ctx context.Context, userEntity *entities.User) (*models.RenewTokenResponse, *errHandler.ErrorBuilder)
	RevokeJwtToken(ctx context.Context, claims jwt.MapClaims) *errHandler.ErrorBuilder
}
