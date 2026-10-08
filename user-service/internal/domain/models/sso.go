package models

type SsoAuthorizeRequest struct {
	ClientID            string `form:"client_id" binding:"required"`
	RedirectURI         string `form:"redirect_uri" binding:"required"`
	ResponseType        string `form:"response_type" binding:"required"`
	Scope               string `form:"scope"`
	State               string `form:"state"`
	Nonce               string `form:"nonce"`
	CodeChallenge       string `form:"code_challenge" binding:"required"`
	CodeChallengeMethod string `form:"code_challenge_method" binding:"required"`
	Prompt              string `form:"prompt"`
}

type SsoAuthorizeResponse struct {
	RedirectURI string `json:"redirect_uri"`
}

type SsoLoginRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	AuthorizeQuery string `json:"authorize_query" binding:"required"`
}

type SsoLoginResponse struct {
	RedirectURI string `json:"redirect_uri"`
}

type SsoTokenRequest struct {
	GrantType    string `json:"grant_type" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	Code         string `json:"code" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"`
}

type SsoTokenResponse struct {
	AccessToken      string                     `json:"access_token"`
	ExpiresIn        int64                      `json:"expires_in"`
	RefreshToken     string                     `json:"refresh_token"`
	RefreshExpiresIn int64                      `json:"refresh_expires_in"`
	IDToken          string                     `json:"id_token"`
	TokenType        string                     `json:"token_type"`
	Scope            string                     `json:"scope"`
	User             *AuthenticatedUserResponse `json:"user,omitempty"`
}

type SsoSessionResponse struct {
	Authenticated bool                       `json:"authenticated"`
	User          *AuthenticatedUserResponse `json:"user,omitempty"`
	ExpiresAt     string                     `json:"expires_at,omitempty"`
}

type SsoLogoutResponse struct {
	Message string `json:"message"`
}
