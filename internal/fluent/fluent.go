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
	errorResponses map[error]apierr.ErrorResponse
	encoder        encoding.Encoder
	decoder        encoding.Decoder
	validator      *validator.Validate
	handler        func(TRequest) error
}

func NewContext[TRequest any](w http.ResponseWriter, r *http.Request) *Context[TRequest] {
	return &Context[TRequest]{responseWriter: w, request: r, errorResponses: map[error]apierr.ErrorResponse{}}
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

func (c *Context[TRequest]) OnError(err error, response apierr.ErrorResponse) *Context[TRequest] {
	c.errorResponses[err] = response
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
		response := apierr.ErrorResponse{Code: apierr.ErrInvalidBody, Message: "Invalid Body"}
		return badRequest(c.responseWriter, c.encoder, response)
	}

	err = c.validator.Struct(request)
	if err != nil {
		errors, ok := err.(validator.ValidationErrors)
		if ok {
			response := goplayground.MapValidationError(errors)
			return badRequest(c.responseWriter, c.encoder, response)
		}
		return err
	}

	err = c.handler(request)
	if err != nil {
		response, exist := c.errorResponses[err]
		if exist {
			return badRequest(c.responseWriter, c.encoder, response)
		}
		return err
	}
	return nil
}

// With response
type ContextWithResponse[TRequest any, TResponse any] struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	errorResponses map[error]apierr.ErrorResponse
	encoder        encoding.Encoder
	decoder        encoding.Decoder
	validator      *validator.Validate
	handler        func(TRequest) (TResponse, error)
}

func NewContextWithResponse[TRequest any, TResponse any](w http.ResponseWriter, r *http.Request) *ContextWithResponse[TRequest, TResponse] {
	return &ContextWithResponse[TRequest, TResponse]{
		responseWriter: w,
		request:        r,
		errorResponses: map[error]apierr.ErrorResponse{},
	}
}

func (c *ContextWithResponse[TRequest, TResponse]) WithEncoder(encoder encoding.Encoder) *ContextWithResponse[TRequest, TResponse] {
	c.encoder = encoder
	return c
}

func (c *ContextWithResponse[TRequest, TResponse]) WithDecoder(decoder encoding.Decoder) *ContextWithResponse[TRequest, TResponse] {
	c.decoder = decoder
	return c
}

func (c *ContextWithResponse[TRequest, TResponse]) ValidatedBy(validator *validator.Validate) *ContextWithResponse[TRequest, TResponse] {
	c.validator = validator
	return c
}

func (c *ContextWithResponse[TRequest, TResponse]) OnError(err error, response apierr.ErrorResponse) *ContextWithResponse[TRequest, TResponse] {
	c.errorResponses[err] = response
	return c
}

func (c *ContextWithResponse[TRequest, TResponse]) WithHandler(handler func(TRequest) (TResponse, error)) *ContextWithResponse[TRequest, TResponse] {
	c.handler = handler
	return c
}

func (c *ContextWithResponse[TRequest, TResponse]) Handle() error {
	var request TRequest
	err := c.decoder.Decode(c.request.Body, &request)
	if err != nil {
		response := apierr.ErrorResponse{Code: apierr.ErrInvalidBody, Message: "Invalid Body"}
		return badRequest(c.responseWriter, c.encoder, response)
	}

	err = c.validator.Struct(request)
	if err != nil {
		errors, ok := err.(validator.ValidationErrors)
		if ok {
			response := goplayground.MapValidationError(errors)
			return badRequest(c.responseWriter, c.encoder, response)
		}
		return err
	}

	result, err := c.handler(request)
	if err != nil {
		response, exist := c.errorResponses[err]
		if exist {
			return badRequest(c.responseWriter, c.encoder, response)
		}
		return err
	}
	return writeResponse(c.responseWriter, c.encoder, http.StatusOK, result)
	//encoded, err := c.encoder.Encode(res)
	//if err != nil {
	//	return err
	//}
	//c.responseWriter.Header().Set("Content-Type", "application/json")
	//c.responseWriter.WriteHeader(http.StatusOK)
	//return nil
}

// common
func badRequest(responseWriter http.ResponseWriter, encoder encoding.Encoder, response interface{}) error {
	//responseWriter.Header().Set("Content-Type", "application/json")
	//responseWriter.WriteHeader(http.StatusBadRequest)
	//bytes, err := encoder.Encode(response)
	//if err != nil {
	//	return err
	//}
	//_, err = responseWriter.Write(bytes)
	//if err != nil {
	//	return err
	//}
	//return nil
	return writeResponse(responseWriter, encoder, http.StatusBadRequest, response)
}

func writeResponse(responseWriter http.ResponseWriter, encoder encoding.Encoder, code int, response interface{}) error {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)
	bytes, err := encoder.Encode(response)
	if err != nil {
		return err
	}
	_, err = responseWriter.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
