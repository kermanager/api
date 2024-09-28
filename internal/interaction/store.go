package interaction

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type InteractionStore interface {
	FindAll() ([]types.Interaction, error)
	FindById(id int) (types.Interaction, error)
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

func (s *Store) FindAll() ([]types.Interaction, error) {
	interactions := []types.Interaction{}
	query := "SELECT * FROM interactions"
	err := s.db.Select(&interactions, query)

	return interactions, err
}

func (s *Store) FindById(id int) (types.Interaction, error) {
	interaction := types.Interaction{}
	query := "SELECT * FROM interactions WHERE id=$1"
	err := s.db.Get(&interaction, query, id)

	return interaction, err
}

func (s *Store) Create(input map[string]interface{}) error {
	query := "INSERT INTO interactions (user_id, kermesse_id, stand_id, type, credit) VALUES ($1, $2, $3, $4, $5)"
	_, err := s.db.Exec(query, input["user_id"], input["kermesse_id"], input["stand_id"], input["type"], input["credit"])

	return err
}

func (s *Store) Update(id int, input map[string]interface{}) error {
	query := "UPDATE interactions SET point=$1 WHERE id=$2"
	_, err := s.db.Exec(query, input["point"], id)

	return err
}
