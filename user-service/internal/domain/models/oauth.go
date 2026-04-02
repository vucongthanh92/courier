package models

type OAuthLoginRequest struct {
	Token       string `json:"token" binding:"required"` // Google id_token or GitHub access_token
	RedirectURI string `json:"redirect_uri,omitempty"`   // for future PKCE/exchange
	Provider    string `json:"-"`                        // set from path param
}

type OAuthLoginResponse struct {
	AccessToken       string `json:"access_token"`
	ExpiresIn         int64  `json:"expires_in"`
	RefreshToken      string `json:"refresh_token"`
	RefreshExpiresIn  int64  `json:"refresh_expires_in"`
	TokenType         string `json:"token_type"`
	NeedPasswordSetup bool   `json:"need_password_setup,omitempty"`
}

type ProviderProfile struct {
	Provider      string
	ProviderUID   string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
}

type OAuthCallbackRequest struct {
	Code        string `json:"code" form:"code" binding:"required"`
	State       string `json:"state" form:"state"`
	RedirectURI string `json:"redirect_uri,omitempty" form:"redirect_uri"`
	Provider    string `json:"-"`
}
