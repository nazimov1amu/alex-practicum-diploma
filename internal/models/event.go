package models

import (
	"github.com/google/uuid"
)

type EventType string

const (
	EventTypeDeposit    EventType = "deposit"
	EventTypeWithdrawal EventType = "withdrawal"
)

type Event struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	EventType   EventType `db:"event_type"`
	Amount      float64   `db:"amount"`
}
