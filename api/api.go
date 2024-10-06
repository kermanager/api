package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/api/handler"
	"github.com/kermanager/internal/interaction"
	"github.com/kermanager/internal/kermesse"
	"github.com/kermanager/internal/stand"
	"github.com/kermanager/internal/ticket"
	"github.com/kermanager/internal/tombola"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/third_party/resend"
	"github.com/rs/cors"
)

type APIServer struct {
	address string
	db      *sqlx.DB
}

func NewAPIServer(address string, db *sqlx.DB) *APIServer {
	return &APIServer{
		address: address,
		db:      db,
	}
}

func (s *APIServer) Start() error {
	router := mux.NewRouter()

	resendService := resend.NewResendService(os.Getenv("RESEND_API_KEY"), os.Getenv("RESEND_FROM_EMAIL"))

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	userStore := user.NewStore(s.db)
	userService := user.NewService(userStore, resendService)
	userHandler := handler.NewUserHandler(userService, userStore)
	userHandler.RegisterRoutes(router)

	standStore := stand.NewStore(s.db)
	standService := stand.NewService(standStore)
	standHandler := handler.NewStandHandler(standService, userStore)
	standHandler.RegisterRoutes(router)

	kermesseStore := kermesse.NewStore(s.db)
	kermesseService := kermesse.NewService(kermesseStore, userStore)
	kermesseHandler := handler.NewKermesseHandler(kermesseService, userStore)
	kermesseHandler.RegisterRoutes(router)

	interactionStore := interaction.NewStore(s.db)
	interactionService := interaction.NewService(interactionStore, standStore, userStore, kermesseStore)
	interactionHandler := handler.NewInteractionHandler(interactionService, userStore)
	interactionHandler.RegisterRoutes(router)

	tombolaStore := tombola.NewStore(s.db)
	tombolaService := tombola.NewService(tombolaStore, kermesseStore)
	tombolaHandler := handler.NewTombolaHandler(tombolaService, userStore)
	tombolaHandler.RegisterRoutes(router)

	ticketStore := ticket.NewStore(s.db)
	ticketService := ticket.NewService(ticketStore, tombolaStore, userStore)
	ticketHandler := handler.NewTicketHandler(ticketService, userStore)
	ticketHandler.RegisterRoutes(router)

	router.HandleFunc("/webhook", handler.HandleWebhook(userService)).Methods(http.MethodPost)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	r := c.Handler(router)

	log.Printf("Starting server on %s", s.address)
	return http.ListenAndServe(s.address, r)
}
