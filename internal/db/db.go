package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var Conn *sql.DB

// Init opens the database at the provided path and ensures schema exists.
func Init(path string) error {
	var err error
	Conn, err = sql.Open("sqlite3", path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS urls(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        code TEXT UNIQUE,
        original_url TEXT
    );`

	_, err = Conn.Exec(createTable)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
