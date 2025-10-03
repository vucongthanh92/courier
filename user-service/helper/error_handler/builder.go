package errorhandler

import (
	"context"

	"github.com/vucongthanh92/go-base-project/internal/domain/models"
)

type ErrorBuilder struct {
	ctx             context.Context
	IsSystemError   bool              `json:"is_system_error"`
	IsMultipleError bool              `json:"is_multiple_error"`
	LogError        error             `json:"log_error"`
	Status          int               `json:"status"`
	Errors          []models.ErrorDTO `json:"errors"`
	Validator       ErrorValidator
}

func InitErrorBuilder(ctx context.Context) *ErrorBuilder {
	return &ErrorBuilder{
		ctx:             ctx,
		IsSystemError:   false,
		IsMultipleError: false,
		Errors:          []models.ErrorDTO{},
	}
}

func (b *ErrorBuilder) SetIsSystemError(req bool) *ErrorBuilder {
	b.IsSystemError = req
	return b
}

func (b *ErrorBuilder) SetIsMultipleError(req bool) *ErrorBuilder {
	b.IsMultipleError = req
	return b
}

func (b *ErrorBuilder) SetLogError(req error) *ErrorBuilder {
	b.LogError = req
	return b
}

func (b *ErrorBuilder) SetError(req models.ErrorDTO) *ErrorBuilder {
	b.Errors = append(b.Errors, req)
	return b
}

func (b *ErrorBuilder) SetArrayError(req []models.ErrorDTO) *ErrorBuilder {
	b.Errors = append(b.Errors, req...)
	return b
}

func (b *ErrorBuilder) SetStatus(req int) *ErrorBuilder {
	b.Status = req
	return b
}
