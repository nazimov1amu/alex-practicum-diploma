package services

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/repositories"
	"go.uber.org/zap"
)

type EventService struct {
	db     *sql.DB
	events *repositories.EventRepository
	user   *repositories.UserRepository
	order  *repositories.OrderRepository
	logger *zap.SugaredLogger
}

func NewEventService(db *sql.DB, logger *zap.SugaredLogger) *EventService {
	return &EventService{
		db:     db,
		events: repositories.NewEventRepository(db),
		user:   repositories.NewUserRepository(db),
		order:  repositories.NewOrderRepository(db),
		logger: logger,
	}
}

func (s *EventService) Create(ctx context.Context, event models.Event) (models.Event, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, "event")
	}
	defer tx.Rollback()
	user, err := s.user.GetByID(ctx, event.UserID)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, "user")
	}

	if event.EventType == models.EventTypeWithdrawal {
		order, err := s.order.GetByOrderNumber(ctx, event.OrderNumber)
		if err != nil {
			return models.Event{}, mapRepositoryError(s.logger, err, "order")
		}
		if order.UserID != user.ID {
			return models.Event{}, mapRepositoryError(s.logger, err, "order")
		}
	}

	delta := eventDelta(event)

	if user.Balance < delta {
		return models.Event{}, mapRepositoryError(s.logger, err, "user")
	}

	err = s.user.AddBalance(ctx, tx, event.UserID, delta)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, "user")
	}

	created, err := s.events.Create(ctx, tx, event)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, "event")
	}

	if err := tx.Commit(); err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, "event")
	}
	return created, nil
}

func (s *EventService) ListWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Event, error) {
	events, err := s.events.ListWithdrawalsByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(s.logger, err, "event")
	}
	return events, nil
}

func eventDelta(event models.Event) float64 {
	if event.EventType == models.EventTypeWithdrawal {
		return -event.Amount
	}
	return event.Amount
}
