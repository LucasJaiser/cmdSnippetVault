package sqlite

import "database/sql"

type SQLiteRepository struct {
	db *sql.DB
}
