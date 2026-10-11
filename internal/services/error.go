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
	notFoundPostfix = "_not_found"
	conflictPostfix = "_already_exists"
)

func mapRepositoryError(logger *zap.SugaredLogger, err error, entityName string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return apperrors.NotFound(entityName + notFoundPostfix)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == resources.UniqueViolation {
		return apperrors.Conflict(entityName + conflictPostfix)
	}

	logger.Errorw("database", "error", err)
	return apperrors.Internal("internal_server_error")
}
