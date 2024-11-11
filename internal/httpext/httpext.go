package httpext

import (
	"errors"
	"net/http"
	"sw/internal/encoding"
	"sw/internal/logging"
)

type Wrapper struct {
	w       http.ResponseWriter
	r       *http.Request
	logger  logging.Logger
	decoder encoding.Decoder
	encoder encoding.Encoder
}

func NewWrapper(
	w http.ResponseWriter,
	r *http.Request,
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
) *Wrapper {
	return &Wrapper{w: w, r: r, logger: logger, decoder: decoder, encoder: encoder}
}

func (w *Wrapper) Bind(model interface{}) error {
	err := w.decoder.Decode(w.r.Body, model)
	if err != nil {
		w.BadRequest(err)
		return errors.New("model binding was failed")
	}
	return nil
}

func (w *Wrapper) BadRequest(err error) {
	w.w.Header().Set("Content-Type", "application/json")
	w.w.WriteHeader(http.StatusBadRequest)
	bytes, err := w.encoder.Encode(err)
	if err != nil {
		w.InternalServerError(err)
		return
	}
	_, err = w.w.Write(bytes)
	if err != nil {
		w.InternalServerError(err)
		return
	}
}

func (w *Wrapper) InternalServerError(err error) {
	w.logger.Println(err)
	http.Error(w.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
