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
	"github.com/kermanager/pkg/hasher"
	"github.com/kermanager/pkg/jwt"
)

type UserService interface {
	Get(ctx context.Context, id int) (types.UserBasic, error)

	SignUp(ctx context.Context, input map[string]interface{}) error
	SignIn(ctx context.Context, input map[string]interface{}) (types.UserBasicWithToken, error)
	GetMe(ctx context.Context) (map[string]interface{}, error)
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
