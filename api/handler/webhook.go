package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/kermanager/internal/user"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
)

func HandleWebhook(userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Request Body Read Error", http.StatusServiceUnavailable)
			return
		}

		signatureHeader := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, signatureHeader, os.Getenv("STRIPE_WEBHOOK_SECRET"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Webhook signature verification failed: %v", err), http.StatusBadRequest)
			return
		}

		if event.Type == "checkout.session.completed" {
			var session stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				http.Error(w, "Webhook Error", http.StatusBadRequest)
				return
			}

			userIDStr := session.Metadata["user_id"]
			userId, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			log.Printf("user ID: %d\n", userId)

			creditStr := session.Metadata["credit"]
			credit, err := strconv.Atoi(creditStr)
			if err != nil {
				http.Error(w, "Invalid credit", http.StatusBadRequest)
				return
			}

			log.Printf("credit: %d\n", credit)

			err = userService.UpdateCredit(map[string]interface{}{
				"id":     userId,
				"credit": credit,
			})
			log.Printf("Error: %v\n", err)
			if err != nil {
				http.Error(w, "Error updating user credit", http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Printf("Unhandled event type: %s\n", event.Type)
		}

		w.WriteHeader(http.StatusOK)
	}
}
