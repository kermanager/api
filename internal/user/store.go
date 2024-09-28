package user

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type UserStore interface {
	FindById(id int) (types.User, error)
	FindByEmail(email string) (types.User, error)
	Create(input map[string]interface{}) error
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindById(id int) (types.User, error) {
	user := types.User{}
	query := "SELECT * FROM users WHERE id=$1"
	err := s.db.Get(&user, query, id)

	return user, err
}

func (s *Store) FindByEmail(email string) (types.User, error) {
	user := types.User{}
	query := "SELECT * FROM users WHERE email=$1"
	err := s.db.Get(&user, query, email)

	return user, err
}

func (s *Store) Create(input map[string]interface{}) error {
	query := "INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4)"
	_, err := s.db.Exec(query, input["name"], input["email"], input["password"], input["role"])

	return err
}

func (s *Store) UpdateCredit(id int, n int) error {
	query := "UPDATE users SET credit=credit+$1 WHERE id=$2"
	_, err := s.db.Exec(query, n, id)

	return err
}
