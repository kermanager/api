package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kermanager/api/middleware"
	"github.com/kermanager/internal/ticket"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/json"
)

type TicketHandler struct {
	service   ticket.TicketService
	userStore user.UserStore
}

func NewTicketHandler(service ticket.TicketService, userStore user.UserStore) *TicketHandler {
	return &TicketHandler{
		service:   service,
		userStore: userStore,
	}
}

func (h *TicketHandler) RegisterRoutes(mux *mux.Router) {
	mux.Handle("/tickets", errors.ErrorHandler(middleware.IsAuth(h.GetAll, h.userStore))).Methods(http.MethodGet)
	mux.Handle("/tickets/{id}", errors.ErrorHandler(middleware.IsAuth(h.Get, h.userStore))).Methods(http.MethodGet)
	mux.Handle("/tickets", errors.ErrorHandler(middleware.IsAuth(h.Create, h.userStore, types.UserRoleChild))).Methods(http.MethodPost)
}

func (h *TicketHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	tickets, err := h.service.GetAll(r.Context())
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, tickets); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	ticket, err := h.service.Get(r.Context(), id)
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, ticket); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if err := h.service.Create(r.Context(), input); err != nil {
		return err
	}

	if err := json.Write(w, http.StatusCreated, nil); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
