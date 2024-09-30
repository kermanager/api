package ticket

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type TicketStore interface {
	FindAll() ([]types.Ticket, error)
	FindById(id int) (types.Ticket, error)
	Create(input map[string]interface{}) error
	CanCreate(input map[string]interface{}) (bool, error)
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) FindAll() ([]types.Ticket, error) {
	tickets := []types.Ticket{}
	query := "SELECT * FROM tickets"
	err := s.db.Select(&tickets, query)

	return tickets, err
}

func (s *Store) FindById(id int) (types.Ticket, error) {
	ticket := types.Ticket{}
	query := "SELECT * FROM tickets WHERE id=$1"
	err := s.db.Get(&ticket, query, id)

	return ticket, err
}

func (s *Store) CanCreate(input map[string]interface{}) (bool, error) {
	var isAssociated bool
	query := "SELECT EXISTS ( SELECT 1 FROM kermesses_users WHERE kermesse_id = $1 AND user_id = $2 ) AS is_associated"
	err := s.db.QueryRow(query, input["kermesse_id"], input["user_id"]).Scan(&isAssociated)

	return isAssociated, err
}

func (s *Store) Create(input map[string]interface{}) error {
	query := "INSERT INTO tickets (user_id, tombola_id) VALUES ($1, $2)"
	_, err := s.db.Exec(query, input["user_id"], input["tombola_id"])

	return err
}
