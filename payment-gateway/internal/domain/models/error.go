package models

// ErrorDTO is the shared API error payload used by the copied helper packages.
type ErrorDTO struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Code    string `json:"code"`
}
