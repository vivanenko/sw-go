package httpext

import (
	"log"
	"net/http"
	"sw/internal/encoding"
)

const (
	InternalServerError = "Internal Server Error"
)

func BadRequest(w http.ResponseWriter, encoder encoding.Encoder, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	bytes, err := encoder.Encode(err)
	if err != nil {
		log.Println(err)
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Println(err)
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
}
