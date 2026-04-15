package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
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

func InitAuthUseCase(
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
		refreshTokenReadRepo:       refreshTokenReadRepo,
		tokenService:               tokenService,
	}
}

// Signup implements interfaces.AuthServiceI
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

	authCredEntity.MappingToAuthCredEntity("bcrypt", req.Password, 1)

	// step 2. check email and phone number exist with user existing
	existed, commonErr := s.userReadRepo.CheckExistingEmailOrPhone(ctx, req.Email, req.PhoneNumber)
	if commonErr != nil {
		return nil, commonErr
	}

	if existed {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetLogError(nil).
			SetStatus(400).
			SetError(models.ErrorDTO{
				Code:    constants.USER_ALREADY_EXISTS,
				Message: constants.UserAlreadyExistsMessage,
				Field:   "email or phone_number",
			})
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
		payload, _ := json.Marshal(map[string]string{
			"email": emailVerifyEntity.Email,
			"token": emailVerifyEntity.TokenHash,
		})

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
	s.auditLogService.CreateAuditLog(ctx,
		models.AuditLogRequest{
			CreatorID: userEntity.ID,
			Action:    constants.AuditLogActionSignup,
			IP:        utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),
			Metadata: datatypes.JSONMap{
				"user": userEntity,
			},
		})

	// return response
	return &userEntity, nil
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
	*models.JwtTokenResponse, *errHandler.ErrorBuilder) {

	// tracing for login usecase, we want to trace the whole flow of login process, from checking user exist,
	// checking email verified, checking password, generating token, saving refresh token to database
	ctx, span := tracing.StartSpanFromContext(ctx, "Login")
	defer span.End()

	// check user exist with email
	user, errUser := s.userReadRepo.GetUserByIdOrEmail(ctx, models.GetUserByIdOrEmailRequest{
		Email: utils.StrPtr(req.Email),
	})
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

	// if password version is 0, it means password not set,
	// we can return specific error to client to ask them to set password
	if cred.PasswordVersion == 0 {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusForbidden).
			SetError(models.ErrorDTO{Code: "password_not_set", Message: "Password not set; please create password"})
	}

	// support bcrypt, sha256 fallback
	if err := cred.ComparePwdHashWithAlgo(ctx, req.Password); err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "invalid_credentials", Message: "Invalid credentials"})
	}

	// Return response with issued tokens and whether password setup is needed
	jwtToken, commonErr := s.tokenService.GenerateJwtToken(ctx, user)
	if commonErr != nil {
		return nil, commonErr
	}

	jwtToken.CheckPasswordSetup(cred.PasswordVersion, cred.PasswordAlgo)

	// insert audit log for user signup action
	s.auditLogService.CreateAuditLog(ctx,
		models.AuditLogRequest{
			CreatorID: user.ID,
			Action:    constants.AuditLogActionLogin,
			IP:        utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),
			Metadata: datatypes.JSONMap{
				"login_response": jwtToken,
			},
		})

	// return response to client
	return jwtToken, nil
}

// RefreshToken implements interfaces.AuthServiceI
func (s *AuthUseCaseImpl) RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (
	*models.RenewTokenResponse, *errHandler.ErrorBuilder) {

	// tracing for refresh token usecase, we want to trace the whole flow of refresh token process, from checking token valid,
	ctx, span := tracing.StartSpanFromContext(ctx, "RefreshToken")
	defer span.End()

	// check refresh token exist and valid, if not valid then return error, only allow refresh when token is valid
	rt, errRT := s.refreshTokenReadRepo.GetByTokenHash(ctx, req.RefreshToken)
	if errRT != nil {
		return nil, errRT
	}

	// check token valid, if not valid then return error, only allow refresh when token is valid
	if rt.UserID != req.UserID {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{
				Code:    "invalid_refresh",
				Message: "Refresh token not owned by user",
			})
	}

	// check token revoked or expired, if revoked or expired then return error
	if rt.RevokedAt != nil {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{
				Code:    "refresh_revoked",
				Message: "Refresh token revoked",
			})
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{
				Code:    "refresh_expired",
				Message: "Refresh token expired",
			})
	}

	// check user exist with id from token, if not exist return error
	userEntity, errUser := s.userReadRepo.GetUserByIdOrEmail(ctx, models.GetUserByIdOrEmailRequest{
		UserID: utils.Uint64Ptr(req.UserID),
	})
	if errUser != nil {
		return nil, errUser
	}

	// check email verified or not, if not verified then return error, only allow refresh when email is verified
	if !userEntity.EmailVerified || userEntity.Status != "verified" {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusForbidden).
			SetError(models.ErrorDTO{Code: "email_not_verified", Message: "Email not verified"})
	}

	jwtToken, commonErr := s.tokenService.RenewJwtToken(ctx, userEntity)
	if commonErr != nil {
		return nil, commonErr
	}

	// insert audit log for user signup action
	s.auditLogService.CreateAuditLog(ctx,
		models.AuditLogRequest{
			CreatorID: req.UserID,
			Action:    constants.AuditLogActionRefresh,
			IP:        utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),
			Metadata: datatypes.JSONMap{
				"renew_token": jwtToken,
			},
		})

	// return response to client
	return jwtToken, nil
}

// Logout implements interfaces.AuthServiceI
// this API will get token jti from context, then block the token in token denylist with expiry same as token expiry,
// so even if the token is not expired, it will be rejected in next request
func (s *AuthUseCaseImpl) Logout(ctx context.Context, claims jwt.MapClaims) *errHandler.ErrorBuilder {

	// tracing for logout usecase, we want to trace the whole flow of logout process, from checking user context,
	ctx, span := tracing.StartSpanFromContext(ctx, "Logout")
	defer span.End()

	// get jti and exp from token claims,
	// then block the token in token denylist with expiry same as token expiry,
	commonErr := s.tokenService.RevokeJwtToken(ctx, claims)
	if commonErr != nil {
		return commonErr
	}

	// insert audit log for user signup action
	s.auditLogService.CreateAuditLog(ctx,
		models.AuditLogRequest{
			CreatorID: utils.ParseUserID(claims["sub"]),
			Action:    constants.AuditLogActionLogout,
			IP:        utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),
			Metadata: datatypes.JSONMap{
				"claims": claims,
			},
		})

	// insert audit log for user logout action
	return nil
}
