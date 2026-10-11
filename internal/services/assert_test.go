package services

import (
	"testing"

	"github.com/nazimov1amu/alex-practicum-diploma/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertAppError(t *testing.T, err error, kind apperrors.Kind, message string) {
	t.Helper()

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, kind, appErr.Kind)
	assert.Equal(t, message, appErr.Message)
}
