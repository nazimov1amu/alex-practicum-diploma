package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order models.Order) (models.Order, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO orders (order_number, user_id, status, accrual)
			VALUES ($1, $2, $3, $4)
			RETURNING id, order_number, user_id, status, accrual, created_at
		)
		SELECT id, order_number, user_id, status, accrual, created_at
		FROM inserted
	`

	var created models.Order
	err := sqlscan.Get(ctx, r.db, &created, q, order.OrderNumber, order.UserID, order.Status, order.Accrual)
	return created, err
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	const q = `
		SELECT id, order_number, user_id, status, accrual, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var orders []models.Order
	err := sqlscan.Select(ctx, r.db, &orders, q, userID)
	return orders, err
}

func (r *OrderRepository) ListPending(ctx context.Context) ([]models.Order, error) {
	const q = `
		SELECT id, order_number, user_id, status, accrual, created_at
		FROM orders
		WHERE status IN ('NEW', 'REGISTERED', 'PROCESSING')
		ORDER BY created_at
	`

	var orders []models.Order
	err := sqlscan.Select(ctx, r.db, &orders, q)
	return orders, err
}

func (r *OrderRepository) ApplyAccrual(ctx context.Context, tx *sql.Tx, number, status string, accrual float64) (uuid.UUID, bool, error) {
	const q = `
		UPDATE orders
		SET status = $2,
		    accrual = CASE WHEN $2 IN ('PROCESSED', 'INVALID') THEN $3 ELSE accrual END,
		    updated_at = NOW()
		WHERE order_number = $1
		  AND status NOT IN ('PROCESSED', 'INVALID')
		RETURNING user_id
	`

	var userID uuid.UUID
	err := sqlscan.Get(ctx, DBorTx(r.db, tx), &userID, q, number, status, accrual)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return userID, err == nil, err
}

func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (models.Order, error) {
	const q = `
		SELECT id, order_number, user_id, status, accrual, created_at
		FROM orders
		WHERE order_number = $1
	`

	var order models.Order
	err := sqlscan.Get(ctx, r.db, &order, q, orderNumber)
	return order, err
}
