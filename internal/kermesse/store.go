package kermesse

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type KermesseStore interface {
	FindAll() ([]types.Kermesse, error)
	FindById(id int) (types.Kermesse, error)
	Create(input map[string]interface{}) error
	Update(id int, input map[string]interface{}) error
	End(id int) error
	CanEnd(id int) (bool, error)

	AddUser(input map[string]interface{}) error
	CanAddStand(standId int) (bool, error)
	AddStand(input map[string]interface{}) error
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindAll() ([]types.Kermesse, error) {
	kermesses := []types.Kermesse{}
	query := "SELECT * FROM kermesses"
	err := s.db.Select(&kermesses, query)

	return kermesses, err
}

func (s *Store) FindById(id int) (types.Kermesse, error) {
	kermesse := types.Kermesse{}
	query := "SELECT * FROM kermesses WHERE id=$1"
	err := s.db.Get(&kermesse, query, id)

	return kermesse, err
}

func (s *Store) Create(input map[string]interface{}) error {
	query := "INSERT INTO kermesses (user_id, name, description) VALUES ($1, $2, $3)"
	_, err := s.db.Exec(query, input["user_id"], input["name"], input["description"])

	return err
}

func (s *Store) Update(id int, input map[string]interface{}) error {
	query := "UPDATE kermesses SET name=$1, description=$2 WHERE id=$3"
	_, err := s.db.Exec(query, input["name"], input["description"], id)

	return err
}

func (s *Store) CanEnd(id int) (bool, error) {
	var isTrue bool
	query := "SELECT EXISTS ( SELECT 1 FROM tombolas WHERE kermesse_id = $1 AND status = $2 ) AS is_true"
	err := s.db.QueryRow(query, id, types.TombolaStatusStarted).Scan(&isTrue)

	return !isTrue, err
}

func (s *Store) End(id int) error {
	query := "UPDATE kermesses SET status=$1 WHERE id=$2"
	_, err := s.db.Exec(query, types.KermesseStatusEnded, id)

	return err
}

func (s *Store) AddUser(input map[string]interface{}) error {
	query := "INSERT INTO kermesses_users (kermesse_id, user_id) VALUES ($1, $2)"
	_, err := s.db.Exec(query, input["kermesse_id"], input["user_id"])

	return err
}

func (s *Store) CanAddStand(standId int) (bool, error) {
	var isTrue bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM kermesses_stands ks
  		JOIN kermesses k ON ks.kermesse_id = k.id
  		WHERE ks.stand_id = $1 AND k.status = $2
		) AS is_associated
 	`
	err := s.db.QueryRow(query, standId, types.KermesseStatusStarted).Scan(&isTrue)

	return !isTrue, err
}

func (s *Store) AddStand(input map[string]interface{}) error {
	query := "INSERT INTO kermesses_stands (kermesse_id, stand_id) VALUES ($1, $2)"
	_, err := s.db.Exec(query, input["kermesse_id"], input["stand_id"])

	return err
}
