package interaction

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/stand"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/utils"
)

type InteractionService interface {
	GetAll(ctx context.Context) ([]types.Interaction, error)
	Get(ctx context.Context, id int) (types.Interaction, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
}

type Service struct {
	store      InteractionStore
	standStore stand.StandStore
	userStore  user.UserStore
}

func NewService(store InteractionStore, standStore stand.StandStore, userStore user.UserStore) *Service {
	return &Service{
		store:      store,
		standStore: standStore,
		userStore:  userStore,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]types.Interaction, error) {
	interactions, err := s.store.FindAll()
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return interactions, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.Interaction, error) {
	interaction, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return interaction, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return interaction, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return interaction, nil
}

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	standId, err := utils.GetIntFromMap(input, "stand_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	stand, err := s.standStore.FindById(standId)
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

	if stand.Type == types.InteractionTypeConsumption {
		quantity, err := utils.GetIntFromMap(input, "quantity")
		if err != nil {
			return errors.CustomError{
				Key: errors.BadRequest,
				Err: err,
			}
		}
		// decrease stand's stock
		if stand.Stock < quantity {
			return errors.CustomError{
				Key: errors.BadRequest,
				Err: goErrors.New("not enough stock"),
			}
		}
		err = s.standStore.UpdateStock(standId, -quantity)
		if err != nil {
			return errors.CustomError{
				Key: errors.InternalServerError,
				Err: err,
			}
		}
		// decrease user's credit
		totalPrice := stand.Price * quantity
		if user.Credit < totalPrice {
			return errors.CustomError{
				Key: errors.BadRequest,
				Err: goErrors.New("not enough credit"),
			}
		}
		err = s.userStore.UpdateCredit(userId, -totalPrice)
		if err != nil {
			return errors.CustomError{
				Key: errors.InternalServerError,
				Err: err,
			}
		}
	} else {
		// decrease user's credit
		totalPrice := stand.Price
		if user.Credit < totalPrice {
			return errors.CustomError{
				Key: errors.BadRequest,
				Err: goErrors.New("not enough credit"),
			}
		}
		err = s.userStore.UpdateCredit(userId, -totalPrice)
		if err != nil {
			return errors.CustomError{
				Key: errors.InternalServerError,
				Err: err,
			}
		}
	}

	input["type"] = stand.Type
	err = s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) Update(ctx context.Context, id int, input map[string]interface{}) error {
	interaction, err := s.store.FindById(id)
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

	if interaction.Type != types.InteractionTypeActivity {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("interaction type is not activity"),
		}
	}

	err = s.store.Update(id, map[string]interface{}{
		"status": types.InteractionStatusEnded,
		"point":  input["point"],
	})
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
