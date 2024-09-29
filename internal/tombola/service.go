package tombola

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/ticket"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/pkg/errors"
)

type TombolaService interface {
	GetAll(ctx context.Context) ([]types.Tombola, error)
	Get(ctx context.Context, id int) (types.Tombola, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
	Start(ctx context.Context, id int) error
	End(ctx context.Context, id int, input map[string]interface{}) error
}

type Service struct {
	store       TombolaStore
	ticketStore ticket.TicketStore
}

func NewService(store TombolaStore, ticketStore ticket.TicketStore) *Service {
	return &Service{
		store:       store,
		ticketStore: ticketStore,
	}
}

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

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
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
	_, err := s.store.FindById(id)
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

	err = s.store.Update(id, input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) Start(ctx context.Context, id int) error {
	_, err := s.store.FindById(id)
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

	err = s.store.UpdateStatus(id, types.TombolaStatusStarted)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) End(ctx context.Context, id int, input map[string]interface{}) error {
	_, err := s.store.FindById(id)
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

	err = s.store.UpdateStatus(id, types.TombolaStatusEnded)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	// TODO: check for bug
	ticket, err := s.ticketStore.FindByUserIdAndTombolaId(input["user_id"].(int), id)
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

	err = s.ticketStore.SetWinner(ticket.Id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
