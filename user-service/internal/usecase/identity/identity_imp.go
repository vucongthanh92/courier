package identity

import (
	"context"

	"github.com/jinzhu/copier"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type IdentityServiceImpl struct {
	identityWriteRepo interfaces.IdentityCommandRepoI
	identityReadRepo  interfaces.IdentityQueryRepoI
}

func InitIdentityService(
	readRepo interfaces.IdentityQueryRepoI,
	writeRepo interfaces.IdentityCommandRepoI,
) interfaces.IdentityServiceI {
	return &IdentityServiceImpl{
		identityReadRepo:  readRepo,
		identityWriteRepo: writeRepo,
	}
}

func (s *IdentityServiceImpl) CreateIdentity(ctx context.Context, req models.CreateIdentityParams) (
	entities.Identity, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "CreateIdentity")
	defer span.End()

	identityEntity := entities.Identity{}
	copier.Copy(&identityEntity, &req)

	res, resErr := s.identityWriteRepo.InserIdentity(ctx, identityEntity)
	if resErr != nil {
		return res, resErr
	}

	return res, nil
}
