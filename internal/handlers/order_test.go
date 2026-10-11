package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/dto"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderHandler_create(t *testing.T) {
	h := &OrderHandler{orders: &fakeOrders{created: models.Order{
		OrderNumber: "2377222919",
		Status:      models.OrderStatusRegistered,
	}}}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("2377222919"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(resources.UserIDKey, uuid.New())

	require.NoError(t, h.create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var got dto.OrderResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "2377222919", got.Number)
	assert.Equal(t, models.OrderStatusRegistered, got.Status)
}

func TestOrderHandler_create_Conflict(t *testing.T) {
	h := &OrderHandler{orders: &fakeOrders{err: apperrors.Conflict(services.OrderAlreadyExists)}}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("2377222919"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(resources.UserIDKey, uuid.New())

	require.NoError(t, h.create(c))
	assert.Equal(t, http.StatusOK, rec.Code)
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
