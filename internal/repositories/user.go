package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/models"
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

func (r *UserRepository) AddBalance(ctx context.Context, tx *sql.Tx, userID uuid.UUID, delta float64) error {
	const q = `
		UPDATE users
		SET balance = balance + $2
		WHERE id = $1
		  AND balance + $2 >= 0
		RETURNING id
	`

	var id uuid.UUID
	err := sqlscan.Get(ctx, tx, &id, q, userID, delta)
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
	err = sqlscan.Get(ctx, tx, &id, existsQ, userID)
	if err != nil {
		return err
	}
	return ErrInsufficientBalance
}

func (r *UserRepository) Update(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error) {
	const q = `
		WITH updated AS (
			UPDATE users
			SET login = COALESCE($2, login),
				password = COALESCE($3, password)
				balance = COALESCE($4, balance)
			WHERE id = $1
			RETURNING id, login, password, balance
		)
		SELECT id, login, password, balance
		FROM updated
	`

	var updated models.User
	err := sqlscan.Get(ctx, tx, &updated, q, user.ID, nullString(user.Login), nullString(user.Password), user.Balance)
	return updated, err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM users WHERE id = $1`

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
