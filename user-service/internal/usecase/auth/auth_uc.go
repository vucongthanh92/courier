package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/vucongthanh92/courier/user-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/logger"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type AuthUseCaseImpl struct {
	txn                        *transaction.ManagerTxn
	auditLogService            interfaces.AuditLogServiceI
	outboxService              interfaces.OutboxServiceI
	userReadRepo               interfaces.UserQueryRepoI
	userWriteRepo              interfaces.UserCommandRepoI
	authCredWriteRepo          interfaces.AuthCredentialCommandRepoI
	authCredReadRepo           interfaces.AuthCredentialQueryRepoI
	refreshTokenWriteRepo      interfaces.RefreshTokenCommandRepoI
	refreshTokenReadRepo       interfaces.RefreshTokenQueryRepoI
	emailVerificationWriteRepo interfaces.EmailVerificationCommandRepoI
	emailVerificationReadRepo  interfaces.EmailVerificationQueryRepoI
	JwtSigner                  interfaces.JWTSignerI
}

func InitAuthUsecase(
	txn *transaction.ManagerTxn,
	auditLogService interfaces.AuditLogServiceI,
	outboxService interfaces.OutboxServiceI,
	userReadRepo interfaces.UserQueryRepoI,
	userWriteRepo interfaces.UserCommandRepoI,
	authCredWriteRepo interfaces.AuthCredentialCommandRepoI,
	authCredReadRepo interfaces.AuthCredentialQueryRepoI,
	emailVerificationWriteRepo interfaces.EmailVerificationCommandRepoI,
	emailVerificationReadRepo interfaces.EmailVerificationQueryRepoI,
	JwtSigner interfaces.JWTSignerI,
	refreshTokenWriteRepo interfaces.RefreshTokenCommandRepoI,
	refreshTokenReadRepo interfaces.RefreshTokenQueryRepoI,
) interfaces.AuthServiceI {
	return &AuthUseCaseImpl{
		txn:                        txn,
		auditLogService:            auditLogService,
		outboxService:              outboxService,
		userReadRepo:               userReadRepo,
		userWriteRepo:              userWriteRepo,
		authCredWriteRepo:          authCredWriteRepo,
		authCredReadRepo:           authCredReadRepo,
		emailVerificationWriteRepo: emailVerificationWriteRepo,
		emailVerificationReadRepo:  emailVerificationReadRepo,
		JwtSigner:                  JwtSigner,
		refreshTokenWriteRepo:      refreshTokenWriteRepo,
		refreshTokenReadRepo:       refreshTokenReadRepo,
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

		// create outbox event for sending verification email
		payload, _ := json.Marshal(map[string]string{"email": emailVerifyEntity.Email, "token": emailVerifyEntity.TokenHash})
		outboxReq := models.CreateOutboxRequest{
			AggregateType: "user_signup",
			AggregateID:   strconv.FormatUint(userEntity.ID, 10),
			EventType:     "email_verification_send",
			Payload:       payload,
		}
		txnErr = s.outboxService.CreateOutbox(txCtx, outboxReq)
		if txnErr != nil {
			return txnErr
		}

		// if all operations in transaction are successful, return nil to commit transaction
		return nil
	})

	// handle error when create user failed
	if err != nil {
		commonErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, commonErr
	}

	// step 5. insert audit log for user signup action
	auditLogReq := models.AuditLogRequest{
		CreatorID: userEntity.ID,
		Action:    "user_signup",
		IP:        utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),
		Metadata: datatypes.JSONMap{
			"user": userEntity,
		},
	}
	logErr := s.auditLogService.CreateAuditLog(ctx, auditLogReq)
	if logErr != nil {
		logger.Error("Failed to log user created event", zap.Any("error", logErr), zap.Uint64("user_id", userEntity.ID))
	}

	// return response
	return nil, nil
}

// VerifyEmail implements interfaces.UserServiceI
// this API will check token valid or not, if valid then mark email verified and token used in transaction
func (s *AuthUseCaseImpl) VerifyEmail(ctx context.Context, req models.VerifyEmailRequest) (
	*models.VerifyEmailResponse, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "VerifyEmail")
	defer span.End()

	verEmail, resErr := s.emailVerificationReadRepo.GetOneByEmail(ctx, req.Email)
	if resErr != nil {
		return nil, resErr
	}

	// check token valid or not (should check token match, expiry and not used before)
	if verEmail.ExpiresAt.Before(time.Now()) {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    "token_expired",
				Message: "Token is expired",
				Field:   "token",
			})
	}

	if verEmail.TokenHash != req.Token {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "invalid_token", Message: "Token is invalid"})
	}
	if time.Now().After(verEmail.ExpiresAt) {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "token_expired", Message: "Token is expired"})
	}

	// mark token used and update user email_verified status in transaction
	err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		if txnErr := s.emailVerificationWriteRepo.MarkUsed(txCtx, verEmail.ID, time.Now()); txnErr != nil {
			return txnErr
		}
		if txnErr := s.userWriteRepo.UpdateEmailVerified(txCtx, verEmail.UserID, "verified"); txnErr != nil {
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
func (s *AuthUseCaseImpl) ResendVerifyEmail(ctx context.Context, req models.ResendVerifyEmailRequest) (
	*models.ResendVerifyEmailResponse, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "ResendVerifyEmail")
	defer span.End()

	// generate new token and expiry
	token := utils.RandString(7)
	expiresAt := time.Now().Add(24 * time.Hour)
	emailVerification := entities.EmailVerification{}

	// update to email_verification table in transaction
	err := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		var txnErr *errHandler.ErrorBuilder
		emailVerification, txnErr = s.emailVerificationReadRepo.GetOneByEmail(txCtx, req.Email)
		if txnErr != nil {
			return txnErr
		}

		// update existing
		return s.emailVerificationWriteRepo.UpdateToken(txCtx, req.Email, token, expiresAt)
	})
	if err != nil {
		commonErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, commonErr
	}

	// send email async via outbox
	payload, _ := json.Marshal(map[string]string{"email": req.Email, "token": token})
	outboxReq := models.CreateOutboxRequest{
		AggregateType: "user_resend_verify",
		AggregateID:   strconv.FormatUint(emailVerification.UserID, 10),
		EventType:     "email_verification_send",
		Payload:       payload,
	}

	_ = s.outboxService.CreateOutbox(ctx, outboxReq)

	// return response
	return &models.ResendVerifyEmailResponse{Message: "Verification email resent"}, nil
}

