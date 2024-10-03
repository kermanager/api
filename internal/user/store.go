package user

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type UserStore interface {
	FindAll(filters map[string]interface{}) ([]types.UserBasic, error)
	FindById(id int) (types.User, error)
	FindByEmail(email string) (types.User, error)
	Create(input map[string]interface{}) error
	Update(id int, input map[string]interface{}) error
	UpdateCredit(id int, amount int) error
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindAll(filters map[string]interface{}) ([]types.UserBasic, error) {
	users := []types.UserBasic{}
	query := `
		SELECT DISTINCT
			u.id AS id,
			u.name AS name,
			u.email AS email,
			u.role AS role,
			u.credit AS credit
		FROM users u
		FULL OUTER JOIN kermesses_users ku ON u.id = ku.user_id
		WHERE 1=1
	`
	if filters["kermesse_id"] != nil {
		query += fmt.Sprintf(" AND ku.kermesse_id = %v", filters["kermesse_id"])
	}
	err := s.db.Select(&users, query)

	return users, err
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
	query := "INSERT INTO users (parent_id, name, email, password, role) VALUES ($1, $2, $3, $4, $5)"
	_, err := s.db.Exec(query, input["parent_id"], input["name"], input["email"], input["password"], input["role"])

	return err
}

func (s *Store) Update(id int, input map[string]interface{}) error {
	query := "UPDATE users SET password=$1 WHERE id=$2"
	_, err := s.db.Exec(query, input["new_password"], id)

	return err
}

func (s *Store) UpdateCredit(id int, amount int) error {
	query := "UPDATE users SET credit=credit+$1 WHERE id=$2"
	_, err := s.db.Exec(query, amount, id)

	return err
}
