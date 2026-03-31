package auth

import (
	"context"
	"net/http"
	"time"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type Oauth3rdUseCaseImpl struct {
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

func InitOauth3rdUsecase(
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
) interfaces.Oauth3rdUseCaseI {
	return &Oauth3rdUseCaseImpl{
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
func (s *Oauth3rdUseCaseImpl) OAuthLogin(ctx context.Context, req models.OAuthLoginRequest) (
	*models.OAuthLoginResponse, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "OAuthLogin")
	defer span.End()

	// Validate provider and token, get user profile from provider
	var client interface {
		Verify(ctx context.Context, token string) (models.ProviderProfile, error)
	}

	// For better security, we should verify the token with the provider's API to ensure it's valid and not tampered with,
	switch req.Provider {
	case "google":
		client = s.googleClient
	case "github":
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

	// Try identity first
	identity, idErr := s.identityReadRepo.GetByProviderUID(ctx, profile.Provider, profile.ProviderUID)
	var user entities.User
	if idErr == nil && identity.ID > 0 {
		user, _ = s.userReadRepo.GetUserByID(ctx, identity.UserID)
	} else {
		// fallback: lookup by email
		u, uErr := s.userReadRepo.GetUserByEmail(ctx, profile.Email)
		if uErr == nil && u.ID > 0 {
			user = u
		}
	}

	// Transaction: create/link user + identity + auth_credential when needed
	errTxn := s.txn.Do(ctx, func(txCtx context.Context) *errHandler.ErrorBuilder {
		if user.ID == 0 {
			user = entities.User{
				Email:         profile.Email,
				DisplayName:   profile.Name,
				AvatarURL:     profile.AvatarURL,
				EmailVerified: true,
				Status:        "verified",
			}
			if txnErr := s.userWriteRepo.InsertUser(txCtx, &user); txnErr != nil {
				return txnErr
			}

			// auth credential with password unset
			authCred := entities.AuthCredential{
				UserID:          user.ID,
				PasswordHash:    "",
				PasswordAlgo:    "",
				MFAEnabled:      false,
				PasswordVersion: 0,
			}
			if txnErr := s.authCredWriteRepo.InsertAuthCredential(txCtx, &authCred); txnErr != nil {
				return txnErr
			}
		}

		identityEntity := entities.Identity{
			UserID:      user.ID,
			Provider:    profile.Provider,
			ProviderUID: profile.ProviderUID,
			EmailAtAuth: utils.StrPtr(profile.Email),
		}
		_, txnErr := s.identityWriteRepo.UpsertIdentity(txCtx, identityEntity)
		return txnErr
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
	accessToken, signErr := s.JwtSigner.SignAccessToken(user, now, accessTTL)
	if signErr != nil {
		return nil, signErr
	}
	refreshPlain := utils.RandString(64)
	refreshHash := utils.HashPwdBySha256(user.Email, refreshPlain)
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

	// Check if user needs to set up password
	// (for better security, in case they want to use non-3rd-party login in the future)
	cred, _ := s.authCredReadRepo.GetByUserID(ctx, user.ID)
	needPwd := cred.PasswordVersion == 0 || cred.PasswordHash == ""

	// Return tokens
	return &models.OAuthLoginResponse{
		AccessToken:       accessToken,
		ExpiresIn:         int64(accessTTL.Seconds()),
		RefreshToken:      refreshPlain,
		RefreshExpiresIn:  int64(refreshTTL.Seconds()),
		TokenType:         "Bearer",
		NeedPasswordSetup: needPwd, // if user has no password set, prompt them to set one for better security
	}, nil
}
