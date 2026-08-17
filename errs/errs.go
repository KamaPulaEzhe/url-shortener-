package errs

import (
	"context"
	"errors"
	"net/http"
)

var ErrURLNotFound = errors.New("URL not found")
var ErrCodeNotFound = errors.New("Code not found")
var ErrAlreadyExists = errors.New("URL already exists")

func WriteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		http.Error(w, err.Error(), http.StatusGatewayTimeout)

	case errors.Is(err, context.Canceled):
		http.Error(w, err.Error(), http.StatusGatewayTimeout)

	case errors.Is(err, ErrCodeNotFound), errors.Is(err, ErrURLNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)

	default:
		http.Error(w, err.Error(), http.StatusBadRequest)

	}
}
