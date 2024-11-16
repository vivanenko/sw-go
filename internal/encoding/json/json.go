package json

import (
	"encoding/json"
	"io"
)

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(rc io.ReadCloser, dst interface{}) error {
	// todo: implement an improved decoder based on this article https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	return json.NewDecoder(rc).Decode(&dst)
}

type Encoder struct{}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode(src interface{}) ([]byte, error) {
	return json.Marshal(src)
}
