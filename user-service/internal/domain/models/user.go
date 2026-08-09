package models

type GetUserByIdOrEmailRequest struct {
	UserID *uint64 `json:"id"`
	Email  *string `json:"email"`
}

type SearchUsersRequest struct {
	SearchKey     string `form:"search_key" json:"search_key" binding:"required"`
	ExcludeUserID uint64 `json:"exclude_user_id"`
	Limit         int    `form:"limit" json:"limit"`
}

type SearchUserResponse struct {
	UserID      uint64 `json:"user_id,string"`
	DisplayName string `json:"display_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar,omitempty"`
}
