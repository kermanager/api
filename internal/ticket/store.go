package ticket

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type TicketStore interface {
	FindAll(filters map[string]interface{}) ([]types.Ticket, error)
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

func (s *Store) FindAll(filters map[string]interface{}) ([]types.Ticket, error) {
	tickets := []types.Ticket{}
	query := `
		SELECT DISTINCT
			t.id AS id,
			t.is_winner AS is_winner,
			u.id AS "user.id",
			u.name AS "user.name",
			u.email AS "user.email",
			u.role AS "user.role",
			tb.id AS "tombola.id",
			tb.name AS "tombola.name",
			tb.status AS "tombola.status",
			tb.price AS "tombola.price",
			tb.gift AS "tombola.gift",
			k.id AS "kermesse.id",
			k.name AS "kermesse.name",
			k.description AS "kermesse.description",
			k.status AS "kermesse.status"
		FROM tickets t
		JOIN users u ON t.user_id = u.id
		JOIN tombolas tb ON t.tombola_id = tb.id
		JOIN kermesses k ON tb.kermesse_id = k.id
		WHERE 1=1
	`
	if filters["manager_id"] != nil {
		query += fmt.Sprintf(" AND k.user_id IS NOT NULL AND k.user_id = %v", filters["manager_id"])
	}
	if filters["parent_id"] != nil {
		query += fmt.Sprintf(" AND u.parent_id IS NOT NULL AND u.parent_id = %v", filters["parent_id"])
	}
	if filters["child_id"] != nil {
		query += fmt.Sprintf(" AND t.user_id IS NOT NULL AND t.user_id = %v", filters["child_id"])
	}
	err := s.db.Select(&tickets, query)

	return tickets, err
}

func (s *Store) FindById(id int) (types.Ticket, error) {
	ticket := types.Ticket{}
	query := `
		SELECT
			t.id AS id,
			t.is_winner AS is_winner,
			u.id AS "user.id",
			u.name AS "user.name",
			u.email AS "user.email",
			u.role AS "user.role",
			tb.id AS "tombola.id",
			tb.name AS "tombola.name",
			tb.status AS "tombola.status",
			tb.price AS "tombola.price",
			tb.gift AS "tombola.gift",
			k.id AS "kermesse.id",
			k.name AS "kermesse.name",
			k.description AS "kermesse.description",
			k.status AS "kermesse.status"
		FROM tickets t
		JOIN users u ON t.user_id = u.id
		JOIN tombolas tb ON t.tombola_id = tb.id
		JOIN kermesses k ON tb.kermesse_id = k.id
		WHERE t.id=$1
	`
	err := s.db.Get(&ticket, query, id)

	return ticket, err
}

func (s *Store) CanCreate(input map[string]interface{}) (bool, error) {
	var isAssociated bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM kermesses_users ku
			JOIN kermesses k ON k.id = ku.kermesse_id
			WHERE ku.kermesse_id = $1 AND ku.user_id = $2 AND k.status = $3
		) AS is_associated
	`
	err := s.db.QueryRow(query, input["kermesse_id"], input["user_id"], types.KermesseStatusStarted).Scan(&isAssociated)

	return isAssociated, err
}

func (s *Store) Create(input map[string]interface{}) error {
	query := "INSERT INTO tickets (user_id, tombola_id) VALUES ($1, $2)"
	_, err := s.db.Exec(query, input["user_id"], input["tombola_id"])

	return err
}
