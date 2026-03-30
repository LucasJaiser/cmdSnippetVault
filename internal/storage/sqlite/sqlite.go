package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// New creates a new SQLiteRepository, initializing the database and running migrations.
func New(dbPath string) (*SQLiteRepository, error) {
	if dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
			return nil, fmt.Errorf("sqlite: create data directory: %w", err)
		}
	}

	// For tests we use :memory: as dbPath.
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not Open SQLite File: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not Ping Database: %w", err)
	}

	repo := &SQLiteRepository{
		db:      db,
		fileDSN: dsn,
		dsn:     fmt.Sprintf("sqlite://%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", dbPath),
	}

	if err := repo.Migrate(); err != nil {
		return nil, fmt.Errorf("could not migrate database: %w", err)
	}

	return repo, nil
}
