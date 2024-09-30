package types

const (
	KermesseStatusStarted string = "STARTED"
	KermesseStatusEnded   string = "ENDED"
)

type Kermesse struct {
	Id          int    `json:"id" db:"id"`
	UserId      int    `json:"user_id" db:"user_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Status      string `json:"status" db:"status"`
}
