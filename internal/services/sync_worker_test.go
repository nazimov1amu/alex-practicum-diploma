package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSyncWorker_tickProcessed(t *testing.T) {
	userID := uuid.New()
	order := models.Order{ID: uuid.New(), UserID: userID, OrderNumber: "12345678903", Status: models.OrderStatusNew}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/12345678903", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(accrualResult{
			Order:   order.OrderNumber,
			Status:  models.OrderStatusProcessed,
			Accrual: decimal.NewFromInt(500),
		}))
	}))
	t.Cleanup(server.Close)

	store := &fakeOrderStore{pending: []models.Order{order}}
	users := &fakeUserBalance{user: models.User{ID: userID}}
	events := &fakeEventStore{}
	eventSvc, mock := newEventService(t, events, users)
	mock.ExpectBegin()
	mock.ExpectCommit()

	worker := NewSyncWorker(server.URL, NewOrderService(nil, zap.NewNop().Sugar(), store), eventSvc, zap.NewNop().Sugar())
	worker.tick(context.Background())

	assert.Equal(t, models.OrderStatusProcessed, store.updated.Status)
	assert.True(t, decimal.NewFromInt(500).Equal(store.updated.Accrual))
	assert.Equal(t, models.Event{
		UserID:      userID,
		OrderNumber: order.OrderNumber,
		EventType:   models.EventTypeDeposit,
		Amount:      decimal.NewFromInt(500),
	}, events.created)
	assert.True(t, decimal.NewFromInt(500).Equal(users.user.Balance))
}

func TestSyncWorker_tickNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	store := &fakeOrderStore{pending: []models.Order{{OrderNumber: "12345678903"}}}
	eventSvc, _ := newEventService(t, &fakeEventStore{}, &fakeUserBalance{})
	worker := NewSyncWorker(server.URL, NewOrderService(nil, zap.NewNop().Sugar(), store), eventSvc, zap.NewNop().Sugar())

	worker.tick(context.Background())

	assert.Zero(t, store.updated)
}

func TestSyncWorker_tickFinishesOrderAfterCancel(t *testing.T) {
	userID := uuid.New()
	order := models.Order{ID: uuid.New(), UserID: userID, OrderNumber: "12345678903", Status: models.OrderStatusNew}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cancel()
		require.NoError(t, json.NewEncoder(w).Encode(accrualResult{
			Order:   order.OrderNumber,
			Status:  models.OrderStatusProcessed,
			Accrual: decimal.NewFromInt(500),
		}))
	}))
	t.Cleanup(server.Close)

	store := &fakeOrderStore{pending: []models.Order{order}}
	users := &fakeUserBalance{user: models.User{ID: userID}}
	events := &fakeEventStore{}
	eventSvc, mock := newEventService(t, events, users)
	mock.ExpectBegin()
	mock.ExpectCommit()

	worker := NewSyncWorker(server.URL, NewOrderService(nil, zap.NewNop().Sugar(), store), eventSvc, zap.NewNop().Sugar())
	worker.tick(ctx)

	assert.Equal(t, models.OrderStatusProcessed, store.updated.Status)
	assert.True(t, decimal.NewFromInt(500).Equal(events.created.Amount))
}

func TestSyncWorker_RunClosesDone(t *testing.T) {
	worker := NewSyncWorker("", nil, nil, zap.NewNop().Sugar())
	go worker.Run(context.Background())

	select {
	case <-worker.Done():
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}

func TestMapAccrualStatus(t *testing.T) {
	assert.Equal(t, models.OrderStatusProcessing, mapAccrualStatus(models.OrderStatusRegistered))
	assert.Equal(t, models.OrderStatusProcessing, mapAccrualStatus(models.OrderStatusProcessing))
	assert.Equal(t, models.OrderStatusProcessed, mapAccrualStatus(models.OrderStatusProcessed))
	assert.Equal(t, models.OrderStatusInvalid, mapAccrualStatus(models.OrderStatusInvalid))
	assert.Empty(t, mapAccrualStatus("UNKNOWN"))
}
