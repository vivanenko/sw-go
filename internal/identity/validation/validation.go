package validation

import (
	v "github.com/go-playground/validator/v10"
	"log"
	"sw/internal/identity/domain"
)

func NewAccountExistsValidator(repository domain.AccountRepository) func(fl v.FieldLevel) bool {
	return func(fl v.FieldLevel) bool {
		email := fl.Field().Interface().(string)
		exists, err := repository.Exists(email)
		if err != nil {
			log.Print(err)
			return false
		}
		return !exists
	}
}
