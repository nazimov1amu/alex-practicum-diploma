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
	NotFoundEvent       = "not_found_event"
	InsufficientBalance = "insufficient_balance"
)

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
		return models.Event{}, mapRepositoryError(s.logger, err, InternalServerError)
	}
	defer tx.Rollback()
	user, err := s.user.GetByID(ctx, event.UserID)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundUser)
	}

	delta := eventDelta(event)

	if user.Balance < delta {
		return models.Event{}, mapRepositoryError(s.logger, err, InsufficientBalance)
	}

	err = s.user.AddBalance(ctx, tx, event.UserID, delta)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundUser)
	}

	created, err := s.events.Create(ctx, tx, event)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundEvent)
	}

	if err := tx.Commit(); err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, InternalServerError)
	}
	return created, nil
}

func eventDelta(event models.Event) float64 {
	if event.EventType == models.EventTypeWithdrawal {
		return -event.Amount
	}
	return event.Amount
}

func (s *EventService) GetByID(ctx context.Context, id uuid.UUID) (models.Event, error) {
	event, err := s.events.GetByID(ctx, id)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundEvent)
	}
	return event, nil
}

func (s *EventService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Event, error) {
	events, err := s.events.ListByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(s.logger, err, NotFoundEvent)
	}
	return events, nil
}

func (s *EventService) Update(ctx context.Context, event models.Event) (models.Event, error) {
	oldEvent, err := s.events.GetByID(ctx, event.ID)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundEvent)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, InternalServerError)
	}
	defer tx.Rollback()

	user, err := s.user.GetByID(ctx, oldEvent.UserID)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundUser)
	}

	delta := eventDelta(event) - eventDelta(oldEvent)
	if user.Balance < delta {
		return models.Event{}, mapRepositoryError(s.logger, err, InsufficientBalance)
	}

	err = s.user.AddBalance(ctx, tx, oldEvent.UserID, delta)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundUser)
	}

	updated, err := s.events.Update(ctx, tx, event)
	if err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, NotFoundEvent)
	}

	if err := tx.Commit(); err != nil {
		return models.Event{}, mapRepositoryError(s.logger, err, InternalServerError)
	}
	return updated, nil
}

func (s *EventService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.events.Delete(ctx, id); err != nil {
		return mapRepositoryError(s.logger, err, NotFoundEvent)
	}
	return nil
}
