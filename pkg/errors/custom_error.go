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
	case TombolaNotEnded:
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

func (ce CustomError) ErrorMessage() string {
	switch ce.Err.Error() {
	case NotAllowed:
		return "Vous n'êtes pas autorisé à effectuer cette action."
	case EmailAlreadyExists:
		return "Cette adresse e-mail est déjà utilisée. Veuillez en essayer une autre."
	case InvalidInput:
		return "Certaines informations saisies sont incorrectes. Veuillez vérifier les champs et réessayer."
	case InvalidCredentials:
		return "Les identifiants sont incorrects. Veuillez vérifier vos informations de connexion."
	case NotEnoughCredits:
		return "Vous n'avez pas assez de crédits pour effectuer cette action."
	case KermesseAlreadyEnded:
		return "La kermesse est déjà terminée."
	case TombolaAlreadyEnded:
		return "La tombola est déjà terminée."
	case NotEnoughStock:
		return "Stock insuffisant pour finaliser l'opération."
	case IsNotAnActivity:
		return "L'activité sélectionnée n'existe pas. Veuillez en choisir une autre."
	case TombolaNotEnded:
		return "La kermesse a une tombola en cours. Veuillez la terminer avant de continuer."
	case ServerError:
	default:
		return "Une erreur serveur est survenue. Veuillez réessayer plus tard."
	}

	return "Une erreur serveur est survenue. Veuillez réessayer plus tard."
}
