package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EventType string

const (
	EventTypeDeposit    EventType = "deposit"
	EventTypeWithdrawal EventType = "withdrawal"
)

type Event struct {
	ID          uuid.UUID       `db:"id"`
	UserID      uuid.UUID       `db:"user_id"`
	OrderNumber string          `db:"order_number"`
	EventType   EventType       `db:"event_type"`
	Amount      decimal.Decimal `db:"amount"`
	CreatedAt   time.Time       `db:"created_at"`
}
