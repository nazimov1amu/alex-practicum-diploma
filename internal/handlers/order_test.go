package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderHandler_create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name   string
		body   string
		orders *fakeOrders
		code   int
	}{
		{
			name:   "new",
			body:   "12345678903",
			orders: &fakeOrders{created: models.Order{OrderNumber: "2377222919"}},
			code:   http.StatusAccepted,
		},
		{
			name:   "same user",
			body:   "12345678903",
			orders: &fakeOrders{err: services.ErrOrderAlreadyUploaded},
			code:   http.StatusOK,
		},
		{
			name:   "other user",
			body:   "12345678903",
			orders: &fakeOrders{err: apperrors.Conflict(services.OrderAlreadyExists)},
			code:   http.StatusConflict,
		},
		{
			name:   "invalid number",
			body:   "12345678902",
			orders: &fakeOrders{},
			code:   http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &OrderHandler{orders: tc.orders}
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set(resources.UserIDKey, userID)

			err := h.create(c)
			if tc.code >= 400 {
				var httpErr *echo.HTTPError
				require.ErrorAs(t, err, &httpErr)
				assert.Equal(t, tc.code, httpErr.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.code, rec.Code)
		})
	}
}

func TestOrderHandler_list(t *testing.T) {
	h := &OrderHandler{orders: &fakeOrders{}}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(resources.UserIDKey, uuid.New())

	require.NoError(t, h.list(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

type fakeOrders struct {
	created models.Order
	err     error
}

func (f *fakeOrders) Create(context.Context, models.Order) (models.Order, error) {
	return f.created, f.err
}

func (f *fakeOrders) List(context.Context, uuid.UUID) ([]models.Order, error) {
	return nil, f.err
}

func (f *fakeOrders) GetByOrderNumber(context.Context, string) (models.Order, error) {
	return models.Order{}, f.err
}
