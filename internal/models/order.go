package models

import (
	"github.com/google/uuid"
)

type Order struct {
	ID          uuid.UUID `db:"id"`
	OrderNumber string    `db:"order_number"`
	UserID      uuid.UUID `db:"user_id"`
	Status      string    `db:"status"`
	Accrual     float64   `db:"accrual"`
}
