package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mini-url/internal/services"
)

type mockShortener struct {
	createFn  func(ctx context.Context, original string) (string, error)
	resolveFn func(ctx context.Context, code string) (string, error)
}

func (m *mockShortener) Create(ctx context.Context, original string) (string, error) {
	return m.createFn(ctx, original)
}

func (m *mockShortener) Resolve(ctx context.Context, code string) (string, error) {
	return m.resolveFn(ctx, code)
}

func TestShortenerHandler_Shorten(t *testing.T) {
	t.Run("rejects non-post method", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
		rr := httptest.NewRecorder()
		h.Shorten(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("{"))
		rr := httptest.NewRecorder()
		h.Shorten(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 500 when service create fails", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "", errors.New("boom")
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
		rr := httptest.NewRecorder()
		h.Shorten(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("returns short url on success", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				if original != "https://example.com" {
					t.Fatalf("original = %q, want %q", original, "https://example.com")
				}
				return "abc123", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
		rr := httptest.NewRecorder()
		h.Shorten(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q, want %q", got, "application/json")
		}
		if body := rr.Body.String(); !strings.Contains(body, `"short_url":"http://localhost:8080/abc123"`) {
			t.Fatalf("unexpected body: %s", body)
		}
	})
}

func TestShortenerHandler_Redirect(t *testing.T) {
	t.Run("returns welcome on root path", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		h.Redirect(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if got := rr.Body.String(); got != "Welcome to Mini URL" {
			t.Fatalf("body = %q, want %q", got, "Welcome to Mini URL")
		}
	})

	t.Run("returns 404 when code is not found", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", services.ErrNotFound
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		rr := httptest.NewRecorder()
		h.Redirect(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 404 for typed not found error", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", &services.NotFoundError{}
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/missing-typed", nil)
		rr := httptest.NewRecorder()
		h.Redirect(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 500 on unexpected resolve error", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", errors.New("db down")
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		rr := httptest.NewRecorder()
		h.Redirect(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("redirects to original url when found", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				if code != "abc123" {
					t.Fatalf("code = %q, want %q", code, "abc123")
				}
				return "https://example.com/path", nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		rr := httptest.NewRecorder()
		h.Redirect(rr, req)

		if rr.Code != http.StatusFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
		}
		if loc := rr.Header().Get("Location"); loc != "https://example.com/path" {
			t.Fatalf("location = %q, want %q", loc, "https://example.com/path")
		}
	})
}
