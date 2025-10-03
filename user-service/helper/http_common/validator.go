package httpcommon

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vucongthanh92/go-base-project/helper/constants"
	"github.com/vucongthanh92/go-base-project/helper/utils"
	"github.com/vucongthanh92/go-base-project/internal/domain/models"
)

type Validator struct {
	validate *validator.Validate
}

func ValidatorParams(req any) []models.ErrorDTO {
	var (
		validationErrors validator.ValidationErrors
		resErr           = make([]models.ErrorDTO, 0)
	)

	v := Validator{
		validate: validator.New(),
	}

	err := v.validate.Struct(req)
	switch {
	case err == nil:
		return nil
	case !errors.As(err, &validationErrors):
		resErr = append(resErr, models.ErrorDTO{
			Message: "Invalid Request",
			Code:    constants.INVALID_FORMAT,
			Field:   "",
		})
	default:
		for _, val := range err.(validator.ValidationErrors) {
			msgErr := "Invalid Request"
			fieldErr := utils.LowerInitial(strings.Split(val.StructNamespace(), ".")[1:])
			resErr = append(resErr, models.ErrorDTO{
				Message: msgErr,
				Code:    constants.INVALID_FORMAT,
				Field:   strings.Join(fieldErr, "."),
			})
		}
	}

	return resErr
}
