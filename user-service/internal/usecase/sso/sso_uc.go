package sso

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"

	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	"github.com/vucongthanh92/go-base-utils/helper/transaction"
	baseUtils "github.com/vucongthanh92/go-base-utils/helper/utils"
	utilsModels "github.com/vucongthanh92/go-base-utils/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type SsoUseCase struct {
	cfg              *utilsModels.SSOConfig
	txn              *transaction.ManagerTxn
	userReadRepo     interfaces.UserQueryRepoI
	authService      interfaces.AuthServiceI
	jwtSigner        interfaces.JWTSignerI
	refreshTokenRepo interfaces.RefreshTokenCommandRepoI
	sessionCmd       interfaces.SsoSessionCommandRepoI
	sessionQuery     interfaces.SsoSessionQueryRepoI
	codeCmd          interfaces.SsoAuthorizationCodeCommandRepoI
	codeQuery        interfaces.SsoAuthorizationCodeQueryRepoI
}

func InitSsoUseCase(
	cfg *config.AppConfig,
	txn *transaction.ManagerTxn,
	userReadRepo interfaces.UserQueryRepoI,
	authService interfaces.AuthServiceI,
	jwtSigner interfaces.JWTSignerI,
	refreshTokenRepo interfaces.RefreshTokenCommandRepoI,
	sessionCmd interfaces.SsoSessionCommandRepoI,
	sessionQuery interfaces.SsoSessionQueryRepoI,
	codeCmd interfaces.SsoAuthorizationCodeCommandRepoI,
	codeQuery interfaces.SsoAuthorizationCodeQueryRepoI,
) interfaces.SsoServiceI {
	return &SsoUseCase{
		cfg:              cfg.SSO,
		txn:              txn,
		userReadRepo:     userReadRepo,
		authService:      authService,
		jwtSigner:        jwtSigner,
		refreshTokenRepo: refreshTokenRepo,
		sessionCmd:       sessionCmd,
		sessionQuery:     sessionQuery,
		codeCmd:          codeCmd,
		codeQuery:        codeQuery,
	}
}

// Authorize handles the SSO authorization request, validating the request parameters,
// checking for an active session, and generating an authorization code if valid.
func (s *SsoUseCase) Authorize(
	ctx context.Context,
	req models.SsoAuthorizeRequest,
	sessionToken string,
) (*models.SsoAuthorizeResponse, *utilsError.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "SsoAuthorize")
	defer span.End()

	// Validate the authorization request parameters and retrieve the client configuration
	client, commonErr := s.validateAuthorizeRequest(ctx, req)
	if commonErr != nil {
		return nil, commonErr
	}

	// Check for an active SSO session using the provided session token
	if sessionToken == "" {
		if req.Prompt == "none" {
			return &models.SsoAuthorizeResponse{RedirectURI: s.oauthErrorRedirect(req.RedirectURI, req.State, "login_required")}, nil
		}

		return &models.SsoAuthorizeResponse{RedirectURI: baseUtils.LoginRedirect(
			s.cfg.LoginURL,
			req.ClientID,
			req.RedirectURI,
			req.ResponseType,
			req.Scope,
			req.State,
			req.Nonce,
			req.CodeChallenge,
			req.CodeChallengeMethod,
		)}, nil
	}
	session, commonErr := s.sessionQuery.GetActiveByHash(ctx, baseUtils.HashSecret(sessionToken), time.Now())
	if commonErr != nil {
		if req.Prompt == "none" {
			return &models.SsoAuthorizeResponse{RedirectURI: s.oauthErrorRedirect(req.RedirectURI, req.State, "login_required")}, nil
		}

		return &models.SsoAuthorizeResponse{RedirectURI: baseUtils.LoginRedirect(
			s.cfg.LoginURL,
			req.ClientID,
			req.RedirectURI,
			req.ResponseType,
			req.Scope,
			req.State,
			req.Nonce,
			req.CodeChallenge,
			req.CodeChallengeMethod,
		)}, nil
	}

	// Generate an authorization code for the active session and client
	code, commonErr := s.createAuthorizationCode(ctx, session, client.ClientID, req)
	if commonErr != nil {
		return nil, commonErr
	}

	// Construct the redirect URI with the authorization code and optional state parameter
	redirectURI, _ := url.Parse(req.RedirectURI)
	query := redirectURI.Query()
	query.Set("code", code)
	if req.State != "" {
		query.Set("state", req.State)
	}

	redirectURI.RawQuery = query.Encode()
	return &models.SsoAuthorizeResponse{RedirectURI: redirectURI.String()}, nil
}