// Login implements interfaces.UserServiceI
// this API will check user exist with email, then check email verified, then check password match,
// if all valid then generate access token and refresh token, save refresh token to database, return access token and refresh token to client
func (s *AuthUseCaseImpl) Login(ctx context.Context, req models.LoginRequest) (
	*models.LoginResponse, *errHandler.ErrorBuilder) {

	// tracing for login usecase, we want to trace the whole flow of login process, from checking user exist,
	// checking email verified, checking password, generating token, saving refresh token to database
	ctx, span := tracing.StartSpanFromContext(ctx, "Login")
	defer span.End()

	// check user exist with email
	user, errUser := s.userReadRepo.GetUserByEmail(ctx, req.Email)
	if errUser != nil {
		return nil, errUser
	}

	// check email verified or not, if not verified then return error, only allow login when email is verified
	if !user.EmailVerified || user.Status != "verified" {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusForbidden).
			SetError(models.ErrorDTO{Code: "email_not_verified", Message: "Email not verified"})
	}

	// get auth credential by user id, then check password match, if not match return error
	cred, errCred := s.authCredReadRepo.GetByUserID(ctx, user.ID)
	if errCred != nil {
		return nil, errCred
	}

	// support bcrypt, sha256 fallback
	if cred.PasswordAlgo == "bcrypt" {
		if err := utils.CheckPwdByBcrypt(cred.PasswordHash, req.Password); err != nil {
			return nil, errHandler.InitErrorBuilder(ctx).
				SetStatus(http.StatusUnauthorized).
				SetError(models.ErrorDTO{Code: "invalid_credentials", Message: "Invalid credentials"})
		}
	} else {
		expected := utils.HashPwdBySha256(user.Email, req.Password)
		if expected != cred.PasswordHash {
			return nil, errHandler.InitErrorBuilder(ctx).
				SetStatus(http.StatusUnauthorized).
				SetError(models.ErrorDTO{Code: "invalid_credentials", Message: "Invalid credentials"})
		}
	}

	// if all valid then generate access token and refresh token,
	// save refresh token to database, return access token and refresh token to client
	accessTTL := 15 * time.Minute
	refreshTTL := 90 * 24 * time.Hour
	now := time.Now()

	// generate access token and refresh token, then save refresh token to database
	accessToken, commonErr := s.JwtSigner.SignAccessToken(user, now, accessTTL)
	if commonErr != nil {
		return nil, commonErr
	}

	// for refresh token, we will generate random string and save hash to database for security,
	// when refresh token request come, we will hash the token and compare with database
	refreshPlain := utils.RandString(64)
	refreshHash := utils.HashPwdBySha256(user.Email, refreshPlain)

	// save refresh token to database
	rt := entities.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(refreshTTL),
		UserAgent: utils.StrPtr(utils.GetUserAgent(ctx)),
		IP:        utils.StrPtr(utils.GetClientIP(ctx)),
	}
	if err := s.refreshTokenWriteRepo.UpsertByUserAgent(ctx, &rt); err != nil {
		return nil, err
	}

	// return response to client
	res := &models.LoginResponse{
		AccessToken:      accessToken,
		ExpiresIn:        int64(accessTTL.Seconds()),
		RefreshToken:     refreshPlain,
		RefreshExpiresIn: int64(refreshTTL.Seconds()),
		TokenType:        "Bearer",
	}

	// insert audit log for user signup action
	auditLogReq := models.AuditLogRequest{
		CreatorID: user.ID,
		Action:    "user_login",
		IP:        utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),
		Metadata: datatypes.JSONMap{
			"login_response": res,
		},
	}
	logErr := s.auditLogService.CreateAuditLog(ctx, auditLogReq)
	if logErr != nil {
		logger.Error("Failed to log user login event", zap.Any("error", logErr), zap.Uint64("user_id", user.ID))
	}

	// return response to client
	return res, nil
}

// RefreshToken implements interfaces.UserServiceI
func (s *AuthUseCaseImpl) Logout(ctx context.Context, req models.LogoutRequest) (*models.LogoutResponse, *errHandler.ErrorBuilder) {
	return nil, nil
}

// RefreshToken implements interfaces.UserServiceI
func (s *AuthUseCaseImpl) RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (
	*models.RefreshTokenResponse, *errHandler.ErrorBuilder) {
	return nil, nil
}
