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
	statsFn   func(ctx context.Context, code string) (services.URLStats, error)
}

func (m *mockShortener) Create(ctx context.Context, original string) (string, error) {
	return m.createFn(ctx, original)
}

func (m *mockShortener) Resolve(ctx context.Context, code string) (string, error) {
	return m.resolveFn(ctx, code)
}

func (m *mockShortener) Stats(ctx context.Context, code string) (services.URLStats, error) {
	return m.statsFn(ctx, code)
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
		req.Host = "short.ly:9090"
		rr := httptest.NewRecorder()
		h.Shorten(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q, want %q", got, "application/json")
		}
		if body := rr.Body.String(); !strings.Contains(body, `"short_url":"http://short.ly:9090/abc123"`) {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("uses forwarded proto when present", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "abc123", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				return "", nil
			},
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
		req.Host = "short.ly"
		req.Header.Set("X-Forwarded-Proto", "https")
		rr := httptest.NewRecorder()
		h.Shorten(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if body := rr.Body.String(); !strings.Contains(body, `"short_url":"https://short.ly/abc123"`) {
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
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

	t.Run("uses request context for resolve", func(t *testing.T) {
		type ctxKey string
		const key ctxKey = "request-id"

		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) {
				return "unused", nil
			},
			resolveFn: func(ctx context.Context, code string) (string, error) {
				if got := ctx.Value(key); got != "abc-123" {
					t.Fatalf("context value = %v, want %q", got, "abc-123")
				}
				return "https://example.com/path", nil
			},
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		req = req.WithContext(context.WithValue(req.Context(), key, "abc-123"))
		rr := httptest.NewRecorder()
		h.Redirect(rr, req)

		if rr.Code != http.StatusFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
		}
	})
}

func TestShortenerHandler_Stats(t *testing.T) {
	t.Run("rejects non-get method", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) { return "unused", nil },
			resolveFn: func(ctx context.Context, code string) (string, error) { return "", nil },
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/stats/abc123", nil)
		rr := httptest.NewRecorder()
		h.Stats(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("returns 404 when code is not found", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) { return "unused", nil },
			resolveFn: func(ctx context.Context, code string) (string, error) { return "", nil },
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, services.ErrNotFound
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/stats/missing", nil)
		rr := httptest.NewRecorder()
		h.Stats(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 500 on unexpected stats error", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) { return "unused", nil },
			resolveFn: func(ctx context.Context, code string) (string, error) { return "", nil },
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				return services.URLStats{}, errors.New("db down")
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/stats/abc123", nil)
		rr := httptest.NewRecorder()
		h.Stats(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("returns stats json when found", func(t *testing.T) {
		h := NewShortenerHandler(&mockShortener{
			createFn: func(ctx context.Context, original string) (string, error) { return "unused", nil },
			resolveFn: func(ctx context.Context, code string) (string, error) { return "", nil },
			statsFn: func(ctx context.Context, code string) (services.URLStats, error) {
				if code != "abc123" {
					t.Fatalf("code = %q, want %q", code, "abc123")
				}
				return services.URLStats{
					OriginalURL: "https://example.com/path",
					ClickCount:  7,
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/stats/abc123", nil)
		rr := httptest.NewRecorder()
		h.Stats(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if got := rr.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q, want %q", got, "application/json")
		}
		body := rr.Body.String()
		if !strings.Contains(body, `"code":"abc123"`) {
			t.Fatalf("unexpected body: %s", body)
		}
		if !strings.Contains(body, `"original_url":"https://example.com/path"`) {
			t.Fatalf("unexpected body: %s", body)
		}
		if !strings.Contains(body, `"click_count":7`) {
			t.Fatalf("unexpected body: %s", body)
		}
	})
}
