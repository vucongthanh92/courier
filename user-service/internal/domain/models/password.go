package models

type GeneratePasswordRequest struct {
	Password string `json:"password" binding:"required"`
	UserID   uint64 `json:"user_id"`
}
