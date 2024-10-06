package errors

import "net/http"

type CustomError struct {
	Err error
}

func (ce CustomError) Error() string {
	return ce.Err.Error()
}

func (ce CustomError) StatusCode() int {
	switch ce.Err.Error() {
	case NotAllowed:
	case NotEnoughCredits:
	case KermesseAlreadyEnded:
	case TombolaAlreadyEnded:
	case NotEnoughStock:
	case IsNotAnActivity:
		return http.StatusForbidden
	case InvalidInput:
	case EmailAlreadyExists:
	case InvalidCredentials:
		return http.StatusBadRequest
	case ServerError:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}
