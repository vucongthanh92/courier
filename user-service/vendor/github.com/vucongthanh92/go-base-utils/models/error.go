package models

type ErrorDTO struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Code    string `json:"code"`
}
