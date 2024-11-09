package encoding

import "io"

type Decoder interface {
	Decode(rc io.ReadCloser, dst interface{}) error
}

type Encoder interface {
	Encode(src interface{}) ([]byte, error)
}
