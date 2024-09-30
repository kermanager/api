package kermesse

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/utils"
)

type KermesseService interface {
	GetAll(ctx context.Context) ([]types.Kermesse, error)
	Get(ctx context.Context, id int) (types.Kermesse, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
	End(ctx context.Context, id int) error

	AddUser(ctx context.Context, input map[string]interface{}) error
	AddStand(ctx context.Context, input map[string]interface{}) error
}

type Service struct {
	store     KermesseStore
	userStore user.UserStore
}

func NewService(store KermesseStore, userStore user.UserStore) *Service {
	return &Service{
		store:     store,
		userStore: userStore,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]types.Kermesse, error) {
	kermesses, err := s.store.FindAll()
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return kermesses, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.Kermesse, error) {
	kermesse, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return kermesse, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return kermesse, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return kermesse, nil
}

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	input["user_id"] = userId

	err := s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) Update(ctx context.Context, id int, input map[string]interface{}) error {
	kermesse, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	err = s.store.Update(id, input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) End(ctx context.Context, id int) error {
	kermesse, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
		}
	}

	cantEnd, err := s.store.CantEnd(id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if cantEnd {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse can't be ended"),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	err = s.store.End(id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) AddUser(ctx context.Context, input map[string]interface{}) error {
	kermesse, err := s.store.FindById(input["kermesse_id"].(int))
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
		}
	}

	managerId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	if kermesse.UserId != managerId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	childId, error := utils.GetIntFromMap(input, "user_id")
	if error != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: error,
		}
	}
	child, err := s.userStore.FindById(childId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if child.Role != types.UserRoleChild {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("user is not a child"),
		}
	}

	// invite child parent if exists
	if child.ParentId != nil {
		input["user_id"] = child.ParentId
		err = s.store.AddUser(input)
		if err != nil {
			return errors.CustomError{
				Key: errors.InternalServerError,
				Err: err,
			}
		}
	}

	err = s.store.AddUser(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) AddStand(ctx context.Context, input map[string]interface{}) error {
	kermesse, err := s.store.FindById(input["kermesse_id"].(int))
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	err = s.store.AddStand(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
