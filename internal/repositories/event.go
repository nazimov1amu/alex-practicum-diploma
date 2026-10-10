package repositories

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, tx *sql.Tx, event models.Event) (models.Event, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO events (user_id, event_type, amount)
			VALUES ($1, $2, $3)
			RETURNING id, user_id, event_type, amount
		)
		SELECT id, user_id, event_type, amount
		FROM inserted
	`

	var created models.Event
	err := sqlscan.Get(ctx, DBorTx(r.db, tx), &created, q, event.UserID, event.EventType, event.Amount)
	return created, err
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Event, error) {
	const q = `
		SELECT id, user_id, event_type, amount
		FROM events
		WHERE id = $1
	`

	var event models.Event
	err := sqlscan.Get(ctx, r.db, &event, q, id)
	return event, err
}

func (r *EventRepository) ListWithdrawalsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Event, error) {
	const q = `
		SELECT id, user_id, order_id, event_type, amount
		FROM events
		WHERE user_id = $1 AND event_type = 'withdraw'
		ORDER BY created_at DESC
	`

	var events []models.Event
	err := sqlscan.Select(ctx, r.db, &events, q, userID)
	return events, err
}
