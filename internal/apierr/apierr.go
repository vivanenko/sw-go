package apierr

const (
	ErrInvalidBody      = "ERR_INVALID_BODY"
	ErrValidationFailed = "ERR_VALIDATION_FAILED"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields"`
}

type FieldError struct {
	Path      string `json:"path"`
	Validator string `json:"validator"`
	Parameter string `json:"parameter"`
}
