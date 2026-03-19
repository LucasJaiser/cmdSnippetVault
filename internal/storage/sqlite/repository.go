package sqlite

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type SQLiteRepository struct {
	db       *sql.DB
	dsn      string
	file_dsn string
}

//go:embed migrations/*.sql
var migrations embed.FS

func (r *SQLiteRepository) Migrate() error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("Could not read migrations: %s", err.Error())
	}

	migrations_instance, err := migrate.NewWithSourceInstance("iofs", source, r.dsn)
	if err != nil {
		return fmt.Errorf("Could not create migrations instance: %s", err.Error())
	}

	err = migrations_instance.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Could not migrate up: %s", err.Error())
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("Could not close database: %s", err.Error())
	}

	return nil
}
