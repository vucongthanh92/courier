package interfaces

import (
	"context"

	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type TokenUseCaseI interface {
	GenerateJwtToken(ctx context.Context, userEntity *entities.User) (*models.JwtTokenResponse, *utilsError.ErrorBuilder)
	RenewJwtToken(ctx context.Context, userEntity *entities.User) (*models.RenewTokenResponse, *utilsError.ErrorBuilder)
	RevokeJwtToken(ctx context.Context, claims jwt.MapClaims) *utilsError.ErrorBuilder
}
