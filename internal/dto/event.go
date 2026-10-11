package dto

import "time"

import "github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (r WithdrawRequest) Validate() error {
	if r.Order == "" || r.Sum <= 0 {
		return apperrors.BadRequest(invalidRequestBody)
	}
	if !ValidLuhn(r.Order) {
		return apperrors.ValidationError(invalidOrderNumber)
	}
	return nil
}

type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
