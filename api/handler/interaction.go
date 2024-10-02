package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kermanager/api/middleware"
	"github.com/kermanager/internal/interaction"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/json"
	"github.com/kermanager/pkg/utils"
)

type InteractionHandler struct {
	service   interaction.InteractionService
	userStore user.UserStore
}

func NewInteractionHandler(service interaction.InteractionService, userStore user.UserStore) *InteractionHandler {
	return &InteractionHandler{
		service:   service,
		userStore: userStore,
	}
}

func (h *InteractionHandler) RegisterRoutes(mux *mux.Router) {
	mux.Handle("/interactions", errors.ErrorHandler(middleware.IsAuth(h.GetAll, h.userStore))).Methods(http.MethodGet)
	mux.Handle("/interactions/{id}", errors.ErrorHandler(middleware.IsAuth(h.Get, h.userStore))).Methods(http.MethodGet)
	mux.Handle("/interactions", errors.ErrorHandler(middleware.IsAuth(h.Create, h.userStore, types.UserRoleParent, types.UserRoleChild))).Methods(http.MethodPost)
	mux.Handle("/interactions/{id}", errors.ErrorHandler(middleware.IsAuth(h.Update, h.userStore, types.UserRoleStandHolder))).Methods(http.MethodPatch)
}

func (h *InteractionHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	interactions, err := h.service.GetAll(r.Context(), utils.GetQueryParams(r))
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, interactions); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (h *InteractionHandler) Get(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	interaction, err := h.service.Get(r.Context(), id)
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, interaction); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (h *InteractionHandler) Create(w http.ResponseWriter, r *http.Request) error {
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

func (h *InteractionHandler) Update(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if err := h.service.Update(r.Context(), id, input); err != nil {
		return err
	}

	if err := json.Write(w, http.StatusAccepted, nil); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
