package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
	"github.com/shopspring/decimal"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO users (login, password)
			VALUES ($1, $2)
			RETURNING id, login, password
		)
		SELECT id, login, password
		FROM inserted
	`

	var created models.User
	err := sqlscan.Get(ctx, r.db, &created, q, user.Login, user.Password)
	return created, err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	const q = `
		SELECT id, login, password, balance
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := sqlscan.Get(ctx, r.db, &user, q, id)
	return user, err
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (models.User, error) {
	const q = `
		SELECT id, login, password, balance
		FROM users
		WHERE login = $1
	`

	var user models.User
	err := sqlscan.Get(ctx, r.db, &user, q, login)
	return user, err
}

func (r *UserRepository) GetBalance(ctx context.Context, userID uuid.UUID) (models.UserBalance, error) {
	const q = `
		SELECT
			u.balance AS current,
			COALESCE(SUM(e.amount) FILTER (WHERE e.event_type = 'withdrawal'), 0) AS withdrawn
		FROM users u
		LEFT JOIN events e ON e.user_id = u.id
		WHERE u.id = $1
		GROUP BY u.id, u.balance
	`

	var balance models.UserBalance
	err := sqlscan.Get(ctx, r.db, &balance, q, userID)
	return balance, err
}

func (r *UserRepository) AddBalance(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta decimal.Decimal) error {
	const q = `
		UPDATE users
		SET balance = balance + $2
		WHERE id = $1
		  AND balance + $2 >= 0
		RETURNING id
	`

	var id uuid.UUID
	err := sqlscan.Get(ctx, DBorTx(r.db, tx), &id, q, userID, delta)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	const existsQ = `
		SELECT id
		FROM users
		WHERE id = $1
	`
	err = sqlscan.Get(ctx, DBorTx(r.db, tx), &id, existsQ, userID)
	if err != nil {
		return err
	}
	return ErrInsufficientBalance
}
