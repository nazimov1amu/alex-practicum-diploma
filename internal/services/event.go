package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/repositories"
	"go.uber.org/zap"
)

const insufficientBalance = "insufficient_balance"

type EventService struct {
	db     *sql.DB
	events *repositories.EventRepository
	user   *repositories.UserRepository
	logger *zap.SugaredLogger
}

func NewEventService(db *sql.DB, logger *zap.SugaredLogger) *EventService {
	return &EventService{
		db:     db,
		events: repositories.NewEventRepository(db),
		user:   repositories.NewUserRepository(db),
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

	if event.EventType == models.EventTypeWithdrawal && user.Balance < event.Amount {
		return models.Event{}, apperrors.PaymentRequired(insufficientBalance)
	}

	err = s.user.AddBalance(ctx, tx, event.UserID, eventDelta(event))
	if errors.Is(err, repositories.ErrInsufficientBalance) {
		return models.Event{}, apperrors.PaymentRequired(insufficientBalance)
	}
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
