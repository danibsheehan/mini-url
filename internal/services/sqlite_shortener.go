package services

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
	"mini-url/internal/db"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const (
	maxBusyRetries    = 5
	initialBusyBackoff = 5 * time.Millisecond
	maxBusyBackoff     = 80 * time.Millisecond
)

type codeGenFunc func(length int) (string, error)

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
	return &sqliteShortener{generateCode: generateCode}
}

func newSQLiteShortenerWithGenerator(gen codeGenFunc) Shortener {
	if gen == nil {
		gen = generateCode
	}
	return &sqliteShortener{generateCode: gen}
}

type sqliteShortener struct {
	generateCode codeGenFunc
}

func (s *sqliteShortener) Create(ctx context.Context, original string) (string, error) {
createLoop:
	for {
		code, genErr := s.generateCode(6)
		if genErr != nil {
			return "", genErr
		}

		for busyAttempt := 0; ; busyAttempt++ {
			_, err := db.Conn.ExecContext(ctx, "INSERT INTO urls(code, original_url) VALUES(?, ?)", code, original)
			if err != nil {
				if sqliteBusyOrLocked(err) {
					if busyAttempt >= maxBusyRetries {
						return "", fmt.Errorf("sqlite remained busy/locked after %d retries: %w", maxBusyRetries, err)
					}
					if err := sleepWithContext(ctx, busyRetryBackoff(busyAttempt)); err != nil {
						return "", err
					}
					continue
				}
				if sqliteUniqueOriginalViolation(err) {
					var existing string
					readErr := db.Conn.QueryRowContext(ctx, "SELECT code FROM urls WHERE original_url = ?", original).Scan(&existing)
					if readErr == nil {
						return existing, nil
					}
					if readErr == sql.ErrNoRows {
						continue createLoop
					}
					return "", readErr
				}
				// Retry only on generated code collisions.
				if sqliteUniqueCodeViolation(err) {
					continue createLoop
				}
				return "", err
			}
			return code, nil
		}
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

func sqliteBusyOrLocked(err error) bool {
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}
	return sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked
}

func sqliteUniqueCodeViolation(err error) bool {
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}
	if sqliteErr.Code != sqlite3.ErrConstraint {
		return false
	}
	return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique &&
		(sqliteErr.Error() == "UNIQUE constraint failed: urls.code")
}

func sqliteUniqueOriginalViolation(err error) bool {
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}
	if sqliteErr.Code != sqlite3.ErrConstraint {
		return false
	}
	return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique &&
		(sqliteErr.Error() == "UNIQUE constraint failed: urls.original_url")
}

func busyRetryBackoff(attempt int) time.Duration {
	backoff := initialBusyBackoff << attempt
	if backoff > maxBusyBackoff {
		return maxBusyBackoff
	}
	return backoff
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
