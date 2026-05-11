package models

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Token string `json:"token" binding:"required"`
}

type ResendVerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailResponse struct {
	Message string `json:"message"`
}

type ResendVerifyEmailResponse struct {
	Message string `json:"message"`
}
