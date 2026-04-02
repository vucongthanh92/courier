package identity

import (
	"context"
	"net/http"
	"time"

	"github.com/lib/pq"
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

type IdentityServiceImpl struct {
	txn                   *transaction.ManagerTxn
	auditLogService       interfaces.AuditLogServiceI
	outboxService         interfaces.OutboxServiceI
	userReadRepo          interfaces.UserQueryRepoI
	userWriteRepo         interfaces.UserCommandRepoI
	identityWriteRepo     interfaces.IdentityCommandRepoI
	identityReadRepo      interfaces.IdentityQueryRepoI
	googleClient          interfaces.GoogleProviderClient
	githubClient          interfaces.GithubProviderClient
	JwtSigner             interfaces.JWTSignerI
	authCredWriteRepo     interfaces.AuthCredentialCommandRepoI
	authCredReadRepo      interfaces.AuthCredentialQueryRepoI
	refreshTokenWriteRepo interfaces.RefreshTokenCommandRepoI
}

func InitIdentityService(
	txn *transaction.ManagerTxn,
	auditLogService interfaces.AuditLogServiceI,
	outboxService interfaces.OutboxServiceI,
	userReadRepo interfaces.UserQueryRepoI,
	userWriteRepo interfaces.UserCommandRepoI,
	identityWriteRepo interfaces.IdentityCommandRepoI,
	identityReadRepo interfaces.IdentityQueryRepoI,
	googleClient interfaces.GoogleProviderClient,
	githubClient interfaces.GithubProviderClient,
	JwtSigner interfaces.JWTSignerI,
	authCredWriteRepo interfaces.AuthCredentialCommandRepoI,
	authCredReadRepo interfaces.AuthCredentialQueryRepoI,
	refreshTokenWriteRepo interfaces.RefreshTokenCommandRepoI,
) interfaces.IdentityServiceI {
	return &IdentityServiceImpl{
		txn:                   txn,
		auditLogService:       auditLogService,
		outboxService:         outboxService,
		userReadRepo:          userReadRepo,
		userWriteRepo:         userWriteRepo,
		identityWriteRepo:     identityWriteRepo,
		identityReadRepo:      identityReadRepo,
		googleClient:          googleClient,
		githubClient:          githubClient,
		JwtSigner:             JwtSigner,
		authCredWriteRepo:     authCredWriteRepo,
		authCredReadRepo:      authCredReadRepo,
		refreshTokenWriteRepo: refreshTokenWriteRepo,
	}
}

// OAuthLogin implements interfaces.Oauth3rdUseCaseI
func (s *IdentityServiceImpl) OAuthLogin(ctx context.Context, req models.OAuthLoginRequest) (
	*models.OAuthLoginResponse, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "OAuthLogin")
	defer span.End()

	// Validate provider and token, get user profile from provider
	var client interfaces.ProviderClient

	// For better security, we should verify the token with the provider's API to ensure it's valid and not tampered with,
	switch req.Provider {
	case constants.GoogleProvider:
		client = s.googleClient
	case constants.GithubProvider:
		client = s.githubClient
	default:
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{Code: "unsupported_provider", Message: "Provider not supported"})
	}

	// Verify token and get user profile from provider
	profile, err := client.Verify(ctx, req.Token)
	if err != nil || profile.Email == "" || !profile.EmailVerified {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "oauth_invalid_token", Message: "Invalid or unverified token"})
	}

	// Log the OAuth login attempt (without sensitive token info)
	var userEntity *entities.User
	identity, logErr := s.identityReadRepo.GetByProviderUID(ctx, profile.Provider, profile.ProviderUID)
	if logErr != nil {
		errHandler.InitErrorBuilder(ctx).SetLogError(logErr.LogError).ExposeLogError()
	}

	// If identity exists, we can get the user ID from it. If not, we will create a new user and identity record in the transaction below.
	var getUserReq = models.GetUserByIdOrEmailRequest{
		Email: utils.StrPtr(profile.Email),
	}
	if identity != nil {
		getUserReq.UserID = utils.Uint64Ptr(identity.UserID)
	}
	userEntity, logErr = s.userReadRepo.GetUserByIdOrEmail(ctx, getUserReq)
	if logErr != nil {
		errHandler.InitErrorBuilder(ctx).SetLogError(logErr.LogError).ExposeLogError()
	}

	// Transaction: create/link user + identity + auth_credential when needed
	errTxn := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		if userEntity == nil {
			userEntity = &entities.User{
				Email:         profile.Email,
				DisplayName:   profile.Name,
				AvatarURL:     profile.AvatarURL,
				EmailVerified: true,
				Status:        "verified",
			}
			userEntity.ID, _ = utils.NewSnowflakeID()
			if txnErr := s.userWriteRepo.InsertUser(txCtx, userEntity); txnErr != nil {
				return txnErr
			}

			// auth credential with password unset
			authCred := entities.AuthCredential{
				UserID:          userEntity.ID,
				PasswordHash:    "",
				PasswordAlgo:    "",
				MFAEnabled:      false,
				PasswordVersion: 0,
			}
			if txnErr := s.authCredWriteRepo.InsertAuthCredential(txCtx, &authCred); txnErr != nil {
				return txnErr
			}
		}

		// If identity record doesn't exist, create one to link the user with the provider's profile
		if identity == nil {
			identity = &entities.Identity{
				UserID:         userEntity.ID,
				Provider:       profile.Provider,
				ProviderUID:    profile.ProviderUID,
				EmailAtAuth:    utils.StrPtr(profile.Email),
				Scopes:         pq.StringArray{"read:user", "user:email"},
				AccessTokenEnc: []byte(req.Token),
			}

			txnErr := s.identityWriteRepo.InserIdentity(txCtx, identity)
			return txnErr
		}

		return nil
	})

	// If transaction failed, return error
	if errTxn != nil {
		commonErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, commonErr
	}

	// Issue tokens
	now := time.Now()
	accessTTL := 30 * time.Minute
	refreshTTL := 90 * 24 * time.Hour
	accessToken, signErr := s.JwtSigner.SignAccessToken(*userEntity, now, accessTTL)
	if signErr != nil {
		return nil, signErr
	}

	// Store refresh token hash in DB for later verification.
	// We will use the same random string as the raw refresh token to return to client,
	// and only store the hash in DB for better security.
	refreshPlain := utils.RandString(64)
	refreshHash := utils.HashPwdBySha256(userEntity.Email, refreshPlain)
	rt := entities.RefreshToken{
		UserID:    userEntity.ID,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(refreshTTL),
		UserAgent: utils.StrPtr(utils.GetUserAgent(ctx)),
		IP:        utils.StrPtr(utils.GetClientIP(ctx)),
	}

	if err := s.refreshTokenWriteRepo.UpsertByUserAgent(ctx, &rt); err != nil {
		return nil, err
	}

	// Check if user needs to set up password
	// (for better security, in case they want to use non-3rd-party login in the future)
	cred, _ := s.authCredReadRepo.GetByUserID(ctx, userEntity.ID)
	needPwd := cred.PasswordVersion == 0 || cred.PasswordHash == ""

	// insert audit log for user signup action
	auditLogReq := models.AuditLogRequest{
		CreatorID: userEntity.ID,
		Action:    req.Provider + "_oauth",
		IP:        utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),
		Metadata: datatypes.JSONMap{
			"identities": identity,
		},
	}
	logErr = s.auditLogService.CreateAuditLog(ctx, auditLogReq)
	if logErr != nil {
		errHandler.InitErrorBuilder(ctx).SetLogError(logErr.LogError).ExposeLogError()
	}

	// Return tokens
	return &models.OAuthLoginResponse{
		AccessToken:       accessToken,
		ExpiresIn:         int64(accessTTL.Seconds()),
		RefreshToken:      refreshPlain,
		RefreshExpiresIn:  int64(refreshTTL.Seconds()),
		TokenType:         "Bearer",
		NeedPasswordSetup: needPwd,
	}, nil
}

// OAuthCallback implements interfaces.IdentityServiceI
func (s *IdentityServiceImpl) OAuthCallback(ctx context.Context, req models.OAuthCallbackRequest) (
	*models.OAuthLoginResponse, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "OAuthCallback")
	defer span.End()

	// For security reasons, we should only support code exchange for providers that require it (like GitHub).
	if req.Provider != "github" {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    "unsupported_provider",
				Message: "Callback not supported for provider",
			})
	}

	// Exchange the authorization code for an access token
	accessToken, err := s.githubClient.(interfaces.GithubCodeExchanger).ExchangeCode(ctx, req.Code, req.RedirectURI)
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusUnauthorized).
			SetError(models.ErrorDTO{Code: "oauth_code_exchange_failed", Message: err.Error()})
	}

	// With the access token, we can now call the same logic as OAuthLogin to verify the token,
	// get the user profile, and issue our own tokens.
	return s.OAuthLogin(ctx, models.OAuthLoginRequest{
		Token:    accessToken,
		Provider: req.Provider,
	})
}
