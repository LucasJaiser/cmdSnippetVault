package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func New(db_path string) (*SQLiteRepository, error) {
	if db_path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(db_path), 0o700); err != nil {
			return nil, fmt.Errorf("sqlite: create data directory: %w", err)
		}
	}

	//For Tests we use :memory: as db_path
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", db_path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not Open SQLite File: %w", err)
	}

	if err := db.Ping(); err != nil {

		return nil, fmt.Errorf("could not Ping Database: %w", err)
	}

	repo := &SQLiteRepository{
		db:       db,
		file_dsn: dsn,
		dsn:      fmt.Sprintf("sqlite://%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", db_path),
	}

	if err := repo.Migrate(); err != nil {
		return nil, fmt.Errorf("could not migrate database: %w", err)
	}

	return repo, nil
}
