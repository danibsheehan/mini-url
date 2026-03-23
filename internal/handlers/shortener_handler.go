package handlers

import (
	"context"
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

	resp := ShortenResponse{ShortURL: "http://localhost:8080/" + code}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ShortenerHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" {
		w.Write([]byte("Welcome to Mini URL"))
		return
	}

	orig, err := h.svc.Resolve(context.Background(), code)
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
