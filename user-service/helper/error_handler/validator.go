package errorhandler

import (
	"net/http"
	"strings"

	"github.com/vucongthanh92/go-base-project/helper/constants"
	"github.com/vucongthanh92/go-base-project/internal/domain/models"
	"gorm.io/gorm"
)

type ErrorValidator interface {
	ErrRecordNotFound(err error) bool
	ErrInvalidDB(err error) bool
	ErrInvalidTransaction(err error) bool
	ErrStatusConflict(err error) bool
}

type SqlValidator struct {
}

func (v *SqlValidator) ErrRecordNotFound(err error) bool {
	return strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error())
}

func (v *SqlValidator) ErrInvalidDB(err error) bool {
	return strings.Contains(err.Error(), gorm.ErrInvalidDB.Error())
}

func (v *SqlValidator) ErrInvalidTransaction(err error) bool {
	return strings.Contains(err.Error(), gorm.ErrInvalidTransaction.Error())
}

func (v *SqlValidator) ErrStatusConflict(err error) bool {
	return strings.Contains(err.Error(), constants.STATUS_CONFLICT)
}

// implement method handle validate error
func (b *ErrorBuilder) ValidateError(err error) *ErrorBuilder {

	b.Validator = &SqlValidator{}

	switch {
	default:
		{
			b.SetStatus(http.StatusInternalServerError).
				SetLogError(err).
				SetIsSystemError(true).
				SetError(models.ErrorDTO{
					Message: constants.SystemErrorMessage,
					Code:    constants.SYSTEM_ERROR,
				})
		}
	case b.Validator.ErrRecordNotFound(err):
		{
			b.SetStatus(http.StatusNotFound).
				SetLogError(err).
				SetError(models.ErrorDTO{
					Message: constants.RecordNotExistMessage,
					Code:    constants.RECORD_NOT_EXIST,
				})
		}
	case b.Validator.ErrInvalidDB(err):
		{
			b.SetStatus(http.StatusInternalServerError).
				SetLogError(err).
				SetIsSystemError(true).
				SetError(models.ErrorDTO{
					Message: constants.SystemErrorMessage,
					Code:    constants.SYSTEM_ERROR,
				})
		}
	case b.Validator.ErrInvalidTransaction(err):
		{
			b.SetStatus(http.StatusInternalServerError).
				SetLogError(err).
				SetIsSystemError(true).
				SetError(models.ErrorDTO{
					Message: constants.SystemErrorMessage,
					Code:    constants.SYSTEM_ERROR,
				})
		}
	case b.Validator.ErrStatusConflict(err):
		{
			b.SetStatus(http.StatusConflict).
				SetLogError(err).
				SetIsSystemError(true).
				SetError(models.ErrorDTO{
					Message: constants.StatusConflictMessage,
					Code:    constants.STATUS_CONFLICT,
				})
		}
	}

	return b
}
