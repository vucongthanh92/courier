package interfaces

import (
	"context"

	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

// repository interface
type UserQueryRepoI interface {
	GetUserByIdOrEmail(ctx context.Context, req models.GetUserByIdOrEmailRequest) (res *entities.User, errRes *utilsError.ErrorBuilder)
	GetUsersByIDs(ctx context.Context, userIDs []uint64) (res []entities.User, errRes *utilsError.ErrorBuilder)
	SearchUsers(ctx context.Context, req models.SearchUsersRequest) (res []entities.User, errRes *utilsError.ErrorBuilder)
	CheckExistingEmailOrPhone(ctx context.Context, email string, phoneNumber string) (res bool, errRes *utilsError.ErrorBuilder)
}

type UserCommandRepoI interface {
	InsertUser(ctx context.Context, entity *entities.User) *utilsError.ErrorBuilder
	UpdateEmailVerified(ctx context.Context, id uint64, status string) *utilsError.ErrorBuilder
}

// service interface
type AuthServiceI interface {
	Signup(ctx context.Context, req models.SignupRequest) (*entities.User, *utilsError.ErrorBuilder)
	Login(ctx context.Context, req models.LoginRequest) (*models.JwtTokenResponse, *utilsError.ErrorBuilder)
	VerifyEmail(ctx context.Context, req models.VerifyEmailRequest) (*models.VerifyEmailResponse, *utilsError.ErrorBuilder)
	ResendVerifyEmail(ctx context.Context, req models.ResendVerifyEmailRequest) (*models.ResendVerifyEmailResponse, *utilsError.ErrorBuilder)
	RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (*models.RenewTokenResponse, *utilsError.ErrorBuilder)
	Logout(ctx context.Context, claims jwt.MapClaims) *utilsError.ErrorBuilder
}

type UserServiceI interface {
	SearchUsers(ctx context.Context, req models.SearchUsersRequest) ([]models.SearchUserResponse, *utilsError.ErrorBuilder)
}
