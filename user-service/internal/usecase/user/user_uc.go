package user_uc

import (
	"context"

	"github.com/jinzhu/copier"
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
	userEntity := entities.User{}
	copier.Copy(&userEntity, &req)

	// step 2. check email and phone number exist with user existing

	// step 3. init transaction to create user with
	// table users, email_verification, auth_credentials, audit_log, ...
	err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		_, resErr := s.userWriteRepo.InsertUser(ctx, userEntity)
		if resErr != nil {
			return resErr
		}

		return nil
	})

	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, resErr
	}

	// step 4. handle after created user successfully, send verify email, sms, ...
	// ...

	return nil, nil
}
