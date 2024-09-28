package types

const (
	InteractionTypeBuyer    string = "CONSUMPTION"
	InteractionTypeActivity string = "ACTIVITY"
)

type Interaction struct {
	Id         int    `json:"id" db:"id"`
	UserId     int    `json:"user_id" db:"user_id"`
	KermesseId int    `json:"kermesse_id" db:"kermesse_id"`
	StandId    int    `json:"stand_id" db:"stand_id"`
	Type       string `json:"type" db:"type"`
	Credit     int    `json:"credit" db:"credit"`
	Point      int    `json:"point" db:"point"`
}
