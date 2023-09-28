package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var fs embed.FS

func MigrateDb(postgresURI string) error {
	db, err := sql.Open("postgres", postgresURI)
	if err != nil {
		return err
	}
	_, err = db.Exec("create schema if not exists twa")
	if err != nil {
		return fmt.Errorf("failed to create 'twa' schema: %w", err)
	}
	databaseInstance, err := postgres.WithInstance(db, &postgres.Config{SchemaName: "twa"})
	if err != nil {
		return fmt.Errorf("failed to init database instance: %w", err)
	}
	sourceDriver, err := iofs.New(fs, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", databaseInstance)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}
