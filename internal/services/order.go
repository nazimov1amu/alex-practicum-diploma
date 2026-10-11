package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"go.uber.org/zap"
)

const (
	InvalidOrderNumber = "invalid_order_number"
	OrderAlreadyExists = "order_already_exists"
)

type OrderService struct {
	orders orderStore
	logger *zap.SugaredLogger
}

type orderStore interface {
	Create(ctx context.Context, order models.Order) (models.Order, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (models.Order, error)
}

func NewOrderService(db *sql.DB, logger *zap.SugaredLogger, orders orderStore) *OrderService {
	return &OrderService{
		orders: orders,
		logger: logger,
	}
}

func (s *OrderService) Create(ctx context.Context, order models.Order) (models.Order, error) {
	existing, err := s.orders.GetByOrderNumber(ctx, order.OrderNumber)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.Order{}, mapRepositoryError(s.logger, err, "order")
	}
	if err == nil && existing.UserID != order.UserID {
		return models.Order{}, apperrors.Conflict(OrderAlreadyExists)
	}
	if err == nil {
		return existing, nil
	}

	order.Status = models.OrderStatusRegistered
	created, err := s.orders.Create(ctx, order)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, "order")
	}
	return created, nil
}

func (s *OrderService) List(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	orders, err := s.orders.ListByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(s.logger, err, "order")
	}
	return orders, nil
}

func (s *OrderService) GetByOrderNumber(ctx context.Context, orderNumber string) (models.Order, error) {
	order, err := s.orders.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, "order")
	}
	return order, nil
}
