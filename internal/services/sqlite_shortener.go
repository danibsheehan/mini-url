package services

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"mini-url/internal/db"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// codeGenerator is overridden in tests to force collision/retry paths.
var codeGenerator = generateCode

func generateCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// NewSQLiteShortener returns a Shortener backed by the package-level DB connection.
func NewSQLiteShortener() Shortener {
	return &sqliteShortener{}
}

type sqliteShortener struct{}

func (s *sqliteShortener) Create(ctx context.Context, original string) (string, error) {
	var existing string
	err := db.Conn.QueryRowContext(ctx, "SELECT code FROM urls WHERE original_url = ?", original).Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if err != sql.ErrNoRows && err != nil {
		return "", err
	}

	for {
		code := codeGenerator(6)
		_, err := db.Conn.ExecContext(ctx, "INSERT INTO urls(code, original_url) VALUES(?, ?)", code, original)
		if err != nil {
			// if constraint violation, try again
			if sqliteBusyOrConstraint(err) {
				continue
			}
			return "", err
		}
		return code, nil
	}
}

func (s *sqliteShortener) Resolve(ctx context.Context, code string) (string, error) {
	var original string
	err := db.Conn.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE code = ?", code).Scan(&original)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return original, nil
}

func sqliteBusyOrConstraint(err error) bool {
	if err == nil {
		return false
	}
	// best-effort: treat any sqlite constraint or busy error as retryable
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	return true
}
