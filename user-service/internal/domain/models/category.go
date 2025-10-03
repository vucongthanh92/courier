package models

import "time"

type CreateCategoryReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Status      bool   `json:"status"`
}

type UpdateCategoryReq struct {
	ID          uint64    `json:"id" validate:"required"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Status      bool      `json:"status"`
	UpdatedAt   time.Time `json:"updated_at" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
}
