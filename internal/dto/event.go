package dto

import (
	"time"

	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/shopspring/decimal"
)

type WithdrawRequest struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}

func (r WithdrawRequest) Validate() error {
	if r.Order == "" || !r.Sum.IsPositive() {
		return apperrors.BadRequest(invalidRequestBody)
	}
	if !ValidLuhn(r.Order) {
		return apperrors.ValidationError(invalidOrderNumber)
	}
	return nil
}

type WithdrawalResponse struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}
