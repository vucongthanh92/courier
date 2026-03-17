package auth_uc

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/vucongthanh92/courier/user-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/datatypes"
)

type AuthUseCaseImpl struct {
	txn                        *transaction.ManagerTxn
	userReadRepo               interfaces.UserQueryRepoI
	userWriteRepo              interfaces.UserCommandRepoI
	authCredWriteRepo          interfaces.AuthCredentialCommandRepoI
	emailVerificationWriteRepo interfaces.EmailVerificationCommandRepoI
	emailVerificationReadRepo  interfaces.EmailVerificationQueryRepoI
	auditLogWriteRepo          interfaces.AuditLogCommandRepoI
	outboxWriteRepo            interfaces.OutboxCommandRepoI
}

func InitAuthUsecase(
	txn *transaction.ManagerTxn,
	userReadRepo interfaces.UserQueryRepoI,
	userWriteRepo interfaces.UserCommandRepoI,
	authCredWriteRepo interfaces.AuthCredentialCommandRepoI,
	emailVerificationWriteRepo interfaces.EmailVerificationCommandRepoI,
	emailVerificationReadRepo interfaces.EmailVerificationQueryRepoI,
	auditLogWriteRepo interfaces.AuditLogCommandRepoI,
	outboxWriteRepo interfaces.OutboxCommandRepoI,
) interfaces.AuthServiceI {
	return &AuthUseCaseImpl{
		txn:                        txn,
		userReadRepo:               userReadRepo,
		userWriteRepo:              userWriteRepo,
		authCredWriteRepo:          authCredWriteRepo,
		emailVerificationWriteRepo: emailVerificationWriteRepo,
		emailVerificationReadRepo:  emailVerificationReadRepo,
		auditLogWriteRepo:          auditLogWriteRepo,
		outboxWriteRepo:            outboxWriteRepo,
	}
}

// Signup implements interfaces.UserServiceI
func (s *AuthUseCaseImpl) Signup(ctx context.Context, req models.SignupRequest) (
	*entities.User, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "Signup")
	defer span.End()

	// step 1. Map request to entity
	var (
		userEntity        = entities.User{}
		emailVerifyEntity = entities.EmailVerification{
			TokenHash: utils.RandString(7), // generate token hash for email verification
		}
		authCredEntity = entities.AuthCredential{}
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
				Code:    constants.USER_ALREADY_EXISTS,
				Message: constants.UserAlreadyExistsMessage,
				Field:   "email or phone_number",
			})
		return nil, commonErr
	}

	// step 3. init transaction to create user with
	// table users, email_verification, auth_credentials, outbox, ...
	err := s.txn.Do(ctx, func(txCtx context.Context) (txnErr *errHandler.ErrorBuilder) {

		// create user
		txnErr = s.userWriteRepo.InsertUser(txCtx, &userEntity)
		if txnErr != nil {
			return txnErr
		}

		// create email verification and auth credential
		emailVerifyEntity.UserID = userEntity.ID
		txnErr = s.emailVerificationWriteRepo.InsertEmailVerification(txCtx, &emailVerifyEntity)
		if txnErr != nil {
			return txnErr
		}

		// create auth credential
		authCredEntity.UserID = userEntity.ID
		txnErr = s.authCredWriteRepo.InsertAuthCredential(txCtx, &authCredEntity)
		if txnErr != nil {
			return txnErr
		}

		// create outbox event for USER_CREATED
		txnErr = s.createUserCreatedOutboxEvent(txCtx, &userEntity)
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

	// step 5. insert audit log for user signup action
	audit := entities.AuditLog{
		UserID:    userEntity.ID,
		Action:    "USER_SIGNUP",
		IP:        utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),
		Metadata: datatypes.JSONMap{
			"user": userEntity,
		},
	}

	// decision: ignore or return?
	_, logErr := s.auditLogWriteRepo.InsertAuditLog(ctx, audit)
	if logErr != nil {
		// Common approach: log and continue
		logErr.ExposeLogError() // or logger.Warn(...)
	}

	return nil, nil
}

// VerifyEmail implements interfaces.UserServiceI
func (s *AuthUseCaseImpl) VerifyEmail(ctx context.Context, req models.VerifyEmailRequest) (*models.VerifyEmailResponse, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "VerifyEmail")
	defer span.End()

	ver, resErr := s.emailVerificationReadRepo.GetActiveByEmail(ctx, req.Email)
	if resErr != nil {
		return nil, resErr
	}

	if ver.TokenHash != req.Token {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "invalid_token", Message: "Token is invalid"})
	}
	if time.Now().After(ver.ExpiresAt) {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "token_expired", Message: "Token is expired"})
	}

	// mark token used and update user email_verified status in transaction
	err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		if txnErr := s.emailVerificationWriteRepo.MarkUsed(txCtx, ver.ID, time.Now()); txnErr != nil {
			return txnErr
		}
		if txnErr := s.userWriteRepo.UpdateEmailVerified(txCtx, ver.UserID, "verified"); txnErr != nil {
			return txnErr
		}
		return nil
	})
	if err != nil {
		commonErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, commonErr
	}

	return &models.VerifyEmailResponse{Message: "Email verified"}, nil
}

// ResendVerifyEmail implements interfaces.UserServiceI
// this API will generate new token and expiry,
// then update to email_verification table, and publish outbox event for sending email
func (s *AuthUseCaseImpl) ResendVerifyEmail(ctx context.Context, req models.ResendVerifyEmailRequest) (*models.ResendVerifyEmailResponse, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "ResendVerifyEmail")
	defer span.End()

	// generate new token and expiry
	token := utils.RandString(7)
	expiresAt := time.Now().Add(24 * time.Hour)

	err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		// if no existing row -> insert
		_, qErr := s.emailVerificationReadRepo.GetActiveByEmail(txCtx, req.Email)
		if qErr != nil {
			// create a new record if not found
			sid, _ := utils.NewSnowflakeID()
			entity := entities.EmailVerification{
				ID:        sid,
				Email:     req.Email,
				UserID:    0, // set if you can look up user by email
				TokenHash: token,
				ExpiresAt: expiresAt,
			}
			return s.emailVerificationWriteRepo.InsertEmailVerification(txCtx, &entity)
		}
		// update existing
		return s.emailVerificationWriteRepo.UpdateToken(txCtx, req.Email, token, expiresAt)
	})
	if err != nil {
		commonErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, commonErr
	}

	// send email async via outbox
	_ = s.createSendVerifyEmailOutbox(ctx, req.Email, token)

	return &models.ResendVerifyEmailResponse{Message: "Verification email resent"}, nil
}

// helper to publish outbox event
// you can move this to a common place if needed by other usecases
func (s *AuthUseCaseImpl) createSendVerifyEmailOutbox(ctx context.Context, email, token string) *errHandler.ErrorBuilder {
	payload, _ := json.Marshal(map[string]string{"email": email, "token": token})
	outboxID, _ := utils.NewSnowflakeID()
	event := entities.Outbox{
		ID:            outboxID,
		AggregateType: "EmailVerification",
		AggregateID:   email,
		EventType:     "EMAIL_VERIFICATION_SEND",
		Payload:       payload,
	}
	_, err := s.outboxWriteRepo.InsertOutbox(ctx, event)
	return err
}
