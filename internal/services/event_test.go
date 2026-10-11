package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeEventStore struct {
	created models.Event
	called  bool
	list    []models.Event
	listErr error
}

func (f *fakeEventStore) Create(_ context.Context, _ *sql.Tx, event models.Event) (models.Event, error) {
	f.called = true
	if f.created.ID == uuid.Nil {
		f.created = event
	}
	return f.created, nil
}

func (f *fakeEventStore) ListWithdrawalsByUserID(context.Context, uuid.UUID) ([]models.Event, error) {
	return f.list, f.listErr
}

type fakeUserBalance struct {
	user  models.User
	added bool
}

func (f *fakeUserBalance) GetByID(context.Context, uuid.UUID) (models.User, error) {
	return f.user, nil
}

func (f *fakeUserBalance) AddBalance(_ context.Context, _ *sql.Tx, _ uuid.UUID, delta float64) error {
	f.user.Balance += delta
	return nil
}

func TestEventService_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		eventType   models.EventType
		amount      float64
		haveBalance float64
		wantBalance float64
	}{
		{name: "withdrawal", eventType: models.EventTypeWithdrawal, amount: 50, haveBalance: 50, wantBalance: 0},
		{name: "deposit", eventType: models.EventTypeDeposit, amount: 150.5, haveBalance: 0, wantBalance: 150.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := models.Event{
				UserID:      userID,
				OrderNumber: "2377222919",
				EventType:   tc.eventType,
				Amount:      tc.amount,
			}
			created := event
			created.ID = uuid.New()

			users := &fakeUserBalance{user: models.User{ID: userID, Balance: tc.haveBalance}}
			events := &fakeEventStore{created: created}
			svc, mock := newEventService(t, events, users)
			mock.ExpectBegin()
			mock.ExpectCommit()

			got, err := svc.Create(context.Background(), event)
			require.NoError(t, err)
			assert.Equal(t, created, got)
			assert.Equal(t, tc.wantBalance, users.user.Balance)
		})
	}
}

func TestEventService_ListWithdrawals(t *testing.T) {
	userID := uuid.New()
	stored := []models.Event{{
		ID:          uuid.New(),
		UserID:      userID,
		OrderNumber: "2377222919",
		EventType:   models.EventTypeWithdrawal,
		Amount:      50,
	}}

	t.Run("ok", func(t *testing.T) {
		events := &fakeEventStore{list: stored}
		svc, _ := newEventService(t, events, &fakeUserBalance{})

		got, err := svc.ListWithdrawals(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, stored, got)
	})

	t.Run("repository error", func(t *testing.T) {
		events := &fakeEventStore{listErr: sql.ErrNoRows}
		svc, _ := newEventService(t, events, &fakeUserBalance{})

		_, err := svc.ListWithdrawals(context.Background(), userID)
		assertAppError(t, err, apperrors.KindNotFound, "event_not_found")
	})
}

func newEventService(t *testing.T, events eventStore, users userBalanceStore) (*EventService, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		mock.ExpectClose()
		assert.NoError(t, db.Close())
	})

	return NewEventService(db, zap.NewNop().Sugar(), events, users), mock
}
