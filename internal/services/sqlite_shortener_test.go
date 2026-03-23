package services

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
	"mini-url/internal/db"
)

func TestSQLiteShortener_CreateAndResolve(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()
	db.Conn.SetMaxOpenConns(1)
	db.Conn.SetMaxIdleConns(1)

	svc := NewSQLiteShortener()
	ctx := context.Background()

	orig := "https://example.com/foo"
	code, err := svc.Create(ctx, orig)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if code == "" {
		t.Fatalf("expected non-empty code")
	}

	// creating again should return same code
	code2, err := svc.Create(ctx, orig)
	if err != nil {
		t.Fatalf("create(second): %v", err)
	}
	if code2 != code {
		t.Fatalf("expected same code on duplicate create: got %s want %s", code2, code)
	}

	// resolve
	got, err := svc.Resolve(ctx, code)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != orig {
		t.Fatalf("resolve returned %s, want %s", got, orig)
	}

	var clickCount int
	if err := db.Conn.QueryRowContext(ctx, "SELECT click_count FROM urls WHERE code = ?", code).Scan(&clickCount); err != nil {
		t.Fatalf("read click_count: %v", err)
	}
	if clickCount != 1 {
		t.Fatalf("click_count = %d, want %d", clickCount, 1)
	}
}

func TestSQLiteShortener_ResolveNotFound(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

	svc := NewSQLiteShortener()
	ctx := context.Background()

	_, err := svc.Resolve(ctx, "no_such_code")
	if err == nil {
		t.Fatalf("expected error for missing code")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteShortener_Stats(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

	svc := NewSQLiteShortener()
	ctx := context.Background()

	orig := "https://example.com/stats"
	code, err := svc.Create(ctx, orig)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Two resolves should produce click_count = 2.
	if _, err := svc.Resolve(ctx, code); err != nil {
		t.Fatalf("resolve(1): %v", err)
	}
	if _, err := svc.Resolve(ctx, code); err != nil {
		t.Fatalf("resolve(2): %v", err)
	}

	stats, err := svc.Stats(ctx, code)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.OriginalURL != orig {
		t.Fatalf("original_url = %q, want %q", stats.OriginalURL, orig)
	}
	if stats.ClickCount != 2 {
		t.Fatalf("click_count = %d, want %d", stats.ClickCount, 2)
	}
}

func TestSQLiteShortener_StatsNotFound(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

	svc := NewSQLiteShortener()
	_, err := svc.Stats(context.Background(), "no_such_code")
	if err == nil {
		t.Fatalf("expected error for missing code")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGenerateCode(t *testing.T) {
	t.Run("returns requested length", func(t *testing.T) {
		got, err := generateCode(6)
		if err != nil {
			t.Fatalf("generateCode returned error: %v", err)
		}
		if len(got) != 6 {
			t.Fatalf("len(code) = %d, want %d", len(got), 6)
		}
	})

	t.Run("uses only allowed charset", func(t *testing.T) {
		got, err := generateCode(64)
		if err != nil {
			t.Fatalf("generateCode returned error: %v", err)
		}
		for _, ch := range got {
			if !strings.ContainsRune(charset, ch) {
				t.Fatalf("unexpected character in code: %q", ch)
			}
		}
	})

	t.Run("returns empty for zero length", func(t *testing.T) {
		got, err := generateCode(0)
		if err != nil {
			t.Fatalf("generateCode returned error: %v", err)
		}
		if got != "" {
			t.Fatalf("code = %q, want empty string", got)
		}
	})
}

func TestSQLiteBusyOrConstraint(t *testing.T) {
	t.Run("nil error is not retryable", func(t *testing.T) {
		if sqliteBusyOrConstraint(nil) {
			t.Fatalf("expected false for nil error")
		}
	})

	t.Run("sql err no rows is not retryable", func(t *testing.T) {
		if sqliteBusyOrConstraint(sql.ErrNoRows) {
			t.Fatalf("expected false for sql.ErrNoRows")
		}
	})

	t.Run("generic non nil error is not retryable", func(t *testing.T) {
		if sqliteBusyOrConstraint(errors.New("constraint failed")) {
			t.Fatalf("expected false for generic error")
		}
	})

	t.Run("sqlite constraint error is retryable", func(t *testing.T) {
		if !sqliteBusyOrConstraint(sqlite3.Error{Code: sqlite3.ErrConstraint}) {
			t.Fatalf("expected true for sqlite constraint error")
		}
	})

	t.Run("sqlite busy error is retryable", func(t *testing.T) {
		if !sqliteBusyOrConstraint(sqlite3.Error{Code: sqlite3.ErrBusy}) {
			t.Fatalf("expected true for sqlite busy error")
		}
	})

	t.Run("sqlite locked error is retryable", func(t *testing.T) {
		if !sqliteBusyOrConstraint(sqlite3.Error{Code: sqlite3.ErrLocked}) {
			t.Fatalf("expected true for sqlite locked error")
		}
	})
}

func TestSQLiteBusyOrLocked(t *testing.T) {
	t.Run("busy is retryable", func(t *testing.T) {
		if !sqliteBusyOrLocked(sqlite3.Error{Code: sqlite3.ErrBusy}) {
			t.Fatalf("expected true for sqlite busy")
		}
	})

	t.Run("locked is retryable", func(t *testing.T) {
		if !sqliteBusyOrLocked(sqlite3.Error{Code: sqlite3.ErrLocked}) {
			t.Fatalf("expected true for sqlite locked")
		}
	})

	t.Run("constraint is not retryable here", func(t *testing.T) {
		if sqliteBusyOrLocked(sqlite3.Error{Code: sqlite3.ErrConstraint}) {
			t.Fatalf("expected false for sqlite constraint")
		}
	})
}

func TestBusyRetryBackoff(t *testing.T) {
	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{name: "attempt 0", attempt: 0, want: 5 * time.Millisecond},
		{name: "attempt 1", attempt: 1, want: 10 * time.Millisecond},
		{name: "attempt 2", attempt: 2, want: 20 * time.Millisecond},
		{name: "attempt 3", attempt: 3, want: 40 * time.Millisecond},
		{name: "attempt 4", attempt: 4, want: 80 * time.Millisecond},
		{name: "attempt capped", attempt: 8, want: 80 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := busyRetryBackoff(tt.attempt)
			if got != tt.want {
				t.Fatalf("busyRetryBackoff(%d) = %s, want %s", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestSleepWithContext(t *testing.T) {
	t.Run("returns context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := sleepWithContext(ctx, 20*time.Millisecond)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("sleeps and returns nil", func(t *testing.T) {
		err := sleepWithContext(context.Background(), 1*time.Millisecond)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

func TestSQLiteShortener_CreateRetriesOnCollision(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

	// First generated code collides with an existing row, second should succeed.
	calls := 0
	gen := func(length int) (string, error) {
		calls++
		if calls == 1 {
			return "ABC123", nil
		}
		return "XYZ789", nil
	}

	_, err := db.Conn.Exec(`INSERT INTO urls(code, original_url) VALUES(?, ?)`, "ABC123", "https://already-exists.example")
	if err != nil {
		t.Fatalf("seed insert: %v", err)
	}

	svc := newSQLiteShortenerWithGenerator(gen)
	got, err := svc.Create(context.Background(), "https://new-url.example")
	if err != nil {
		t.Fatalf("create after collision: %v", err)
	}
	if got != "XYZ789" {
		t.Fatalf("code = %q, want %q", got, "XYZ789")
	}
	if calls < 2 {
		t.Fatalf("expected at least 2 generation attempts, got %d", calls)
	}
}

func TestSQLiteShortener_CreateReturnsCodeGenerationError(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

	wantErr := errors.New("rng unavailable")
	gen := func(length int) (string, error) {
		return "", wantErr
	}

	svc := newSQLiteShortenerWithGenerator(gen)
	_, err := svc.Create(context.Background(), "https://new-url.example")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestSQLiteShortener_CreateConcurrentSameOriginalReturnsSameCode(t *testing.T) {
	tmp, err := os.CreateTemp("", "mini-url-concurrency-*.db")
	if err != nil {
		t.Fatalf("create temp db file: %v", err)
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		t.Fatalf("close temp db file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmpPath)
	})

	if err := db.Init(tmpPath); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

	svc := NewSQLiteShortener()
	ctx := context.Background()
	orig := "https://example.com/concurrent"

	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers)

	codes := make(chan string, workers)
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			code, err := svc.Create(ctx, orig)
			if err != nil {
				errs <- err
				return
			}
			codes <- code
		}()
	}
	wg.Wait()
	close(codes)
	close(errs)

	for err := range errs {
		t.Fatalf("create failed under concurrency: %v", err)
	}

	var first string
	count := 0
	for code := range codes {
		if code == "" {
			t.Fatalf("expected non-empty code")
		}
		if first == "" {
			first = code
		}
		if code != first {
			t.Fatalf("expected same code for same original URL, got %q and %q", first, code)
		}
		count++
	}
	if count != workers {
		t.Fatalf("received %d codes, want %d", count, workers)
	}

	var rows int
	if err := db.Conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls WHERE original_url = ?", orig).Scan(&rows); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if rows != 1 {
		t.Fatalf("rows for original_url = %d, want 1", rows)
	}
}
