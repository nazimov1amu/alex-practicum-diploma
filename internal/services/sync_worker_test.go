package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSyncWorker_tickProcessed(t *testing.T) {
	userID := uuid.New()
	order := models.Order{UserID: userID, OrderNumber: "12345678903", Status: models.OrderStatusNew}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/12345678903", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(accrualResult{
			Order:   order.OrderNumber,
			Status:  models.OrderStatusProcessed,
			Accrual: 500,
		}))
	}))
	t.Cleanup(server.Close)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		assert.NoError(t, db.Close())
	})
	mock.ExpectBegin()
	mock.ExpectCommit()

	orders := &fakeAccrualOrders{pending: []models.Order{order}, userID: userID}
	users := &fakeUserBalance{user: models.User{ID: userID}}
	worker := NewSyncWorker(server.URL, db, orders, users, zap.NewNop().Sugar())

	worker.tick(context.Background())

	require.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, models.OrderStatusProcessed, orders.status)
	assert.Equal(t, 500.0, orders.accrual)
	assert.Equal(t, 500.0, users.user.Balance)
}

func TestSyncWorker_tickNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		assert.NoError(t, db.Close())
	})

	orders := &fakeAccrualOrders{pending: []models.Order{{OrderNumber: "12345678903"}}}
	worker := NewSyncWorker(server.URL, db, orders, &fakeUserBalance{}, zap.NewNop().Sugar())

	worker.tick(context.Background())

	require.NoError(t, mock.ExpectationsWereMet())
	assert.Empty(t, orders.status)
}

func TestMapAccrualStatus(t *testing.T) {
	assert.Equal(t, models.OrderStatusProcessing, mapAccrualStatus(models.OrderStatusRegistered))
	assert.Equal(t, models.OrderStatusProcessing, mapAccrualStatus(models.OrderStatusProcessing))
	assert.Equal(t, models.OrderStatusProcessed, mapAccrualStatus(models.OrderStatusProcessed))
	assert.Equal(t, models.OrderStatusInvalid, mapAccrualStatus(models.OrderStatusInvalid))
	assert.Empty(t, mapAccrualStatus("UNKNOWN"))
}

type fakeAccrualOrders struct {
	pending []models.Order
	userID  uuid.UUID
	status  string
	accrual float64
}

func (f *fakeAccrualOrders) ListPending(context.Context) ([]models.Order, error) {
	return f.pending, nil
}

func (f *fakeAccrualOrders) ApplyAccrual(_ context.Context, _ *sql.Tx, _ string, status string, accrual float64) (uuid.UUID, bool, error) {
	f.status = status
	f.accrual = accrual
	return f.userID, true, nil
}