// Login handles the SSO login request, authenticating the user and generating a session token.
// It also processes the authorization request parameters to generate an authorization code if valid.
func (s *SsoUseCase) Login(
	ctx context.Context,
	req models.SsoLoginRequest,
) (*models.SsoLoginResponse, string, time.Time, *utilsError.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "SsoLogin")
	defer span.End()

	// Authenticate the user using the provided email and password
	loginRes, commonErr := s.authService.Login(ctx, models.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if commonErr != nil {
		return nil, "", time.Time{}, commonErr
	}

	// Create a new SSO session for the authenticated user
	userID := loginRes.User.ID
	userEntity, commonErr := s.userReadRepo.GetUserByIdOrEmail(ctx, models.GetUserByIdOrEmailRequest{
		UserID: &userID,
	})
	if commonErr != nil {
		return nil, "", time.Time{}, commonErr
	}

	// Create a new session for the authenticated user
	sessionToken := baseUtils.RandString(64)
	sessionID, _ := baseUtils.NewSnowflakeID()
	session := entities.SsoSession{}
	session.Initialize(
		sessionID,
		baseUtils.HashSecret(sessionToken),
		baseUtils.StringPtr(baseUtils.GetUserAgent(ctx)),
		baseUtils.StringPtr(baseUtils.GetClientIP(ctx)),
		baseUtils.CheckSessionTTLMinutes(s.cfg.SessionTTLMinutes),
		userEntity.ID,
	)

	commonErr = s.sessionCmd.CreateSession(ctx, &session)
	if commonErr != nil {
		return nil, "", time.Time{}, commonErr
	}

	// If an authorization query is provided, parse it and reconstruct the SSO authorization request
	query, err := url.ParseQuery(req.AuthorizeQuery)
	if err != nil {
		return nil, "", time.Time{}, badRequest(ctx, "invalid_authorize_query", "Invalid authorize query")
	}

	// Reconstruct the SSO authorization request from the query parameters
	authReq := models.SsoAuthorizeRequest{
		ClientID:            query.Get("client_id"),
		RedirectURI:         query.Get("redirect_uri"),
		ResponseType:        query.Get("response_type"),
		Scope:               query.Get("scope"),
		State:               query.Get("state"),
		Nonce:               query.Get("nonce"),
		CodeChallenge:       query.Get("code_challenge"),
		CodeChallengeMethod: query.Get("code_challenge_method"),
	}

	authRes, commonErr := s.Authorize(ctx, authReq, sessionToken)
	if commonErr != nil {
		return nil, "", time.Time{}, commonErr
	}

	return &models.SsoLoginResponse{RedirectURI: authRes.RedirectURI}, sessionToken, session.ExpiresAt, nil
}

