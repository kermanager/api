package tombola

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type TombolaStore interface {
	FindAll() ([]types.Tombola, error)
	FindById(id int) (types.Tombola, error)
	Create(input map[string]interface{}) error
	Update(id int, input map[string]interface{}) error
	UpdateStatus(id int, status string) error
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindAll() ([]types.Tombola, error) {
	tombolas := []types.Tombola{}
	query := "SELECT * FROM tombolas"
	err := s.db.Select(&tombolas, query)

	return tombolas, err
}

func (s *Store) FindById(id int) (types.Tombola, error) {
	tombola := types.Tombola{}
	query := "SELECT * FROM tombolas WHERE id=$1"
	err := s.db.Get(&tombola, query, id)

	return tombola, err
}

func (s *Store) Create(input map[string]interface{}) error {
	query := "INSERT INTO tombolas (kermesse_id, user_id, name, price, gift) VALUES ($1, $2, $3, $4, $5)"
	_, err := s.db.Exec(query, input["kermesse_id"], input["user_id"], input["name"], input["price"], input["gift"])

	return err
}

func (s *Store) Update(id int, input map[string]interface{}) error {
	query := "UPDATE tombolas SET name=$1, price=$2, gift=$3 WHERE id=$4"
	_, err := s.db.Exec(query, input["name"], input["price"], input["gift"], id)

	return err
}

func (s *Store) UpdateStatus(id int, status string) error {
	query := "UPDATE tombolas SET status=$1 WHERE id=$4"
	_, err := s.db.Exec(query, status, id)

	return err
}
