package types

const (
	InteractionTypeConsumption string = "CONSUMPTION"
	InteractionTypeActivity    string = "ACTIVITY"

	InteractionStatusStarted string = "STARTED"
	InteractionStatusEnded   string = "ENDED"
)

type Interaction struct {
	Id         int    `json:"id" db:"id"`
	UserId     int    `json:"user_id" db:"user_id"`
	KermesseId int    `json:"kermesse_id" db:"kermesse_id"`
	StandId    int    `json:"stand_id" db:"stand_id"`
	Type       string `json:"type" db:"type"`
	Status     string `json:"status" db:"status"`
	Credit     int    `json:"credit" db:"credit"`
	Point      int    `json:"point" db:"point"`
}
