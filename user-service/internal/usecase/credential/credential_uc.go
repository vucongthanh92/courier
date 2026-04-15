package credential

import (
	"context"
	"net/http"

	"github.com/vucongthanh92/courier/user-service/helper/constants"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/datatypes"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
)

type AuthCredentialUseCaseImpl struct {
	txn                        *transaction.ManagerTxn
	auditLogService            interfaces.AuditLogServiceI
	outboxService              interfaces.OutboxServiceI
	userReadRepo               interfaces.UserQueryRepoI
	userWriteRepo              interfaces.UserCommandRepoI
	authCredWriteRepo          interfaces.AuthCredentialCommandRepoI
	authCredReadRepo           interfaces.AuthCredentialQueryRepoI
	refreshTokenReadRepo       interfaces.RefreshTokenQueryRepoI
	emailVerificationWriteRepo interfaces.EmailVerificationCommandRepoI
	emailVerificationReadRepo  interfaces.EmailVerificationQueryRepoI
	tokenService               interfaces.TokenUseCaseI
}

func InitCredentialUseCase(
	txn *transaction.ManagerTxn,
	auditLogService interfaces.AuditLogServiceI,
	outboxService interfaces.OutboxServiceI,
	userReadRepo interfaces.UserQueryRepoI,
	userWriteRepo interfaces.UserCommandRepoI,
	authCredWriteRepo interfaces.AuthCredentialCommandRepoI,
	authCredReadRepo interfaces.AuthCredentialQueryRepoI,
	emailVerificationWriteRepo interfaces.EmailVerificationCommandRepoI,
	emailVerificationReadRepo interfaces.EmailVerificationQueryRepoI,
	refreshTokenReadRepo interfaces.RefreshTokenQueryRepoI,
	tokenService interfaces.TokenUseCaseI,
) interfaces.AuthCredentialServiceI {
	return &AuthCredentialUseCaseImpl{
		txn:                        txn,
		auditLogService:            auditLogService,
		outboxService:              outboxService,
		userReadRepo:               userReadRepo,
		userWriteRepo:              userWriteRepo,
		authCredWriteRepo:          authCredWriteRepo,
		authCredReadRepo:           authCredReadRepo,
		emailVerificationWriteRepo: emailVerificationWriteRepo,
		emailVerificationReadRepo:  emailVerificationReadRepo,
		refreshTokenReadRepo:       refreshTokenReadRepo,
		tokenService:               tokenService,
	}
}

// SetPassword implements AuthCredentialServiceI
func (s *AuthCredentialUseCaseImpl) SetPassword(ctx context.Context, req models.GeneratePasswordRequest) *errHandler.ErrorBuilder {

	// step 1. start transaction
	ctx, span := tracing.StartSpanFromContext(ctx, "SetPassword")
	defer span.End()

	// step 2. get auth credential by user ID
	credential, commonErr := s.authCredReadRepo.GetByUserID(ctx, req.UserID)
	if commonErr != nil {
		return commonErr
	}

	if credential.PasswordVersion != 0 || credential.PasswordHash != "" {
		return errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusConflict).
			SetError(models.ErrorDTO{Code: "password_already_set", Message: "Password already exists"})
	}

	// step 3. update password and password version
	// We can consider to have a separate table to store password history for better audit and security in the future
	credential.MappingToAuthCredEntity("bcrypt", req.Password, 1)
	if commonErr := s.authCredWriteRepo.UpdatePassword(ctx, &credential); commonErr != nil {
		return commonErr
	}

	// step 4. insert audit log for user set password action
	s.auditLogService.CreateAuditLog(ctx,
		models.AuditLogRequest{
			CreatorID: req.UserID,
			Action:    constants.AuditLogActionSetPassword,
			IP:        utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),
			Metadata: datatypes.JSONMap{
				"auth_credential": credential,
			},
		})

	return nil
}
