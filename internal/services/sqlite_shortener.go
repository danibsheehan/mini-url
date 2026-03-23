package services

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"

	sqlite3 "github.com/mattn/go-sqlite3"
	"mini-url/internal/db"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// codeGenerator is overridden in tests to force collision/retry paths.
var codeGenerator = generateCode

func generateCode(length int) (string, error) {
	// Use crypto/rand for unpredictable short codes.
	b := make([]byte, length)
	for i := range b {
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("crypto random generation failed: %w", err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
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
		code, genErr := codeGenerator(6)
		if genErr != nil {
			return "", genErr
		}
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

	if _, err := db.Conn.ExecContext(ctx, "UPDATE urls SET click_count = click_count + 1 WHERE code = ?", code); err != nil {
		return "", err
	}

	return original, nil
}

func (s *sqliteShortener) Stats(ctx context.Context, code string) (URLStats, error) {
	var stats URLStats
	err := db.Conn.QueryRowContext(
		ctx,
		"SELECT original_url, click_count FROM urls WHERE code = ?",
		code,
	).Scan(&stats.OriginalURL, &stats.ClickCount)
	if err == sql.ErrNoRows {
		return URLStats{}, ErrNotFound
	}
	if err != nil {
		return URLStats{}, err
	}
	return stats, nil
}

func sqliteBusyOrConstraint(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint || sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked
	}
	return false
}
