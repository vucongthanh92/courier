package auditlog

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/utils"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

type auditLogUsecase struct {
	auditLogCmd interfaces.AuditLogCommandRepoI
}

func InitAuditLogUsecase(auditLogCmd interfaces.AuditLogCommandRepoI) interfaces.AuditLogServiceI {
	return &auditLogUsecase{
		auditLogCmd: auditLogCmd,
	}
}

func (u *auditLogUsecase) Write(ctx context.Context, input models.AuditLogInput) *errHandler.ErrorBuilder {
	id, err := utils.NewSnowflakeID()
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	auditLog := entities.NewAuditLog(
		id,
		input.ActorType,
		input.ActorID,
		input.Action,
		input.ResourceType,
		input.ResourceID,
		input.IP,
		input.Metadata,
	)

	return u.auditLogCmd.Create(ctx, auditLog)
}
