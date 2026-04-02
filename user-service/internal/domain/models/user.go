package models

type GetUserByIdOrEmailRequest struct {
	UserID *uint64 `json:"id"`
	Email  *string `json:"email"`
}
