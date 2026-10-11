package dto

import (
	"strings"
	"time"

	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
)

type UploadOrderRequest struct {
	Number string
}

func (r UploadOrderRequest) Validate() error {
	if strings.TrimSpace(r.Number) == "" {
		return apperrors.BadRequest(invalidRequestBody)
	}
	return nil
}

type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
