package auditlog_uc

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type AuditLogUseCaseImpl struct {
	auditLogWriteRepo interfaces.AuditLogCommandRepoI
}

func InitAuditLogUsecase(
	auditLogWriteRepo interfaces.AuditLogCommandRepoI,
) interfaces.AuditLogServiceI {
	return &AuditLogUseCaseImpl{
		auditLogWriteRepo: auditLogWriteRepo,
	}
}

// CreateAuditLog implements interfaces.AuditLogServiceI
func (s *AuditLogUseCaseImpl) CreateAuditLog(ctx context.Context, req models.AuditLogRequest) *errHandler.ErrorBuilder {

	ctx, span := tracing.StartSpanFromContext(ctx, "CreateAuditLog")
	defer span.End()

	// Create audit log entity
	auditLog := entities.AuditLog{
		UserID:    req.CreatorID,
		Action:    req.Action,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		Metadata:  req.Metadata,
	}

	// Insert audit log
	_, txnErr := s.auditLogWriteRepo.InsertAuditLog(ctx, auditLog)
	if txnErr != nil {
		return txnErr
	}

	return nil
}
