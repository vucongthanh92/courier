package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type AuditLogQueryRepoI interface {
}

type AuditLogCommandRepoI interface {
	InsertAuditLog(ctx context.Context, entity entities.AuditLog) (
		entities.AuditLog, *errHandler.ErrorBuilder)
}

type AuditLogServiceI interface {
}
