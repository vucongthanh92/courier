package models

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type JwtTokenResponse struct {
	AccessToken       string `json:"access_token"`
	ExpiresIn         int64  `json:"expires_in"`
	RefreshToken      string `json:"refresh_token"`
	RefreshExpiresIn  int64  `json:"refresh_expires_in"`
	TokenType         string `json:"token_type"`
	NeedPasswordSetup bool   `json:"need_password_setup,omitempty"`
}

func (r *JwtTokenResponse) CheckPasswordSetup(pwdVersion int16, pwdAlgo string) {
	r.NeedPasswordSetup = pwdVersion == 0 || pwdAlgo == ""
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	UserID       uint64 `json:"user_id" binding:"required"`
}

type RenewTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}
