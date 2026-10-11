package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/common"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeUserStore struct {
	created    models.User
	byLogin    models.User
	loginErr   error
	byID       models.User
	idErr      error
	balance    models.UserBalance
	balanceErr error
}

func (f *fakeUserStore) Create(_ context.Context, user models.User) (models.User, error) {
	user.ID = uuid.New()
	f.created = user
	return user, nil
}

func (f *fakeUserStore) GetByLogin(context.Context, string) (models.User, error) {
	return f.byLogin, f.loginErr
}

func (f *fakeUserStore) GetByID(context.Context, uuid.UUID) (models.User, error) {
	return f.byID, f.idErr
}

func (f *fakeUserStore) GetBalance(context.Context, uuid.UUID) (models.UserBalance, error) {
	return f.balance, f.balanceErr
}

func TestUserService_Register(t *testing.T) {
	users := &fakeUserStore{}
	svc := newUserService(t, users)

	tokens, err := svc.Register(context.Background(), "alice", "secret")
	require.NoError(t, err)
	assertTokenPair(t, tokens, users.created.ID)
	assert.Equal(t, "alice", users.created.Login)
	assert.NotEmpty(t, users.created.Password)
	assert.NotEqual(t, "secret", users.created.Password)
}

func TestUserService_Register_InvalidCredentials(t *testing.T) {
	svc := newUserService(t, &fakeUserStore{})

	_, err := svc.Register(context.Background(), "", "secret")
	assertAppError(t, err, apperrors.KindBadRequest, InvalidCredentials)
}

func TestUserService_Login(t *testing.T) {
	userID := uuid.New()
	svc := newUserService(t, &fakeUserStore{})
	hash, err := svc.hashPassword("secret")
	require.NoError(t, err)

	users := &fakeUserStore{byLogin: models.User{ID: userID, Login: "alice", Password: hash}}
	svc = newUserService(t, users)

	tokens, err := svc.Login(context.Background(), "alice", "secret")
	require.NoError(t, err)
	assertTokenPair(t, tokens, userID)
}

func TestUserService_Login_Unauthorized(t *testing.T) {
	svc := newUserService(t, &fakeUserStore{})
	hash, err := svc.hashPassword("secret")
	require.NoError(t, err)

	tests := []struct {
		name     string
		users    *fakeUserStore
		login    string
		password string
	}{
		{name: "empty", users: &fakeUserStore{}, login: "", password: "secret"},
		{name: "unknown user", users: &fakeUserStore{loginErr: sql.ErrNoRows}, login: "alice", password: "secret"},
		{name: "wrong password", users: &fakeUserStore{byLogin: models.User{Password: hash}}, login: "alice", password: "nope"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newUserService(t, tc.users)
			_, err := svc.Login(context.Background(), tc.login, tc.password)
			kind := apperrors.KindUnauthorized
			if tc.login == "" {
				kind = apperrors.KindBadRequest
			}
			assertAppError(t, err, kind, InvalidCredentials)
		})
	}
}

func TestUserService_Refresh(t *testing.T) {
	userID := uuid.New()
	svc := newUserService(t, &fakeUserStore{byID: models.User{ID: userID}})
	refresh, err := common.Sign(userID, common.TokenTypeRefresh, time.Hour)
	require.NoError(t, err)

	tokens, err := svc.Refresh(context.Background(), refresh)
	require.NoError(t, err)
	assertTokenPair(t, tokens, userID)
}

func TestUserService_Refresh_InvalidToken(t *testing.T) {
	svc := newUserService(t, &fakeUserStore{})

	_, err := svc.Refresh(context.Background(), "not-a-token")
	assertAppError(t, err, apperrors.KindUnauthorized, InvalidRefreshToken)
}

func TestUserService_GetBalance(t *testing.T) {
	userID := uuid.New()
	stored := models.UserBalance{Current: 100.5, Withdrawn: 20}

	t.Run("ok", func(t *testing.T) {
		svc := newUserService(t, &fakeUserStore{balance: stored})
		got, err := svc.GetBalance(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, stored, got)
	})

	t.Run("not found", func(t *testing.T) {
		svc := newUserService(t, &fakeUserStore{balanceErr: sql.ErrNoRows})
		_, err := svc.GetBalance(context.Background(), userID)
		assertAppError(t, err, apperrors.KindNotFound, "user_not_found")
	})
}

func newUserService(t *testing.T, users userStore) *UserService {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret")
	return NewUserService(nil, zap.NewNop().Sugar(), users)
}

func assertTokenPair(t *testing.T, tokens TokenPair, userID uuid.UUID) {
	t.Helper()

	accessID, err := common.UserIDFromToken(tokens.AccessToken, common.TokenTypeAccess)
	require.NoError(t, err)
	refreshID, err := common.UserIDFromToken(tokens.RefreshToken, common.TokenTypeRefresh)
	require.NoError(t, err)
	assert.Equal(t, userID, accessID)
	assert.Equal(t, userID, refreshID)
}
