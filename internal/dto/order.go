package dto

import (
	"strings"
	"time"

	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/shopspring/decimal"
)

const invalidRequestBody = "invalid_request_body"
const invalidOrderNumber = "invalid_order_number"

type UploadOrderRequest struct {
	Number string
}

func (r UploadOrderRequest) Validate() error {
	number := strings.TrimSpace(r.Number)
	if number == "" {
		return apperrors.BadRequest(invalidRequestBody)
	}
	if !ValidLuhn(number) {
		return apperrors.ValidationError(invalidOrderNumber)
	}
	return nil
}

type OrderResponse struct {
	Number     string           `json:"number"`
	Status     string           `json:"status"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time        `json:"uploaded_at"`
}

func ValidLuhn(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		d := number[i]
		if d < '0' || d > '9' {
			return false
		}
		n := int(d - '0')
		if double {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}
	return sum%10 == 0
}
