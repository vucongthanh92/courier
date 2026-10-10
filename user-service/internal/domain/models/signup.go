package models

import (
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	baseUtils "github.com/vucongthanh92/go-base-utils/helper/utils"
)

type SignupRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
}

func (r *SignupRequest) MappingToUserEntity(entity *entities.User) {
	entity.ID, _ = baseUtils.NewSnowflakeID()
	entity.Email = r.Email
	entity.DisplayName = r.DisplayName
	entity.PhoneNumber = r.PhoneNumber
	entity.Status = "pending"
}

func (r *SignupRequest) MappingToEmailVerifyEntity(entity *entities.EmailVerification) {
	entity.Email = r.Email
	entity.TokenHash = baseUtils.RandString(7)
	entity.ExpiresAt = time.Now().Add(24 * time.Hour)
}
