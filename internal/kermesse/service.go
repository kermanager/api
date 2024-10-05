package kermesse

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermanager/internal/types"
	"github.com/kermanager/internal/user"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/utils"
)

type KermesseService interface {
	GetAll(ctx context.Context) ([]types.Kermesse, error)
	GetUsersInvite(ctx context.Context, id int) ([]types.UserBasic, error)
	Get(ctx context.Context, id int) (types.KermesseWithStats, error)
	Create(ctx context.Context, input map[string]interface{}) error
	Update(ctx context.Context, id int, input map[string]interface{}) error
	End(ctx context.Context, id int) error

	AddUser(ctx context.Context, input map[string]interface{}) error
	AddStand(ctx context.Context, input map[string]interface{}) error
}

type Service struct {
	store     KermesseStore
	userStore user.UserStore
}

func NewService(store KermesseStore, userStore user.UserStore) *Service {
	return &Service{
		store:     store,
		userStore: userStore,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]types.Kermesse, error) {
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
	} else if userRole == types.UserRoleStandHolder {
		filters["stand_holder_id"] = userId
	}

	kermesses, err := s.store.FindAll(filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return kermesses, nil
}

func (s *Service) GetUsersInvite(ctx context.Context, id int) ([]types.UserBasic, error) {
	users, err := s.store.FindUsersInvite(id)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return users, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.KermesseWithStats, error) {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return types.KermesseWithStats{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	userRole, ok := ctx.Value(types.UserRoleKey).(string)
	if !ok {
		return types.KermesseWithStats{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user role not found in context"),
		}
	}

	kermesse, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.KermesseWithStats{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.KermesseWithStats{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	filters := map[string]interface{}{}
	if userRole == types.UserRoleManager {
		filters["manager_id"] = userId
	} else if userRole == types.UserRoleParent {
		filters["parent_id"] = userId
	} else if userRole == types.UserRoleChild {
		filters["child_id"] = userId
	} else if userRole == types.UserRoleStandHolder {
		filters["stand_holder_id"] = userId
	}

	stats, err := s.store.Stats(id, filters)
	if err != nil {
		return types.KermesseWithStats{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	kermesseWithStats := types.KermesseWithStats{
		Id:                kermesse.Id,
		UserId:            kermesse.UserId,
		Name:              kermesse.Name,
		Description:       kermesse.Description,
		Status:            kermesse.Status,
		StandCount:        stats.StandCount,
		TombolaCount:      stats.TombolaCount,
		UserCount:         stats.UserCount,
		InteractionCount:  stats.InteractionCount,
		InteractionIncome: stats.InteractionIncome,
		TombolaIncome:     stats.TombolaIncome,
	}

	return kermesseWithStats, nil
}

func (s *Service) Create(ctx context.Context, input map[string]interface{}) error {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	input["user_id"] = userId

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
	kermesse, err := s.store.FindById(id)
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

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
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

func (s *Service) End(ctx context.Context, id int) error {
	kermesse, err := s.store.FindById(id)
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

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
		}
	}

	canEnd, err := s.store.CanEnd(id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if !canEnd {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse can't be ended"),
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

	err = s.store.End(id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) AddUser(ctx context.Context, input map[string]interface{}) error {
	kermesse, err := s.store.FindById(input["kermesse_id"].(int))
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

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
		}
	}

	managerId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	if kermesse.UserId != managerId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	childId, error := utils.GetIntFromMap(input, "user_id")
	if error != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: error,
		}
	}
	child, err := s.userStore.FindById(childId)
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
	if child.Role != types.UserRoleChild {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("user is not a child"),
		}
	}

	// invite child
	err = s.store.AddUser(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	// invite child parent if exists
	if child.ParentId != nil {
		input["user_id"] = child.ParentId
		err = s.store.AddUser(input)
		if err != nil {
			return errors.CustomError{
				Key: errors.InternalServerError,
				Err: err,
			}
		}
	}

	return nil
}

func (s *Service) AddStand(ctx context.Context, input map[string]interface{}) error {
	kermesse, err := s.store.FindById(input["kermesse_id"].(int))
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

	if kermesse.Status == types.KermesseStatusEnded {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is already ended"),
		}
	}

	standId, err := utils.GetIntFromMap(input, "stand_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	canAddStand, err := s.store.CanAddStand(standId)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if !canAddStand {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("stand is already associated with kermesse"),
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

	err = s.store.AddStand(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
