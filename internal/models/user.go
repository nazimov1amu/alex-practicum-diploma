package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID       uuid.UUID       `db:"id"`
	Login    string          `db:"login"`
	Password string          `db:"password"`
	Balance  decimal.Decimal `db:"balance"`
}

type UserBalance struct {
	Current   decimal.Decimal `db:"current"`
	Withdrawn decimal.Decimal `db:"withdrawn"`
}
