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
	"github.com/nazimov1amu/alex-practicum-diploma/internal/dto"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_register(t *testing.T) {
	h := &UserHandler{users: &fakeUsers{tokens: services.TokenPair{AccessToken: "a", RefreshToken: "r"}}}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"login":"alice","password":"secret"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, h.register(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Bearer a", rec.Header().Get(echo.HeaderAuthorization))

	var got dto.TokenPairResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "a", got.AccessToken)
	assert.Equal(t, "r", got.RefreshToken)
}

func TestUserHandler_balance(t *testing.T) {
	h := &UserHandler{users: &fakeUsers{balance: models.UserBalance{Current: 100, Withdrawn: 20}}}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(resources.UserIDKey, uuid.New())

	require.NoError(t, h.balance(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var got dto.BalanceResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, 100.0, got.Current)
	assert.Equal(t, 20.0, got.Withdrawn)
}

type fakeUsers struct {
	tokens  services.TokenPair
	balance models.UserBalance
}

func (f *fakeUsers) Register(context.Context, string, string) (services.TokenPair, error) {
	return f.tokens, nil
}

func (f *fakeUsers) Login(context.Context, string, string) (services.TokenPair, error) {
	return f.tokens, nil
}

func (f *fakeUsers) Refresh(context.Context, string) (services.TokenPair, error) {
	return f.tokens, nil
}

func (f *fakeUsers) GetBalance(context.Context, uuid.UUID) (models.UserBalance, error) {
	return f.balance, nil
}
