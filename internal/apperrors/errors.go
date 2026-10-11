package apperrors

import "errors"

type Kind int

const (
	KindNotFound Kind = iota
	KindBadRequest
	KindUnauthorized
	KindForbidden
	KindConflict
	KindInternal
	KindValidationError
	KindPaymentRequired
)

const (
	InvalidAccessToken  = "invalid_access_token"
	InternalServerError = "internal_server_error"
)

type AppError struct {
	Kind    Kind
	Message string
}

func (e *AppError) Error() string {
	if e.Message == "" {
		return "app_error"
	}
	return e.Message
}

func (e *AppError) Is(target error) bool {
	var other *AppError
	if !errors.As(target, &other) {
		return false
	}
	return e.Kind == other.Kind
}

var (
	ErrNotFound        = &AppError{Kind: KindNotFound}
	ErrBadRequest      = &AppError{Kind: KindBadRequest}
	ErrUnauthorized    = &AppError{Kind: KindUnauthorized}
	ErrForbidden       = &AppError{Kind: KindForbidden}
	ErrConflict        = &AppError{Kind: KindConflict}
	ErrInternal        = &AppError{Kind: KindInternal}
	ErrValidationError = &AppError{Kind: KindValidationError}
	ErrPaymentRequired = &AppError{Kind: KindPaymentRequired}
)

func NotFound(message string) error {
	return &AppError{Kind: KindNotFound, Message: message}
}

func BadRequest(message string) error {
	return &AppError{Kind: KindBadRequest, Message: message}
}

func Unauthorized(message string) error {
	return &AppError{Kind: KindUnauthorized, Message: message}
}

func Forbidden(message string) error {
	return &AppError{Kind: KindForbidden, Message: message}
}

func Conflict(message string) error {
	return &AppError{Kind: KindConflict, Message: message}
}

func ValidationError(message string) error {
	return &AppError{Kind: KindValidationError, Message: message}
}

func PaymentRequired(message string) error {
	return &AppError{Kind: KindPaymentRequired, Message: message}
}

func Internal(message string) error {
	return &AppError{Kind: KindInternal, Message: message}
}
