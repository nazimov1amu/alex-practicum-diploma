package repositories

import (
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
)

func nullString(value string) *string {
	if value == "" {
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

func DBorTx(db *sql.DB, tx *sql.Tx) sqlscan.Querier {
	if tx != nil {
		return tx
	}
	return db
}
