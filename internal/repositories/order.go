package repositories

import (
	"context"
	"database/sql"

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
			RETURNING id, order_number, user_id, status, accrual
		)
		SELECT id, order_number, user_id, status, accrual
		FROM inserted
	`

	var created models.Order
	err := sqlscan.Get(ctx, r.db, &created, q, order.OrderNumber, order.UserID, order.Status, order.Accrual)
	return created, err
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Order, error) {
	const q = `
		SELECT id, order_number, user_id, status, accrual
		FROM orders
		WHERE id = $1
	`

	var order models.Order
	err := sqlscan.Get(ctx, r.db, &order, q, id)
	return order, err
}

func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (models.Order, error) {
	const q = `
		SELECT id, order_number, user_id, status, accrual
		FROM orders
		WHERE order_number = $1
	`

	var order models.Order
	err := sqlscan.Get(ctx, r.db, &order, q, orderNumber)
	return order, err
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	const q = `
		SELECT id, order_number, user_id, status, accrual
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var orders []models.Order
	err := sqlscan.Select(ctx, r.db, &orders, q, userID)
	return orders, err
}

func (r *OrderRepository) Update(ctx context.Context, order models.Order) (models.Order, error) {
	const q = `
		WITH updated AS (
			UPDATE orders
			SET order_number = COALESCE($2, order_number),
				user_id = COALESCE($3, user_id),
				status = COALESCE($4, status),
				accrual = $5
			WHERE id = $1
			RETURNING id, order_number, user_id, status, accrual
		)
		SELECT id, order_number, user_id, status, accrual
		FROM updated
	`

	var updated models.Order
	err := sqlscan.Get(
		ctx,
		r.db,
		&updated,
		q,
		order.ID,
		nullString(order.OrderNumber),
		nullUUID(order.UserID),
		nullString(order.Status),
		order.Accrual,
	)
	return updated, err
}

func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM orders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
