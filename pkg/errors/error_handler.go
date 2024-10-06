package errors

import (
	"net/http"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request) error

func (f ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		if e, ok := err.(CustomError); ok {
			http.Error(w, e.Error(), e.StatusCode())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
