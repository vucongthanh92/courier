package interfaces

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type AuditLogQueryRepoI interface {
}

type AuditLogCommandRepoI interface {
	InsertAuditLog(ctx context.Context, entity entities.AuditLog) (
		entities.AuditLog, *utilsError.ErrorBuilder)
}

type AuditLogServiceI interface {
	CreateAuditLog(ctx context.Context, req models.AuditLogRequest)
}
