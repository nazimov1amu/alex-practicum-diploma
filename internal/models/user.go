package models

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	Password string    `db:"password"`
	Balance  float64   `db:"balance"`
}

type UserBalance struct {
	Current   float64 `db:"current"`
	Withdrawn float64 `db:"withdrawn"`
}
