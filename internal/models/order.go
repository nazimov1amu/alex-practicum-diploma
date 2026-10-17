package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	OrderStatusNew        = "NEW"
	OrderStatusRegistered = "REGISTERED"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	ID          uuid.UUID       `db:"id"`
	OrderNumber string          `db:"order_number"`
	UserID      uuid.UUID       `db:"user_id"`
	Status      string          `db:"status"`
	Accrual     decimal.Decimal `db:"accrual"`
	CreatedAt   time.Time       `db:"created_at"`
}
