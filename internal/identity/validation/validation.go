package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
	"sw/internal/identity/domain"
	"sw/internal/logging"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator(validate *validator.Validate) *CustomValidator {
	return &CustomValidator{validator: validate}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err)
		//return err
	}
	return nil
}

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
