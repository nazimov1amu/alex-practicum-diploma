package services

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
	"go.uber.org/zap"
)

const (
	InternalServerError = "internal_server_error"
)

func mapRepositoryError(logger *zap.SugaredLogger, err error, message string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return apperrors.NotFound(message)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == resources.UniqueViolation {
		return apperrors.Conflict(message)
	}

	logger.Errorw("database", "error", err)
	return apperrors.Internal(InternalServerError)
}
