package interaction

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/kermesse"
	"github.com/kermanager/internal/stand"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/utils"
)

type InteractionService interface {
	GetAll(ctx context.Context, params map[string]interface{}) ([]types.InteractionBasic, error)
	Get(ctx context.Context, id int) (types.Interaction, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
}

type Service struct {
	store         InteractionStore
	standStore    stand.StandStore
	userStore     user.UserStore
	kermesseStore kermesse.KermesseStore
}

func NewService(store InteractionStore, standStore stand.StandStore, userStore user.UserStore, kermesseStore kermesse.KermesseStore) *Service {
	return &Service{
		store:         store,
		standStore:    standStore,
		userStore:     userStore,
		kermesseStore: kermesseStore,
	}
}

func (s *Service) GetAll(ctx context.Context, params map[string]interface{}) ([]types.InteractionBasic, error) {
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
	if userRole == types.UserRoleParent {
		filters["parent_id"] = userId
	} else if userRole == types.UserRoleChild {
		filters["child_id"] = userId
	} else if userRole == types.UserRoleStandHolder {
		filters["stand_holder_id"] = userId
	}
	if params["kermesse_id"] != nil {
		filters["kermesse_id"] = params["kermesse_id"]
	}

	interactions, err := s.store.FindAll(filters)
	if err != nil {
		return nil, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return interactions, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.Interaction, error) {
	interaction, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return interaction, errors.CustomError{
				Err: goErrors.New(errors.InvalidInput),
			}
		}
		return interaction, errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return interaction, nil
}

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	standId, err := utils.GetIntFromMap(input, "stand_id")
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.InvalidInput),
		}
	}
	stand, err := s.standStore.FindById(standId)
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
	user, err := s.userStore.FindById(userId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Err: goErrors.New(errors.NotAllowed),
			}
		}
		return errors.CustomError{
			Err: goErrors.New(errors.NotAllowed),
		}
	}

	canCreate, err := s.store.CanCreate(map[string]interface{}{
		"user_id":  userId,
		"stand_id": standId,
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

	// calculate total price
	quantity := 1
	totalPrice := stand.Price
	if stand.Type == types.InteractionTypeConsumption {
		quantity, err = utils.GetIntFromMap(input, "quantity")
		if err != nil {
			return errors.CustomError{
				Err: goErrors.New(errors.InvalidInput),
			}
		}
		totalPrice = stand.Price * quantity
	}

	// check stand's stock and user credit
	if stand.Type == types.InteractionTypeConsumption {
		if stand.Stock < quantity {
			return errors.CustomError{
				Err: goErrors.New(errors.NotEnoughStock),
			}
		}
	}

	// check user's credit
	if user.Credit < totalPrice {
		return errors.CustomError{
			Err: goErrors.New(errors.NotEnoughCredits),
		}
	}

	// decrease stand's stock
	if stand.Type == types.InteractionTypeConsumption {
		err = s.standStore.UpdateStock(standId, -quantity)
		if err != nil {
			return errors.CustomError{
				Err: goErrors.New(errors.ServerError),
			}
		}
	}

	// decrease user's credit
	err = s.userStore.UpdateCredit(userId, -totalPrice)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	// increase stand holder's credit
	err = s.userStore.UpdateCredit(stand.UserId, totalPrice)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	input["user_id"] = user.Id
	input["type"] = stand.Type
	input["credit"] = totalPrice

	err = s.store.Create(input)
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}

func (s *Service) Update(ctx context.Context, id int, input map[string]interface{}) error {
	interaction, err := s.store.FindById(id)
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

	if interaction.Type != types.InteractionTypeActivity {
		return errors.CustomError{
			Err: goErrors.New(errors.IsNotAnActivity),
		}
	}

	kermesse, err := s.kermesseStore.FindById(interaction.Kermesse.Id)
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

	stand, err := s.standStore.FindById(interaction.Stand.Id)
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

	err = s.store.Update(id, map[string]interface{}{
		"status": types.InteractionStatusEnded,
		"point":  input["point"],
	})
	if err != nil {
		return errors.CustomError{
			Err: goErrors.New(errors.ServerError),
		}
	}

	return nil
}
