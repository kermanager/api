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
)

type UserService interface {
	Get(ctx context.Context, id int) (types.UserBasic, error)
	Invite(ctx context.Context, input map[string]interface{}) error
	Pay(ctx context.Context, input map[string]interface{}) error

	SignUp(ctx context.Context, input map[string]interface{}) error
	SignIn(ctx context.Context, input map[string]interface{}) (types.UserBasicWithToken, error)
	GetMe(ctx context.Context) (types.UserBasic, error)
}

type Service struct {
	store UserStore
}

func NewService(store UserStore) *Service {
	return &Service{
		store: store,
	}
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

// TODO: all users with role parent
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

	// TODO: Send email :
	// - To: input["email"]
	// - Subject: "Invitation to join our platform"
	// - Body: "You have been invited to join our platform. Your credentials are as follows: Email: input["email"], Password: randomPassword"

	return nil
}

// TODO: all users with role parent
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

func (s *Service) SignIn(ctx context.Context, input map[string]interface{}) (types.UserBasicWithToken, error) {
	user, err := s.store.FindByEmail(input["email"].(string))
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.UserBasicWithToken{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.UserBasicWithToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if !hasher.Compare(user.Password, input["password"].(string)) {
		return types.UserBasicWithToken{}, errors.CustomError{
			Key: errors.InvalidCredentials,
			Err: goErrors.New("invalid credentials"),
		}
	}

	expiresIn, err := strconv.Atoi(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil {
		return types.UserBasicWithToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	token, err := jwt.Create(os.Getenv("JWT_SECRET"), expiresIn, user.Id)
	if err != nil {
		if goErrors.Is(err, goJwt.ErrTokenExpired) || goErrors.Is(err, goJwt.ErrSignatureInvalid) {
			return types.UserBasicWithToken{}, errors.CustomError{
				Key: errors.Unauthorized,
				Err: err,
			}
		}
		return types.UserBasicWithToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return types.UserBasicWithToken{
		Id:     user.Id,
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		Credit: user.Credit,
		Token:  token,
	}, nil
}

func (s *Service) GetMe(ctx context.Context) (types.UserBasic, error) {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		return types.UserBasic{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}

	user, err := s.store.FindById(userId)
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
