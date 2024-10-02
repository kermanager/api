package tombola

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type TombolaStore interface {
	FindAll(filters map[string]interface{}) ([]types.Tombola, error)
	FindById(id int) (types.Tombola, error)
	Create(input map[string]interface{}) error
	Update(id int, input map[string]interface{}) error
	UpdateStatus(id int, status string) error
	SetWinner(id int) error
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindAll(filters map[string]interface{}) ([]types.Tombola, error) {
	tombolas := []types.Tombola{}
	query := `
		SELECT DISTINCT
			t.id AS id,
			t.kermesse_id AS kermesse_id,
			t.name AS name,
			t.status AS status,
			t.price AS price,
			t.gift AS gift
		FROM tombolas t
		WHERE 1=1
	`
	if filters["kermesse_id"] != nil {
		query += fmt.Sprintf(" AND t.kermesse_id = %v", filters["kermesse_id"])
	}
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
	query := "INSERT INTO tombolas (kermesse_id, name, price, gift) VALUES ($1, $2, $3, $4)"
	_, err := s.db.Exec(query, input["kermesse_id"], input["name"], input["price"], input["gift"])

	return err
}

func (s *Store) Update(id int, input map[string]interface{}) error {
	query := "UPDATE tombolas SET name=$1, price=$2, gift=$3 WHERE id=$4"
	_, err := s.db.Exec(query, input["name"], input["price"], input["gift"], id)

	return err
}

func (s *Store) UpdateStatus(id int, status string) error {
	query := "UPDATE tombolas SET status=$1 WHERE id=$2"
	_, err := s.db.Exec(query, status, id)

	return err
}

func (s *Store) SetWinner(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query := "UPDATE tombolas SET status=$1 WHERE id=$2"
	_, err = tx.Exec(query, types.TombolaStatusEnded, id)
	if err != nil {
		return err
	}

	query = `
		UPDATE tickets
		SET is_winner = true
		WHERE id = (
			SELECT id
			FROM tickets
			WHERE tombola_id = $1
			ORDER BY RANDOM()
			LIMIT 1
		)
		AND tombola_id = $1
	`
	_, err = tx.Exec(query, id)
	if err != nil {
		return err
	}

	return err
}
