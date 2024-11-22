package validation

import (
	"github.com/go-playground/validator/v10"
	"sw/internal/identity/domain"
	"sw/internal/logging"
)

func NewAccountNotExistValidator(
	repository domain.AccountRepository,
	logger logging.Logger,
) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		email := fl.Field().Interface().(string)
		exists, err := repository.Exists(email)
		if err != nil {
			logger.Println(err)
			return false
		}
		return !exists
	}
}

func NewAccountExistsValidator(
	repository domain.AccountRepository,
	logger logging.Logger,
) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		email := fl.Field().Interface().(string)
		exists, err := repository.Exists(email)
		if err != nil {
			logger.Println(err)
			return false
		}
		return exists
	}
}
