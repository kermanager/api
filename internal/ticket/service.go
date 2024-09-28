package ticket

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/types"
	"github.com/kermanager/pkg/errors"
)

type TicketService interface {
	GetAll(ctx context.Context) ([]types.Ticket, error)
	Get(ctx context.Context, id int) (types.Ticket, error)
	Create(ctx context.Context, input map[string]interface{}) error
}

type Service struct {
	store TicketStore
}

func NewService(store TicketStore) *Service {
	return &Service{
		store: store,
	}
}

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
