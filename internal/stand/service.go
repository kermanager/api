package stand

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/types"
	"github.com/kermanager/pkg/errors"
)

type StandService interface {
	GetAll(ctx context.Context, params map[string]interface{}) ([]types.Stand, error)
	Get(ctx context.Context, id int) (types.Stand, error)
	GetCurrent(ctx context.Context) (types.Stand, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
	UpdateCurrent(ctx context.Context, input map[string]interface{}) error
}

type Service struct {
	store StandStore
}

func NewService(store StandStore) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) GetAll(ctx context.Context, params map[string]interface{}) ([]types.Stand, error) {
	filters := map[string]interface{}{}
	if params["kermesse_id"] != nil {
		filters["kermesse_id"] = params["kermesse_id"]
	}
	if params["is_free"] != nil {
		filters["is_free"] = params["is_free"]
	}

	stands, err := s.store.FindAll(params)
	if err != nil {
		return nil, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return stands, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.Stand, error) {
	stand, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return stand, errors.CustomError{
				Err: goErrors.New(errors.InvalidInput),
			}
		}
		return stand, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return stand, nil
}

func (s *Service) GetCurrent(ctx context.Context) (types.Stand, error) {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return types.Stand{}, errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	stand, err := s.store.FindByUserId(userId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return stand, errors.CustomError{
				Err: goErrors.New(errors.InvalidInput),
			}
		}
		return stand, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return stand, nil
}

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}
	input["user_id"] = userId

	err := s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}

func (s *Service) Update(ctx context.Context, id int, input map[string]interface{}) error {
	stand, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Err: goErrors.New(errors.InvalidInput),
			}
		}
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}
	if stand.UserId != userId {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	err = s.store.Update(id, input)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}

func (s *Service) UpdateCurrent(ctx context.Context, input map[string]interface{}) error {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	err := s.store.UpdateByUserId(userId, input)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}
