package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	OrderStatusRegistered = "REGISTERED"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	ID          uuid.UUID `db:"id"`
	OrderNumber string    `db:"order_number"`
	UserID      uuid.UUID `db:"user_id"`
	Status      string    `db:"status"`
	Accrual     float64   `db:"accrual"`
	CreatedAt   time.Time `db:"created_at"`
}
