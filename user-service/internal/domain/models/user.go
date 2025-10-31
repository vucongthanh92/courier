package models

import (
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type SignupRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
}

func (r *SignupRequest) MappingToUserEntity(entity *entities.User) {
	entity.ID, _ = utils.NewSnowflakeID()
	entity.Email = r.Email
	entity.DisplayName = r.DisplayName
	entity.PhoneNumber = r.PhoneNumber
	entity.Status = "pending"
}

func (r *SignupRequest) MappingToEmailVerifyEntity(entity *entities.EmailVerification) {
	entity.Email = r.Email
}

func (r *SignupRequest) MappingToAuthCredEntity(entity *entities.AuthCredential) {
	entity.PasswordAlgo = "bcrypt"

	hash, err := utils.HashPwdByBcrypt(r.Password)
	if err != nil {
		hash = utils.HashPwdBySha256(r.Email, r.Password)
		entity.PasswordAlgo = "sha256"
	}

	entity.PasswordHash = hash
	entity.MFAEnabled = false
	entity.PasswordVersion = 1
}
