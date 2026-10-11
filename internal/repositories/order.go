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
