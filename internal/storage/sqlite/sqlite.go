package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func New(db_path string) (*SQLiteRepository, error) {
	//For Tests we use :memory: as db_path
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", db_path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not Open SQLite File: %s", err.Error())
	}

	if err := db.Ping(); err != nil {

		return nil, fmt.Errorf("Could not Ping Database: %s", err.Error())
	}

	repo := &SQLiteRepository{
		db:       db,
		file_dsn: dsn,
		dsn:      fmt.Sprintf("sqlite://%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", db_path),
	}

	if err := repo.Migrate(); err != nil {
		return nil, fmt.Errorf("Could not migrate database: %s", err.Error())
	}

	return repo, nil
}
