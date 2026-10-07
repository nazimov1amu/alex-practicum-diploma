package repositories

import (
	"github.com/google/uuid"
)

func nullString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nullUUID(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	return &value
}

func nullFloat64(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}
