package stand

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type StandStore interface {
	FindAll() ([]types.Stand, error)
	FindById(id int) (types.Stand, error)
	Create(input map[string]interface{}) error
	Update(id int, input map[string]interface{}) error
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindAll() ([]types.Stand, error) {
	stands := []types.Stand{}
	query := "SELECT * FROM stands"
	err := s.db.Select(&stands, query)

	return stands, err
}

func (s *Store) FindById(id int) (types.Stand, error) {
	stand := types.Stand{}
	query := "SELECT * FROM stands WHERE id=$1"
	err := s.db.Get(&stand, query, id)

	return stand, err
}

func (s *Store) Create(input map[string]interface{}) error {
	query := "INSERT INTO stands (user_id, name, description, price, stock) VALUES ($1, $2, $3, $4, $5)"
	_, err := s.db.Exec(query, input["user_id"], input["name"], input["description"], input["price"], input["stock"])

	return err
}

func (s *Store) Update(id int, input map[string]interface{}) error {
	query := "UPDATE stands SET name=$1, description=$2, price=$3, stock=$4 WHERE id=$5"
	_, err := s.db.Exec(query, input["name"], input["description"], input["price"], input["stock"], id)

	return err
}
