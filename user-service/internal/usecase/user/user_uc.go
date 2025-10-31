package user_uc

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type UserUseCaseImpl struct {
	txn               *transaction.ManagerTxn
	userReadRepo      interfaces.UserQueryRepoI
	userWriteRepo     interfaces.UserCommandRepoI
	authCredWriteRepo interfaces.AuthCredentialCommandRepoI
	emailVerWriteRepo interfaces.EmailVerificationCommandRepoI
	auditLogWriteRepo interfaces.AuditLogCommandRepoI
	outboxWriteRepo   interfaces.OutboxCommandRepoI
}

func InitUserUsecase(
	txn *transaction.ManagerTxn,
	userReadRepo interfaces.UserQueryRepoI,
	userWriteRepo interfaces.UserCommandRepoI,
	authCredWriteRepo interfaces.AuthCredentialCommandRepoI,
	emailVerWriteRepo interfaces.EmailVerificationCommandRepoI,
	auditLogWriteRepo interfaces.AuditLogCommandRepoI,
	outboxWriteRepo interfaces.OutboxCommandRepoI,
) interfaces.UserServiceI {
	return &UserUseCaseImpl{
		txn:               txn,
		userReadRepo:      userReadRepo,
		userWriteRepo:     userWriteRepo,
		authCredWriteRepo: authCredWriteRepo,
		emailVerWriteRepo: emailVerWriteRepo,
		auditLogWriteRepo: auditLogWriteRepo,
		outboxWriteRepo:   outboxWriteRepo,
	}
}

// Signup implements interfaces.UserServiceI
func (s *UserUseCaseImpl) Signup(ctx context.Context, req models.SignupRequest) (
	*entities.User, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "Signup")
	defer span.End()

	// step 1. Map request to entity
	var (
		userEntity        = entities.User{}
		emailVerifyEntity = entities.EmailVerification{}
		authCredEntity    = entities.AuthCredential{}
	)

	req.MappingToUserEntity(&userEntity)
	req.MappingToEmailVerifyEntity(&emailVerifyEntity)
	req.MappingToAuthCredEntity(&authCredEntity)

	// step 2. check email and phone number exist with user existing
	existed, commonErr := s.userReadRepo.CheckExistingEmailOrPhone(ctx, req.Email, req.PhoneNumber)
	if commonErr != nil {
		return nil, commonErr
	}

	if existed {
		commonErr := errHandler.InitErrorBuilder(ctx).
			SetLogError(nil).
			SetStatus(400).
			SetError(models.ErrorDTO{
				Code:    "USER_ALREADY_EXISTS",
				Message: "Email or phone number already in use",
				Field:   "email/phone_number",
			})
		return nil, commonErr
	}

	// step 3. init transaction to create user with
	// table users, email_verification, auth_credentials, audit_log, ...
	err := s.txn.Do(ctx, func(txCtx context.Context) (txnErr *errHandler.ErrorBuilder) {

		// create user
		txnErr = s.userWriteRepo.InsertUser(txCtx, &userEntity)
		if txnErr != nil {
			return txnErr
		}

		// create email verification and auth credential
		emailVerifyEntity.UserID = userEntity.ID
		txnErr = s.emailVerWriteRepo.InsertEmailVerification(txCtx, &emailVerifyEntity)
		if txnErr != nil {
			return txnErr
		}

		// create auth credential
		authCredEntity.UserID = userEntity.ID
		txnErr = s.authCredWriteRepo.InsertAuthCredential(txCtx, &authCredEntity)
		if txnErr != nil {
			return txnErr
		}

		return nil
	})

	// handle error when create user failed
	if err != nil {
		commonErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, commonErr
	}

	// step 4. handle after created user successfully, send verify email, sms, ...
	// ...

	return nil, nil
}
