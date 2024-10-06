package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

func HandleWebhook(userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, errors.ServerError, http.StatusServiceUnavailable)
			return
		}

		signatureHeader := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, signatureHeader, os.Getenv("STRIPE_WEBHOOK_SECRET"))
		if err != nil {
			http.Error(w, errors.InvalidInput, http.StatusBadRequest)
			return
		}

		if event.Type == "checkout.session.completed" {
			var session stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				http.Error(w, errors.InvalidInput, http.StatusBadRequest)
				return
			}

			creditStr, ok := session.Metadata["credit"]
			if !ok {
				http.Error(w, errors.InvalidInput, http.StatusBadRequest)
				return
			}
			credit, err := strconv.Atoi(creditStr)
			if err != nil {
				http.Error(w, errors.InvalidInput, http.StatusBadRequest)
				return
			}

			userIdStr, ok := session.Metadata["user_id"]
			if !ok {
				http.Error(w, errors.InvalidInput, http.StatusBadRequest)
				return
			}
			userId, err := strconv.Atoi(userIdStr)
			if err != nil {
				http.Error(w, errors.InvalidInput, http.StatusBadRequest)
				return
			}

			err = userService.UpdateCredit(userId, credit)
			if err != nil {
				http.Error(w, errors.ServerError, http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
