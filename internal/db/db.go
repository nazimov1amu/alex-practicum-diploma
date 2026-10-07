package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nazimov1amu/alex-practicum-diploma/internal/migrations"
)

func NewDB(databaseURL string) (*sql.DB, error) {
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := RunMigrations(databaseURL); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return database, nil
}

func RunMigrations(databaseURL string) error {
	migrationURL, err := withSimpleProtocol(databaseURL)
	if err != nil {
		return err
	}

	// Separate pool: migrate closes the *sql.DB it owns.
	// simple_protocol keeps the plpgsql function body in one query.
	migrationDB, err := sql.Open("pgx", migrationURL)
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		_ = migrationDB.Close()
		return fmt.Errorf("migration source: %w", err)
	}

	driver, err := postgres.WithInstance(migrationDB, &postgres.Config{})
	if err != nil {
		_ = migrationDB.Close()
		return fmt.Errorf("migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		_ = driver.Close()
		return fmt.Errorf("migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		_, _ = m.Close()
		return fmt.Errorf("migrate up: %w", err)
	}

	if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
		return fmt.Errorf("close migrate: source: %v, database: %v", sourceErr, dbErr)
	}

	return nil
}

func withSimpleProtocol(databaseURL string) (string, error) {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parse database url: %w", err)
	}

	query := parsed.Query()
	if query.Get("default_query_exec_mode") == "" {
		query.Set("default_query_exec_mode", "simple_protocol")
	}
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
