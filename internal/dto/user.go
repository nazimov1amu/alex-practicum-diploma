package dto

import "github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"


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
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
