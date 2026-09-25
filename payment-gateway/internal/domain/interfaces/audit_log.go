package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

type AuditLogCommandRepoI interface {
	Create(context.Context, *entities.AuditLog) *errHandler.ErrorBuilder
}

type AuditLogServiceI interface {
	Write(context.Context, models.AuditLogInput) *errHandler.ErrorBuilder
}
