package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeOrderStore struct {
	existing models.Order
	getErr   error
	created  models.Order
	stored   models.Order
	list     []models.Order
	listErr  error
}

func (f *fakeOrderStore) GetByOrderNumber(context.Context, string) (models.Order, error) {
	if f.getErr != nil {
		return models.Order{}, f.getErr
	}
	return f.existing, nil
}

func (f *fakeOrderStore) Create(_ context.Context, order models.Order) (models.Order, error) {
	f.stored = order
	return f.created, nil
}

func (f *fakeOrderStore) ListByUserID(context.Context, uuid.UUID) ([]models.Order, error) {
	return f.list, f.listErr
}

func TestOrderService_Create(t *testing.T) {
	userID := uuid.New()
	order := models.Order{UserID: userID, OrderNumber: "2377222919"}
	created := order
	created.ID = uuid.New()
	created.Status = models.OrderStatusRegistered
	existing := models.Order{ID: uuid.New(), UserID: userID, OrderNumber: order.OrderNumber}
	foreign := models.Order{ID: uuid.New(), UserID: uuid.New(), OrderNumber: order.OrderNumber}

	tests := []struct {
		name     string
		store    fakeOrderStore
		want     models.Order
		wantKind apperrors.Kind
		wantMsg  string
		stored   bool
	}{
		{
			name:   "new",
			store:  fakeOrderStore{getErr: sql.ErrNoRows, created: created},
			want:   created,
			stored: true,
		},
		{
			name:  "same user",
			store: fakeOrderStore{existing: existing},
			want:  existing,
		},
		{
			name:     "other user",
			store:    fakeOrderStore{existing: foreign},
			wantKind: apperrors.KindConflict,
			wantMsg:  OrderAlreadyExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := tc.store
			svc := NewOrderService(nil, zap.NewNop().Sugar(), &store)

			got, err := svc.Create(context.Background(), order)
			if tc.wantMsg != "" {
				assertAppError(t, err, tc.wantKind, tc.wantMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			if tc.stored {
				assert.Equal(t, models.OrderStatusRegistered, store.stored.Status)
			} else {
				assert.Zero(t, store.stored)
			}
		})
	}
}

func TestOrderService_List(t *testing.T) {
	userID := uuid.New()
	stored := []models.Order{{
		ID:          uuid.New(),
		UserID:      userID,
		OrderNumber: "2377222919",
		Status:      models.OrderStatusProcessed,
	}}

	t.Run("ok", func(t *testing.T) {
		svc := NewOrderService(nil, zap.NewNop().Sugar(), &fakeOrderStore{list: stored})
		got, err := svc.List(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, stored, got)
	})

	t.Run("repository error", func(t *testing.T) {
		svc := NewOrderService(nil, zap.NewNop().Sugar(), &fakeOrderStore{listErr: sql.ErrNoRows})
		_, err := svc.List(context.Background(), userID)
		assertAppError(t, err, apperrors.KindNotFound, "order_not_found")
	})
}
