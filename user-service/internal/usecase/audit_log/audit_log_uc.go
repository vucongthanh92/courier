package auditlog_uc

import (
	"context"
	"encoding/json"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/datatypes"
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

// LogUserCreated implements interfaces.AuditLogServiceI
func (s *AuditLogUseCaseImpl) LogUserCreated(
	ctx context.Context,
	user *entities.User,
	ip string,
	userAgent string,
) *errHandler.ErrorBuilder {

	ctx, span := tracing.StartSpanFromContext(ctx, "LogUserCreated")
	defer span.End()

	// Serialize user entity to JSON for meta field
	userJSON, err := json.Marshal(user)
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).
			SetLogError(err).
			SetStatus(500).
			SetError(models.ErrorDTO{
				Code:    "AUDIT_LOG_SERIALIZATION_ERROR",
				Message: "Failed to serialize user entity for audit log",
			})
	}

	// Parse JSON to map for JSONB field
	var userMeta map[string]interface{}
	if err := json.Unmarshal(userJSON, &userMeta); err != nil {
		return errHandler.InitErrorBuilder(ctx).
			SetLogError(err).
			SetStatus(500).
			SetError(models.ErrorDTO{
				Code:    "AUDIT_LOG_META_ERROR",
				Message: "Failed to parse user meta",
			})
	}

	// Create audit log entity
	auditLog := entities.AuditLog{
		UserID:    user.ID,
		Action:    "USER_SIGNUP",
		IP:        ip,
		UserAgent: userAgent,
		Metadata:  datatypes.JSONMap(userMeta),
	}

	// Insert audit log
	_, txnErr := s.auditLogWriteRepo.InsertAuditLog(ctx, auditLog)
	if txnErr != nil {
		return txnErr
	}

	return nil
}
