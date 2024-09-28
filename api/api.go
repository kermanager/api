package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/api/handler"
	"github.com/kermanager/internal/interaction"
	"github.com/kermanager/internal/kermesse"
	"github.com/kermanager/internal/stand"
	"github.com/kermanager/internal/user"
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

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	userStore := user.NewStore(s.db)
	userService := user.NewService(userStore)
	userHandler := handler.NewUserHandler(userService, userStore)
	userHandler.RegisterRoutes(router)

	standStore := stand.NewStore(s.db)
	standService := stand.NewService(standStore)
	standHandler := handler.NewStandHandler(standService, userStore)
	standHandler.RegisterRoutes(router)

	kermesseStore := kermesse.NewStore(s.db)
	kermesseService := kermesse.NewService(kermesseStore)
	kermesseHandler := handler.NewKermesseHandler(kermesseService, userStore)
	kermesseHandler.RegisterRoutes(router)

	interactionStore := interaction.NewStore(s.db)
	interactionService := interaction.NewService(interactionStore)
	interactionHandler := handler.NewInteractionHandler(interactionService, userStore)
	interactionHandler.RegisterRoutes(router)

	log.Printf("🚀 Starting server on %s", s.address)
	return http.ListenAndServe(s.address, router)
}
