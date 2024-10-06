package user

import (
	"context"
	"database/sql"
	goErrors "errors"
	"os"
	"strconv"

	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/kermanager/internal/types"
	"github.com/kermanager/pkg/errors"
	"github.com/kermanager/pkg/generator"
	"github.com/kermanager/pkg/hasher"
	"github.com/kermanager/pkg/jwt"
	"github.com/kermanager/pkg/utils"
	"github.com/kermanager/third_party/resend"
)

type UserService interface {
	GetAll(ctx context.Context, params map[string]interface{}) ([]types.UserBasic, error)
	GetAllChildren(ctx context.Context, params map[string]interface{}) ([]types.UserBasic, error)
	Get(ctx context.Context, id int) (types.UserBasic, error)
	Update(ctx context.Context, id int, input map[string]interface{}) error
	UpdateCredit(userId, credit int) error
	Invite(ctx context.Context, input map[string]interface{}) error
	Pay(ctx context.Context, input map[string]interface{}) error

	SignUp(ctx context.Context, input map[string]interface{}) error
	SignIn(ctx context.Context, input map[string]interface{}) (types.UserMe, error)
	GetMe(ctx context.Context) (types.UserMe, error)
}

type Service struct {
	store         UserStore
	resendService resend.ResendService
}

func NewService(store UserStore, resendService resend.ResendService) *Service {
	return &Service{
		store:         store,
		resendService: resendService,
	}
}

func (s *Service) GetAll(ctx context.Context, params map[string]interface{}) ([]types.UserBasic, error) {
	filters := map[string]interface{}{}
	if params["kermesse_id"] != nil {
		filters["kermesse_id"] = params["kermesse_id"]
	}

	users, err := s.store.FindAll(filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return users, nil
}

func (s *Service) GetAllChildren(ctx context.Context, params map[string]interface{}) ([]types.UserBasic, error) {
	filters := map[string]interface{}{}
	if params["kermesse_id"] != nil {
		filters["kermesse_id"] = params["kermesse_id"]
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}

	users, err := s.store.FindAllChildren(userId, filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return users, nil
}

func (s *Service) Get(ctx context.Context, id int) (types.UserBasic, error) {
	user, err := s.store.FindById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.UserBasic{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.UserBasic{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return types.UserBasic{
		Id:     user.Id,
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		Credit: user.Credit,
	}, nil
}

func (s *Service) Update(ctx context.Context, id int, input map[string]interface{}) error {
	user, err := s.store.FindById(id)
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
	if user.Id != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	if !hasher.Compare(user.Password, input["password"].(string)) {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("invalid password"),
		}
	}

	hashedPassword, err := hasher.Hash(input["new_password"].(string))
	if err != nil {
		return err
	}
	input["new_password"] = hashedPassword

	err = s.store.Update(id, input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) Invite(ctx context.Context, input map[string]interface{}) error {
	_, err := s.store.FindByEmail(input["email"].(string))
	if err == nil {
		return errors.CustomError{
			Key: errors.EmailAlreadyExists,
			Err: goErrors.New("email already exists"),
		}
	}

	randomPassword, err := generator.RandomPassword(8)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	hashedPassword, err := hasher.Hash(randomPassword)
	if err != nil {
		return err
	}

	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}

	err = s.store.Create(map[string]interface{}{
		"name":      input["name"],
		"email":     input["email"],
		"password":  hashedPassword,
		"role":      types.UserRoleChild,
		"parent_id": userId,
	})
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	// send email to child
	_, err = s.resendService.SendInvitationEmail(input["email"].(string), input["email"].(string), randomPassword)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) UpdateCredit(userId, credit int) error {
	user, err := s.store.FindById(userId)
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
	if user.Role != types.UserRoleParent {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	err = s.store.UpdateCredit(userId, credit)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) Pay(ctx context.Context, input map[string]interface{}) error {
	childId, err := utils.GetIntFromMap(input, "child_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	child, err := s.store.FindById(childId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("child not found"),
		}
	}

	parentId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	parent, err := s.store.FindById(parentId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("parent not found"),
		}
	}

	if child.ParentId == nil || *child.ParentId != parent.Id {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	amount, error := utils.GetIntFromMap(input, "amount")
	if error != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: error,
		}
	}
	if parent.Credit < amount {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("insufficient credit"),
		}
	}

	err = s.store.UpdateCredit(childId, amount)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	err = s.store.UpdateCredit(parentId, -amount)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (s *Service) SignUp(ctx context.Context, input map[string]interface{}) error {
	_, err := s.store.FindByEmail(input["email"].(string))
	if err == nil {
		return errors.CustomError{
			Key: errors.EmailAlreadyExists,
			Err: goErrors.New("email already exists"),
		}
	}

	hashedPassword, err := hasher.Hash(input["password"].(string))
	if err != nil {
		return err
	}
	input["password"] = hashedPassword
	input["parent_id"] = nil

	if input["role"] == types.UserRoleChild {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("role cannot be child"),
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

func (s *Service) SignIn(ctx context.Context, input map[string]interface{}) (types.UserMe, error) {
	user, err := s.store.FindByEmail(input["email"].(string))
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.UserMe{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.UserMe{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if !hasher.Compare(user.Password, input["password"].(string)) {
		return types.UserMe{}, errors.CustomError{
			Key: errors.InvalidCredentials,
			Err: goErrors.New("invalid credentials"),
		}
	}

	expiresIn, err := strconv.Atoi(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil {
		return types.UserMe{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	token, err := jwt.Create(os.Getenv("JWT_SECRET"), expiresIn, user.Id)
	if err != nil {
		if goErrors.Is(err, goJwt.ErrTokenExpired) || goErrors.Is(err, goJwt.ErrSignatureInvalid) {
			return types.UserMe{}, errors.CustomError{
				Key: errors.Unauthorized,
				Err: err,
			}
		}
		return types.UserMe{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	hasStand, err := s.store.HasStand(user.Id)
	if err != nil {
		return types.UserMe{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return types.UserMe{
		Id:       user.Id,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		Credit:   user.Credit,
		HasStand: hasStand,
		Token:    token,
	}, nil
}

func (s *Service) GetMe(ctx context.Context) (types.UserMe, error) {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return types.UserMe{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}

	user, err := s.store.FindById(userId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.UserMe{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.UserMe{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	hasStand, err := s.store.HasStand(userId)
	if err != nil {
		return types.UserMe{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return types.UserMe{
		Id:       user.Id,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		Credit:   user.Credit,
		HasStand: hasStand,
		Token:    "",
	}, nil
}
