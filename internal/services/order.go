package services

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/repositories"
	"go.uber.org/zap"
)

const (
	NotFoundOrder = "not_found_order"
)

type OrderService struct {
	orders *repositories.OrderRepository
	logger *zap.SugaredLogger
}

func NewOrderService(db *sql.DB, logger *zap.SugaredLogger) *OrderService {
	return &OrderService{
		orders: repositories.NewOrderRepository(db),
		logger: logger,
	}
}

func (s *OrderService) Create(ctx context.Context, order models.Order) (models.Order, error) {
	created, err := s.orders.Create(ctx, order)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, NotFoundOrder)
	}
	return created, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (models.Order, error) {
	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, NotFoundOrder)
	}
	return order, nil
}

func (s *OrderService) GetByOrderNumber(ctx context.Context, orderNumber string) (models.Order, error) {
	order, err := s.orders.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, NotFoundOrder)
	}
	return order, nil
}

func (s *OrderService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	orders, err := s.orders.ListByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(s.logger, err, NotFoundOrder)
	}
	return orders, nil
}

func (s *OrderService) Update(ctx context.Context, order models.Order) (models.Order, error) {
	updated, err := s.orders.Update(ctx, order)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, NotFoundOrder)
	}
	return updated, nil
}

func (s *OrderService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.orders.Delete(ctx, id); err != nil {
		return mapRepositoryError(s.logger, err, NotFoundOrder)
	}
	return nil
}
