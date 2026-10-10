package refreshtoken

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	"github.com/vucongthanh92/go-base-utils/helper/transaction"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"
)

type refreshTokenQueryRepo struct {
	readDb *gorm.DB
}

func InitRefreshTokenQueryRepository(readDb *database.GormReadDb) interfaces.RefreshTokenQueryRepoI {
	return &refreshTokenQueryRepo{readDb: *readDb}
}

// GetByTokenHash retrieves a refresh token entity based on the provided token hash.
func (r *refreshTokenQueryRepo) GetByTokenHash(ctx context.Context, hash string) (
	entities.RefreshToken, *utilsError.ErrorBuilder) {

	// Start tracing span for the GetByTokenHash operation
	ctx, span := tracing.StartSpanFromContext(ctx, "GetRefreshTokenByHash")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.readDb)

	var res entities.RefreshToken
	err := run.Model(&entities.RefreshToken{}).Where("token_hash = ?", hash).Take(&res).Error
	if err != nil {
		return res, utilsError.InitErrorBuilder(ctx).ValidateError(err)
	}
	return res, nil
}
