package models

type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"displayName" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}
