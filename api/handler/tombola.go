package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kermanager/api/middleware"
	"github.com/kermanager/internal/tombola"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/json"
	"github.com/kermanager/pkg/utils"
)

type TombolaHandler struct {
	service   tombola.TombolaService
	userStore user.UserStore
}

func NewTombolaHandler(service tombola.TombolaService, userStore user.UserStore) *TombolaHandler {
	return &TombolaHandler{
		service:   service,
		userStore: userStore,
	}
}

func (h *TombolaHandler) RegisterRoutes(mux *mux.Router) {
	mux.Handle("/tombolas", errors.ErrorHandler(middleware.IsAuth(h.GetAll, h.userStore))).Methods(http.MethodGet)
	mux.Handle("/tombolas/{id}", errors.ErrorHandler(middleware.IsAuth(h.Get, h.userStore))).Methods(http.MethodGet)
	mux.Handle("/tombolas", errors.ErrorHandler(middleware.IsAuth(h.Create, h.userStore, types.UserRoleManager))).Methods(http.MethodPost)
	mux.Handle("/tombolas/{id}", errors.ErrorHandler(middleware.IsAuth(h.Update, h.userStore, types.UserRoleManager))).Methods(http.MethodPatch)
	mux.Handle("/tombolas/{id}/start", errors.ErrorHandler(middleware.IsAuth(h.Start, h.userStore, types.UserRoleManager))).Methods(http.MethodPatch)
	mux.Handle("/tombolas/{id}/end", errors.ErrorHandler(middleware.IsAuth(h.End, h.userStore, types.UserRoleManager))).Methods(http.MethodPatch)
}

func (h *TombolaHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	tombolas, err := h.service.GetAll(r.Context(), utils.GetQueryParams(r))
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, tombolas); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (h *TombolaHandler) Get(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	tombola, err := h.service.Get(r.Context(), id)
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, tombola); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (h *TombolaHandler) Create(w http.ResponseWriter, r *http.Request) error {
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

func (h *TombolaHandler) Update(w http.ResponseWriter, r *http.Request) error {
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

func (h *TombolaHandler) Start(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if err := h.service.Start(r.Context(), id); err != nil {
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

func (h *TombolaHandler) End(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if err := h.service.End(r.Context(), id); err != nil {
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
