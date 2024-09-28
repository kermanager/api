package types

type Ticket struct {
	Id        int `json:"id" db:"id"`
	UserId    int `json:"user_id" db:"user_id"`
	TombolaId int `json:"tombola_id" db:"tombola_id"`
}
