package goplayground

import (
	"github.com/go-playground/validator/v10"
	"sw/internal/apierr"
)

func MapValidationError(errors validator.ValidationErrors) apierr.ValidationErrorResponse {
	//errors := err.(validator.ValidationErrors)
	fields := make([]apierr.FieldError, 0)
	for _, err := range errors {
		ns := err.Namespace()
		tag := err.Tag()
		param := err.Param()
		fields = append(fields, apierr.FieldError{Path: ns, Validator: tag, Parameter: param})
	}
	return apierr.ValidationErrorResponse{Code: apierr.ErrValidationFailed, Message: "", Fields: fields}
}
