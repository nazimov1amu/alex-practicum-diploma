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
