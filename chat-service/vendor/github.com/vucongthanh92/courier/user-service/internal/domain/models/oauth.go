package models

type OAuthLoginRequest struct {
	Token       string `json:"token" binding:"required"` // Google id_token or GitHub access_token
	RedirectURI string `json:"redirect_uri,omitempty"`   // for future PKCE/exchange
	Provider    string `json:"-"`                        // set from path param
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