// Token handles the SSO token request, validating the authorization code and generating access, refresh, and ID tokens for the authenticated user.
// It also consumes the authorization code to prevent reuse.
func (s *SsoUseCase) Token(
	ctx context.Context,
	req models.SsoTokenRequest,
) (*models.SsoTokenResponse, *utilsError.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "SsoToken")
	defer span.End()

	// Validate the grant type and ensure it is "authorization_code"
	if req.GrantType != "authorization_code" {
		return nil, badRequest(ctx, "unsupported_grant_type", "Only authorization_code is supported")
	}

	// Validate the client ID and redirect URI, ensuring they match the registered client configuration
	client, ok := s.findClient(req.ClientID)
	if !ok || !baseUtils.Contains(client.RedirectURIs, req.RedirectURI) {
		return nil, badRequest(ctx, "invalid_client", "Invalid client or redirect URI")
	}

	// Validate the authorization code, ensuring it is active, matches the client and redirect URI, and passes PKCE verification
	now := time.Now()
	codeHash := baseUtils.HashSecret(req.Code)
	authCode, commonErr := s.codeQuery.GetActiveByHash(ctx, codeHash, now)
	switch {
	case commonErr != nil:
		return nil, badRequest(ctx, "invalid_code", "Authorization code is invalid or expired")
	case authCode.ClientID != req.ClientID || authCode.RedirectURI != req.RedirectURI:
		return nil, badRequest(ctx, "invalid_code", "Authorization code does not match request")
	case !baseUtils.VerifyPKCE(authCode.CodeChallenge, req.CodeVerifier):
		return nil, badRequest(ctx, "invalid_grant", "PKCE verification failed")
	}

	// Retrieve the user associated with the authorization code
	userEntity, commonErr := s.userReadRepo.GetUserByIdOrEmail(ctx, models.GetUserByIdOrEmailRequest{
		UserID: &authCode.UserID,
	})
	if commonErr != nil {
		return nil, commonErr
	}

	// Consume the authorization code to prevent reuse and generate access, refresh, and ID tokens for the authenticated user
	if commonErr := s.codeCmd.ConsumeCode(ctx, codeHash, now); commonErr != nil {
		return nil, commonErr
	}

	// Generate access token, ID token, and refresh token for the authenticated user
	accessTTL := 30 * time.Minute
	idTokenTTL := time.Duration(baseUtils.CheckIdTokenTTLMinutes(s.cfg.IDTokenTTLMinutes)) * time.Minute
	sessionID := ""
	if authCode.SsoSessionID != nil {
		sessionID = strconv.FormatUint(*authCode.SsoSessionID, 10)
	}

	accessToken, commonErr := s.jwtSigner.SignAccessTokenForClient(*userEntity, now, accessTTL, req.ClientID, sessionID, authCode.Scope)
	if commonErr != nil {
		return nil, commonErr
	}

	// Generate ID token with optional nonce if provided in the authorization code
	nonce := ""
	if authCode.Nonce != nil {
		nonce = *authCode.Nonce
	}
	idToken, commonErr := s.jwtSigner.SignIDToken(*userEntity, now, idTokenTTL, req.ClientID, nonce, now, sessionID)
	if commonErr != nil {
		return nil, commonErr
	}

	// Generate refresh token and store its hash in the database for the authenticated user
	refreshTTL := 90 * 24 * time.Hour
	refreshPlain := baseUtils.RandString(64)
	refreshHash := baseUtils.HashPwdBySha256(userEntity.Email, refreshPlain)
	rt := entities.RefreshToken{
		UserID:    userEntity.ID,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(refreshTTL),
		UserAgent: baseUtils.StringPtr(baseUtils.GetUserAgent(ctx)),
		IP:        baseUtils.StringPtr(baseUtils.GetClientIP(ctx)),
	}
	if commonErr := s.refreshTokenRepo.UpsertByUserAgent(ctx, &rt); commonErr != nil {
		return nil, commonErr
	}

	// Return the generated tokens and user information in the response
	return &models.SsoTokenResponse{
		AccessToken:      accessToken,
		ExpiresIn:        int64(accessTTL.Seconds()),
		RefreshToken:     refreshPlain,
		RefreshExpiresIn: int64(refreshTTL.Seconds()),
		IDToken:          idToken,
		TokenType:        "Bearer",
		Scope:            authCode.Scope,
		User: &models.AuthenticatedUserResponse{
			ID:          userEntity.ID,
			DisplayName: userEntity.DisplayName,
			AvatarURL:   userEntity.AvatarURL,
		},
	}, nil
}

// Session retrieves the current SSO session information for the provided session token, returning the authenticated user details and session expiration time.
// If the session is not active or invalid, it returns an unauthenticated response.
func (s *SsoUseCase) Session(ctx context.Context, sessionToken string) (*models.SsoSessionResponse, *utilsError.ErrorBuilder) {

	// Retrieve the active SSO session associated with the provided session token
	if sessionToken == "" {
		return nil, badRequest(ctx, "missing_sso_session", "Missing SSO session")
	}
	session, commonErr := s.sessionQuery.GetActiveByHash(ctx, baseUtils.HashSecret(sessionToken), time.Now())
	if commonErr != nil {
		return &models.SsoSessionResponse{Authenticated: false}, nil
	}

	// Retrieve the user associated with the active session to include user details in the response
	userEntity, commonErr := s.userReadRepo.GetUserByIdOrEmail(ctx, models.GetUserByIdOrEmailRequest{
		UserID: &session.UserID,
	})
	if commonErr != nil {
		return nil, commonErr
	}

	return &models.SsoSessionResponse{
		Authenticated: true,
		ExpiresAt:     session.ExpiresAt.Format(time.RFC3339),
		User: &models.AuthenticatedUserResponse{
			ID:          userEntity.ID,
			DisplayName: userEntity.DisplayName,
			AvatarURL:   userEntity.AvatarURL,
		},
	}, nil
}

