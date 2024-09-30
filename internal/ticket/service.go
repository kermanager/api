package ticket

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/tombola"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/utils"
)

type TicketService interface {
	GetAll(ctx context.Context) ([]types.Ticket, error)
	Get(ctx context.Context, id int) (types.Ticket, error)
	Create(ctx context.Context, input map[string]interface{}) error
}

type Service struct {
	store        TicketStore
	tombolaStore tombola.TombolaStore
}

func NewService(store TicketStore, tombolaStore tombola.TombolaStore) *Service {
	return &Service{
		store:        store,
		tombolaStore: tombolaStore,
	}
}

// TODO: Permissions not decided yet
func (s *Service) GetAll(ctx context.Context) ([]types.Ticket, error) {
	tickets, err := s.store.FindAll()
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return tickets, nil
}

// TODO: Permissions not decided yet
func (s *Service) Get(ctx context.Context, id int) (types.Ticket, error) {
	ticket, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return ticket, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return ticket, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return ticket, nil
}

// TODO: all users with role child
func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	tombolaId, err := utils.GetIntFromMap(input, "tombola_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	tombola, err := s.tombolaStore.FindById(tombolaId)
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

	if tombola.Status != types.TombolaStatusStarted {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("tombola is not started or already finished"),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	input["user_id"] = userId

	canCreate, err := s.store.CanCreate(map[string]interface{}{
		"tombola_id": tombolaId,
		"user_id":    userId,
	})
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if !canCreate {
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
