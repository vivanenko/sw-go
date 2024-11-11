package validation

import (
	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(i interface{}) error
}

type Error struct {
	InvalidJson bool         `json:"invalid_json"`
	Fields      []FieldError `json:"fields"`
}

func (e Error) Error() string {
	return "validation failed"
}

type FieldError struct {
	Path      string `json:"path"`
	Validator string `json:"validator"`
	Parameter string `json:"parameter"`
}

type DefaultValidator struct {
	validator *validator.Validate
}

func NewDefaultValidator(validator *validator.Validate) *DefaultValidator {
	return &DefaultValidator{validator: validator}
}

func (r DefaultValidator) Validate(i interface{}) error {
	err := r.validator.Struct(i)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		fields := make([]FieldError, 0)
		for _, err := range errors {
			ns := err.Namespace()
			tag := err.Tag()
			param := err.Param()
			fields = append(fields, FieldError{Path: ns, Validator: tag, Parameter: param})
		}
		return Error{Fields: fields}
	}
	return nil
}
