package kermesse

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kermanager/internal/types"
)

type KermesseStore interface {
	FindAll(filters map[string]interface{}) ([]types.Kermesse, error)
	FindUsersInvite(id int) ([]types.UserBasic, error)
	FindById(id int) (types.Kermesse, error)
	Stats(id int, filters map[string]interface{}) (types.KermesseStats, error)
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

func (s *Store) FindAll(filters map[string]interface{}) ([]types.Kermesse, error) {
	kermesses := []types.Kermesse{}
	query := `
		SELECT DISTINCT
			k.id AS id,
			k.user_id AS user_id,
			k.name AS name,
			k.description AS description,
			k.status AS status
		FROM kermesses k
		FULL OUTER JOIN kermesses_users ku ON k.id = ku.kermesse_id
		FULL OUTER JOIN kermesses_stands ks ON k.id = ks.kermesse_id
		FULL OUTER JOIN stands s ON ks.stand_id = s.id
		WHERE 1=1
	`
	if filters["manager_id"] != nil {
		query += fmt.Sprintf(" AND k.user_id = %v", filters["manager_id"])
	}
	if filters["parent_id"] != nil {
		query += fmt.Sprintf(" AND ku.user_id = %v", filters["parent_id"])
	}
	if filters["child_id"] != nil {
		query += fmt.Sprintf(" AND ku.user_id = %v", filters["child_id"])
	}
	if filters["stand_holder_id"] != nil {
		query += fmt.Sprintf(" AND ks.stand_id IS NOT NULL AND s.user_id = %v", filters["stand_holder_id"])
	}
	err := s.db.Select(&kermesses, query)

	return kermesses, err
}

func (s *Store) FindUsersInvite(id int) ([]types.UserBasic, error) {
	users := []types.UserBasic{}
	query := `
		SELECT DISTINCT
			u.id AS id,
			u.name AS name,
			u.email AS email,
			u.role AS role,
			u.credit AS credit
		FROM users u
		LEFT JOIN kermesses_users ku ON u.id = ku.user_id
		WHERE u.id IS NOT NULL AND role='CHILD' AND ku.kermesse_id IS NULL OR ku.kermesse_id != $1
	`
	err := s.db.Select(&users, query, id)

	return users, err
}

func (s *Store) Stats(id int, filters map[string]interface{}) (types.KermesseStats, error) {
	standCount := 0
	query := "SELECT COUNT(*) FROM kermesses_stands WHERE kermesse_id=$1"
	err := s.db.Get(&standCount, query, id)
	if err != nil {
		return types.KermesseStats{}, err
	}

	tombolaCount := 0
	if filters["manager_id"] != nil {
		query := "SELECT COUNT(*) FROM tombolas WHERE kermesse_id=$1"
		err := s.db.Get(&tombolaCount, query, id)
		if err != nil {
			return types.KermesseStats{}, err
		}
	}

	userCount := 0
	if filters["manager_id"] != nil || filters["parent_id"] != nil {
		query := `
			SELECT COUNT(*)
			FROM kermesses_users ku
			JOIN users u ON ku.user_id = u.id
			WHERE ku.kermesse_id=$1
		`
		if filters["parent_id"] != nil {
			query += fmt.Sprintf(" AND u.role='%v' AND u.parent_id=%v", types.UserRoleChild, filters["parent_id"])
		}
		err := s.db.Get(&userCount, query, id)
		if err != nil {
			return types.KermesseStats{}, err
		}
	}

	interactionCount := 0
	if filters["manager_id"] != nil || filters["stand_holder_id"] != nil {
		query := `
			SELECT COUNT(*)
			FROM interactions i
			JOIN stands s ON i.stand_id = s.id
			WHERE i.kermesse_id=$1
		`
		if filters["stand_holder_id"] != nil {
			query += fmt.Sprintf(" AND s.user_id=%v", filters["stand_holder_id"])
		}
		err := s.db.Get(&interactionCount, query, id)
		if err != nil {
			return types.KermesseStats{}, err
		}
	}

	interactionIncome := 0
	if filters["manager_id"] != nil || filters["stand_holder_id"] != nil {
		query := `
			SELECT COALESCE(SUM(i.credit), 0)
			FROM interactions i
			JOIN stands s ON i.stand_id = s.id
			WHERE i.kermesse_id=$1
		`
		if filters["stand_holder_id"] != nil {
			query += fmt.Sprintf(" AND s.user_id=%v", filters["stand_holder_id"])
		}
		err := s.db.Get(&interactionIncome, query, id)
		if err != nil {
			return types.KermesseStats{}, err
		}
	}

	tombolaIncome := 0
	if filters["manager_id"] != nil {
		query := `
		SELECT COALESCE(SUM(tb.price), 0)
		FROM tickets t
		JOIN tombolas tb ON t.tombola_id = tb.id
		WHERE tb.kermesse_id=$1
	`
		err := s.db.Get(&tombolaIncome, query, id)
		if err != nil {
			return types.KermesseStats{}, err
		}
	}

	points := 0
	if filters["child_id"] != nil {
		query := "SELECT COALESCE(SUM(point), 0) FROM interactions WHERE kermesse_id=$1 AND user_id=$2"
		err = s.db.Get(&points, query, id, filters["child_id"])
	}

	return types.KermesseStats{
		StandCount:        standCount,
		TombolaCount:      tombolaCount,
		UserCount:         userCount,
		InteractionCount:  interactionCount,
		InteractionIncome: interactionIncome,
		TombolaIncome:     tombolaIncome,
		Points:            points,
	}, err
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