// Logout revokes the active SSO session and associated refresh tokens for the provided session token, effectively logging the user out of the SSO system.
// If the session is not active or invalid, it still returns a successful logout response to avoid exposing session state.
func (s *SsoUseCase) Logout(ctx context.Context, sessionToken string) (*models.SsoLogoutResponse, *utilsError.ErrorBuilder) {

	// Retrieve the active SSO session associated with the provided session token
	if sessionToken == "" {
		return nil, badRequest(ctx, "missing_sso_session", "Missing SSO session")
	}
	session, commonErr := s.sessionQuery.GetActiveByHash(ctx, baseUtils.HashSecret(sessionToken), time.Now())
	if commonErr == nil {
		now := time.Now()
		sessionHash := baseUtils.HashSecret(sessionToken)
		if commonErr := s.sessionCmd.RevokeByHash(ctx, sessionHash, now); commonErr != nil {
			return nil, commonErr
		}
		if commonErr := s.refreshTokenRepo.RevokeByUser(ctx, session.UserID, now); commonErr != nil {
			return nil, commonErr
		}
	}

	return &models.SsoLogoutResponse{Message: "SSO logout successful"}, nil
}

// createAuthorizationCode generates a new authorization code for the given SSO session and client, storing it in the database with an expiration time.
// It returns the generated code or an error if the operation fails.
func (s *SsoUseCase) createAuthorizationCode(
	ctx context.Context,
	session *entities.SsoSession,
	clientID string,
	req models.SsoAuthorizeRequest,
) (string, *utilsError.ErrorBuilder) {
	now := time.Now()
	codeTTL := time.Duration(baseUtils.CodeTTLSeconds(s.cfg.CodeTTLSeconds)) * time.Second
	code := baseUtils.RandString(48)
	codeID, _ := baseUtils.NewSnowflakeID()
	entity := entities.SsoAuthorizationCode{
		ID:                  codeID,
		CodeHash:            baseUtils.HashSecret(code),
		UserID:              session.UserID,
		ClientID:            clientID,
		RedirectURI:         req.RedirectURI,
		Scope:               baseUtils.RemoveWhiteSpace(req.Scope),
		State:               baseUtils.OptionalString(req.State),
		Nonce:               baseUtils.OptionalString(req.Nonce),
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		SsoSessionID:        &session.ID,
		ExpiresAt:           now.Add(codeTTL),
	}
	if commonErr := s.codeCmd.CreateCode(ctx, &entity); commonErr != nil {
		return "", commonErr
	}
	return code, nil
}

func (s *SsoUseCase) validateAuthorizeRequest(
	ctx context.Context,
	req models.SsoAuthorizeRequest,
) (utilsModels.SSOClientConfig, *utilsError.ErrorBuilder) {
	if req.ResponseType != "code" {
		return utilsModels.SSOClientConfig{}, badRequest(ctx, "unsupported_response_type", "Only code response type is supported")
	}
	if req.CodeChallengeMethod != "S256" {
		return utilsModels.SSOClientConfig{}, badRequest(ctx, "invalid_code_challenge_method", "Only S256 PKCE is supported")
	}
	client, ok := s.findClient(req.ClientID)
	if !ok || !baseUtils.Contains(client.RedirectURIs, req.RedirectURI) {
		return utilsModels.SSOClientConfig{}, badRequest(ctx, "invalid_client", "Invalid client or redirect URI")
	}
	return client, nil
}

func (s *SsoUseCase) findClient(clientID string) (utilsModels.SSOClientConfig, bool) {
	if s.cfg == nil {
		return utilsModels.SSOClientConfig{}, false
	}
	for _, client := range s.cfg.Clients {
		if client.ClientID == clientID {
			return client, true
		}
	}
	return utilsModels.SSOClientConfig{}, false
}

func (s *SsoUseCase) oauthErrorRedirect(redirectURI string, state string, code string) string {
	parsedURL, _ := url.Parse(redirectURI)
	query := parsedURL.Query()
	query.Set("error", code)
	if state != "" {
		query.Set("state", state)
	}
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}

func badRequest(ctx context.Context, code string, message string) *utilsError.ErrorBuilder {
	return utilsError.InitErrorBuilder(ctx).
		SetStatus(http.StatusBadRequest).
		SetLogError(errors.New(message)).
		SetError(utilsModels.ErrorDTO{
			Code: code, Message: message})
}
