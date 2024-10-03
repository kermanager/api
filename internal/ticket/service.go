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
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	userRole, ok := ctx.Value(types.UserRoleKey).(string)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user role not found in context"),
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
	user, err := s.userStore.FindById(userId)
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

	// check if user has enough credit
	if user.Credit < tombola.Price {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("not enough credit"),
		}
	}

	// check if user belongs to the kermesse
	canCreate, err := s.store.CanCreate(map[string]interface{}{
		"kermesse_id": tombola.KermesseId,
		"user_id":     userId,
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

	// decrease user's credit
	err = s.userStore.UpdateCredit(userId, -tombola.Price)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	input["user_id"] = userId

	err = s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
