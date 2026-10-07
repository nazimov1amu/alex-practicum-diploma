package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/resources"
)

func mapAppError(err error) error {
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		return echo.NewHTTPError(http.StatusInternalServerError, resources.InternalServerError)
	}

	status := http.StatusInternalServerError
	switch appErr.Kind {
	case apperrors.KindNotFound:
		status = http.StatusNotFound
	case apperrors.KindBadRequest:
		status = http.StatusBadRequest
	case apperrors.KindUnauthorized:
		status = http.StatusUnauthorized
	case apperrors.KindForbidden:
		status = http.StatusForbidden
	case apperrors.KindConflict:
		status = http.StatusConflict
	case apperrors.KindInternal:
		status = http.StatusInternalServerError
	}

	message := appErr.Message
	if message == "" {
		message = resources.InternalServerError
	}
	return echo.NewHTTPError(status, message)
}
