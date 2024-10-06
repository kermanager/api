package ticket

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/tombola"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
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
	userStore    user.UserStore
}

func NewService(store TicketStore, tombolaStore tombola.TombolaStore, userStore user.UserStore) *Service {
	return &Service{
		store:        store,
		tombolaStore: tombolaStore,
		userStore:    userStore,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]types.Ticket, error) {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return nil, errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}
	userRole, ok := ctx.Value(types.UserRoleKey).(string)
	if !ok {
		return nil, errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	filters := map[string]interface{}{}
	if userRole == types.UserRoleManager {
		filters["manager_id"] = userId
	} else if userRole == types.UserRoleParent {
		filters["parent_id"] = userId
	} else if userRole == types.UserRoleChild {
		filters["child_id"] = userId
	}

	tickets, err := s.store.FindAll(filters)
	if err != nil {
		return nil, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return tickets, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.Ticket, error) {
	ticket, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return ticket, errors.CustomError{
				Err: goErrors.New(errors.InvalidInput),
			}
		}
		return ticket, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return ticket, nil
}

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	tombolaId, err := utils.GetIntFromMap(input, "tombola_id")
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.InvalidInput),
		}
	}
	tombola, err := s.tombolaStore.FindById(tombolaId)
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

	if tombola.Status != types.TombolaStatusStarted {
		return errors.CustomError{
			Err: goErrors.New(errors.TombolaAlreadyEnded),
		}
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}
	user, err := s.userStore.FindById(userId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Err: goErrors.New(errors.NotAllowed),
			}
		}
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	// check if user has enough credit
	if user.Credit < tombola.Price {
		return errors.CustomError{
			Err: goErrors.New(errors.NotEnoughCredits),
		}
	}

	// check if user belongs to the kermesse
	canCreate, err := s.store.CanCreate(map[string]interface{}{
		"kermesse_id": tombola.KermesseId,
		"user_id":     userId,
	})
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}
	if !canCreate {
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	// decrease user's credit
	err = s.userStore.UpdateCredit(userId, -tombola.Price)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	input["user_id"] = userId
	err = s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}
