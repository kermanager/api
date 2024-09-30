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
	GetAll(ctx context.Context) ([]types.Tombola, error)
	Get(ctx context.Context, id int) (types.Tombola, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
	Start(ctx context.Context, id int) error
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

// TODO: Permissions not decided yet
func (s *Service) GetAll(ctx context.Context) ([]types.Tombola, error) {
	tombolas, err := s.store.FindAll()
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return tombolas, nil
}

// TODO: Permissions not decided yet
func (s *Service) Get(ctx context.Context, id int) (types.Tombola, error) {
	tombola, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return tombola, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return tombola, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return tombola, nil
}

// TODO: All users with role manager, and manager of kermesse
func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	kermesseId, error := utils.GetIntFromMap(input, "kermesse_id")
	if error != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: error,
		}
	}
	kermesse, err := s.kermesseStore.FindById(kermesseId)
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

	err = s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

// TODO: All users with role manager, and manager of kermesse
func (s *Service) Update(ctx context.Context, id int, input map[string]interface{}) error {
	tombola, err := s.store.FindById(id)
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

	kermesse, err := s.kermesseStore.FindById(tombola.KermesseId)
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

// TODO: All users with role manager, and manager of kermesse
func (s *Service) Start(ctx context.Context, id int) error {
	tombola, err := s.store.FindById(id)
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

	kermesse, err := s.kermesseStore.FindById(tombola.KermesseId)
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

	err = s.store.UpdateStatus(id, types.TombolaStatusStarted)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

// TODO: All users with role manager, and manager of kermesse
func (s *Service) End(ctx context.Context, id int) error {
	tombola, err := s.store.FindById(id)
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

	kermesse, err := s.kermesseStore.FindById(tombola.KermesseId)
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

	if tombola.Status != types.TombolaStatusStarted {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("tombola is not started"),
		}
	}

	err = s.store.SetWinner(id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
