package json

import (
	"encoding/json"
	"io"
	"sw/internal/validation"
)

type Decoder struct {
	validator validation.Validator
}

func NewDecoder(validator validation.Validator) *Decoder {
	return &Decoder{validator: validator}
}

func (d *Decoder) Decode(rc io.ReadCloser, dst interface{}) error {
	// todo: implement an improved decoder based on this article https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	err := json.NewDecoder(rc).Decode(&dst)
	if err != nil {
		return validation.Error{InvalidJson: true, Fields: make([]validation.FieldError, 0)}
	}
	return d.validator.Validate(dst)
}

type Encoder struct{}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode(src interface{}) ([]byte, error) {
	return json.Marshal(src)
}
