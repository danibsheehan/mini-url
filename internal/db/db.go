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
	_, err = Conn.Exec(`PRAGMA busy_timeout = 3000;`)
	if err != nil {
		return fmt.Errorf("set busy_timeout pragma: %w", err)
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS urls(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        code TEXT UNIQUE,
        original_url TEXT,
        click_count INTEGER NOT NULL DEFAULT 0
    );`

	_, err = Conn.Exec(createTable)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	_, err = Conn.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url_unique ON urls(original_url)`)
	if err != nil {
		return fmt.Errorf("migrate original_url unique index: %w", err)
	}

	var hasClickCount bool
	rows, err := Conn.Query(`PRAGMA table_info(urls);`)
	if err != nil {
		return fmt.Errorf("inspect schema: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("scan schema: %w", err)
		}
		if name == "click_count" {
			hasClickCount = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate schema: %w", err)
	}

	if !hasClickCount {
		_, err = Conn.Exec(`ALTER TABLE urls ADD COLUMN click_count INTEGER NOT NULL DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("add click_count column: %w", err)
		}
	}

	return nil
}
