package handlers

import (
	"encoding/json"
	"net/http"

	"mini-url/internal/services"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type StatsResponse struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
	ClickCount  int    `json:"click_count"`
}

type ShortenerHandler struct {
	svc services.Shortener
}

func NewShortenerHandler(s services.Shortener) *ShortenerHandler {
	return &ShortenerHandler{svc: s}
}

func (h *ShortenerHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code, err := h.svc.Create(r.Context(), req.URL)
	if err != nil {
		http.Error(w, "Could not save URL", http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}

	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}

	resp := ShortenResponse{ShortURL: scheme + "://" + host + "/" + code}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ShortenerHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" {
		w.Write([]byte("Welcome to Mini URL"))
		return
	}

	orig, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		if _, ok := err.(*services.NotFoundError); ok || err == services.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, orig, http.StatusFound)
}

func (h *ShortenerHandler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	const statsPrefix = "/stats/"
	if len(r.URL.Path) <= len(statsPrefix) || r.URL.Path[:len(statsPrefix)] != statsPrefix {
		http.NotFound(w, r)
		return
	}

	code := r.URL.Path[len(statsPrefix):]
	if code == "" {
		http.NotFound(w, r)
		return
	}

	stats, err := h.svc.Stats(r.Context(), code)
	if err != nil {
		if _, ok := err.(*services.NotFoundError); ok || err == services.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	resp := StatsResponse{
		Code:        code,
		OriginalURL: stats.OriginalURL,
		ClickCount:  stats.ClickCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
