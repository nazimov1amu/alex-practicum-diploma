package services

import (
	"context"
	"database/sql"

	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/repositories"
	"go.uber.org/zap"
)

const (
	InvalidOrderNumber = "invalid_order_number"
	OrderAlreadyExists = "order_already_exists"
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
	existing, _ := s.orders.GetByOrderNumber(ctx, order.OrderNumber)
	if existing.UserID != order.UserID {
		return models.Order{}, apperrors.Conflict(OrderAlreadyExists)
	}

	if !luhn(order.OrderNumber) {
		return models.Order{}, apperrors.ValidationError(InvalidOrderNumber)
	}

	created, err := s.orders.Create(ctx, order)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, "order")
	}
	return created, nil
}

func (s *OrderService) GetByOrderNumber(ctx context.Context, orderNumber string) (models.Order, error) {
	order, err := s.orders.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return models.Order{}, mapRepositoryError(s.logger, err, "order")
	}
	return order, nil
}

func luhn(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := number[i]
		if digit < '0' || digit > '9' {
			return false
		}

		n := int(digit - '0')
		if double {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}
	return sum%10 == 0
}
