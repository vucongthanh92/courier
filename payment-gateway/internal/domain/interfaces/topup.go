package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

type TopUpQueryRepoI interface {
	GetForUpdateByPaymentCode(context.Context, string, string) (*entities.TopUpIntent, *errHandler.ErrorBuilder)
}

type TopUpCommandRepoI interface {
	Create(context.Context, *entities.TopUpIntent) *errHandler.ErrorBuilder
	MarkSucceeded(context.Context, uint64, string) *errHandler.ErrorBuilder
}

type TopUpServiceI interface {
	CreateTopUp(context.Context, models.CreateTopUpRequest) (models.CheckoutInstruction, bool, *errHandler.ErrorBuilder)
}
