package repositories

import (
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
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

func DBorTx(db *sql.DB, tx *sql.Tx) sqlscan.Querier {
	if tx != nil {
		return tx
	}
	return db
}
