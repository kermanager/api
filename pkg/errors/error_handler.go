package errors

import (
	"log"
	"net/http"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request) error

func (f ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		log.Printf("ERROR: %v", err)
		if e, ok := err.(CustomError); ok {
			http.Error(w, e.ErrorMessage(), e.StatusCode())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
