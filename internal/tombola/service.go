package tombola

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/kermesse"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/utils"
)

type TombolaService interface {
	GetAll(ctx context.Context, params map[string]interface{}) ([]types.Tombola, error)
	Get(ctx context.Context, id int) (types.Tombola, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
	End(ctx context.Context, id int) error
}

type Service struct {
	store         TombolaStore
	kermesseStore kermesse.KermesseStore
}

func NewService(store TombolaStore, kermesseStore kermesse.KermesseStore) *Service {
	return &Service{
		store:         store,
		kermesseStore: kermesseStore,
	}
}

func (s *Service) GetAll(ctx context.Context, params map[string]interface{}) ([]types.Tombola, error) {
	filters := map[string]interface{}{}
	if params["kermesse_id"] != nil {
		filters["kermesse_id"] = params["kermesse_id"]
	}

	tombolas, err := s.store.FindAll(filters)
	if err != nil {
		return nil, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return tombolas, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.Tombola, error) {
	tombola, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return tombola, errors.CustomError{
				Err: goErrors.New(errors.InvalidInput),
			}
		}
		return tombola, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return tombola, nil
}

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	kermesseId, error := utils.GetIntFromMap(input, "kermesse_id")
	if error != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.InvalidInput),
		}
	}
	kermesse, err := s.kermesseStore.FindById(kermesseId)
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

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Err: goErrors.New(errors.KermesseAlreadyEnded),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	err = s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}

func (s *Service) Update(ctx context.Context, id int, input map[string]interface{}) error {
	tombola, err := s.store.FindById(id)
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

	kermesse, err := s.kermesseStore.FindById(tombola.KermesseId)
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

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Err: goErrors.New(errors.KermesseAlreadyEnded),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}
	if kermesse.UserId != userId {
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

func (s *Service) End(ctx context.Context, id int) error {
	tombola, err := s.store.FindById(id)
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

	kermesse, err := s.kermesseStore.FindById(tombola.KermesseId)
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

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Err: goErrors.New(errors.KermesseAlreadyEnded),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	if tombola.Status != types.TombolaStatusStarted {
		return errors.CustomError{
			Err: goErrors.New(errors.TombolaAlreadyEnded),
		}
	}

	err = s.store.SetWinner(id)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}
