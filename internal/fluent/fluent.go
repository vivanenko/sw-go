package fluent

import (
	"github.com/go-playground/validator/v10"
	"net/http"
	"sw/internal/apierr"
	"sw/internal/apierr/goplayground"
	"sw/internal/encoding"
)

type Context[TRequest any] struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	encoder        encoding.Encoder
	decoder        encoding.Decoder
	validator      *validator.Validate
	handler        func(TRequest) error
}

func NewContext[TRequest any](w http.ResponseWriter, r *http.Request) *Context[TRequest] {
	return &Context[TRequest]{responseWriter: w, request: r}
}

func (c *Context[TRequest]) WithEncoder(encoder encoding.Encoder) *Context[TRequest] {
	c.encoder = encoder
	return c
}

func (c *Context[TRequest]) WithDecoder(decoder encoding.Decoder) *Context[TRequest] {
	c.decoder = decoder
	return c
}

func (c *Context[TRequest]) ValidatedBy(validator *validator.Validate) *Context[TRequest] {
	c.validator = validator
	return c
}

func (c *Context[TRequest]) WithHandler(handler func(TRequest) error) *Context[TRequest] {
	c.handler = handler
	return c
}

func (c *Context[TRequest]) Handle() error {
	var request TRequest
	err := c.decoder.Decode(c.request.Body, &request)
	if err != nil {
		c.responseWriter.Header().Set("Content-Type", "application/json")
		c.responseWriter.WriteHeader(http.StatusBadRequest)
		response := apierr.ErrorResponse{Code: apierr.ErrInvalidBody, Message: "Invalid Body"}
		bytes, err := c.encoder.Encode(response)
		if err != nil {
			return err
		}
		_, err = c.responseWriter.Write(bytes)
		if err != nil {
			return err
		}
		return nil
	}

	err = c.validator.Struct(request)
	if err != nil {
		errors, ok := err.(validator.ValidationErrors)
		if ok {
			response := goplayground.MapValidationError(errors)
			c.responseWriter.Header().Set("Content-Type", "application/json")
			c.responseWriter.WriteHeader(http.StatusBadRequest)
			bytes, err := c.encoder.Encode(response)
			if err != nil {
				return err
			}
			_, err = c.responseWriter.Write(bytes)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	return c.handler(request)
}