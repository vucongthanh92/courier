package models

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"` // seconds
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"` // "Bearer"
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	RefreshExpiresIn int64  `json:"refresh_expires_in,omitempty"`
	TokenType        string `json:"token_type"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}
