package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	sqlite3 "github.com/mattn/go-sqlite3"
	"mini-url/internal/db"
)

func TestSQLiteShortener_CreateAndResolve(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

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

func TestSQLiteShortener_CreateRetriesOnCollision(t *testing.T) {
	if err := db.Init(":memory:"); err != nil {
		t.Fatalf("db init: %v", err)
	}
	defer db.Conn.Close()

	origGenerator := codeGenerator
	defer func() { codeGenerator = origGenerator }()

	// First generated code collides with an existing row, second should succeed.
	calls := 0
	codeGenerator = func(length int) (string, error) {
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

	svc := NewSQLiteShortener()
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

	origGenerator := codeGenerator
	defer func() { codeGenerator = origGenerator }()

	wantErr := errors.New("rng unavailable")
	codeGenerator = func(length int) (string, error) {
		return "", wantErr
	}

	svc := NewSQLiteShortener()
	_, err := svc.Create(context.Background(), "https://new-url.example")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
