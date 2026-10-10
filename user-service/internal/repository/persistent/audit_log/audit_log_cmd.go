package auditlog

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"

	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	"github.com/vucongthanh92/go-base-utils/helper/transaction"
)

type auditLogCmdRepository struct {
	writeDb *gorm.DB
}

func InitAuditLogCmdRepository(writeDb *database.GormWriteDb) interfaces.AuditLogCommandRepoI {
	return &auditLogCmdRepository{
		writeDb: *writeDb,
	}
}

func (repo *auditLogCmdRepository) InsertAuditLog(ctx context.Context, entity entities.AuditLog) (
	entities.AuditLog, *utilsError.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertAuditLog")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert audit log record
	err := run.Model(entities.AuditLog{}).Create(&entity).Error
	if err != nil {
		resErr := utilsError.InitErrorBuilder(ctx).ValidateError(err)
		return entity, resErr
	}

	return entity, nil
}
