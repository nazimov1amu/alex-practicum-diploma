package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventHandler_withdraw(t *testing.T) {
	userID := uuid.New()
	events := &fakeEvents{}
	h := &EventHandler{events: events}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"order":"2377225624","sum":50}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(resources.UserIDKey, userID)

	require.NoError(t, h.withdraw(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, userID, events.got.UserID)
	assert.Equal(t, "2377225624", events.got.OrderNumber)
	assert.True(t, decimal.NewFromInt(50).Equal(events.got.Amount))
	assert.Equal(t, models.EventTypeWithdrawal, events.got.EventType)
}

func TestEventHandler_withdrawals(t *testing.T) {
	h := &EventHandler{events: &fakeEvents{}}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(resources.UserIDKey, uuid.New())

	require.NoError(t, h.withdrawals(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

type fakeEvents struct {
	got  models.Event
	list []models.Event
}

func (f *fakeEvents) Create(_ context.Context, event models.Event) (models.Event, error) {
	f.got = event
	return event, nil
}

func (f *fakeEvents) ListWithdrawals(context.Context, uuid.UUID) ([]models.Event, error) {
	return f.list, nil
}
