package dto

import (
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/shopspring/decimal"
)

func init() {
	// Gophermart sends amounts as JSON numbers. DECIMAL(10, 2) fits in that format.
	decimal.MarshalJSONWithoutQuotes = true
}

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (c Credentials) Validate() error {
	if c.Login == "" || c.Password == "" {
		return apperrors.BadRequest(invalidRequestBody)
	}
	return nil
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r RefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return apperrors.BadRequest(invalidRequestBody)
	}
	return nil
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type BalanceResponse struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}
