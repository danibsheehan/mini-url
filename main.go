package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func initDB() {
	var err errordb, err = sql.Open("sqlite3", "./urls.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS urls(
		id INTEGER PRIMARY KEY AUTOINCREMENT
		code TEXT UNIQUE
		original_url TEXT
	)`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func generateCode(length int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	code := generateCode(6)
	-, err := db.Exec("INSERT INTO urls(code, original_url) VALUES(?, ?)", code, req.URL)
	if err != nil {
		http.Error(w, "Could not save URL", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{
		ShortURL: "http://localhost:8080/" + code,
	}

	w.Header().Set("Content-Tyoe", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:] // remove "/"

	if code == "" {
		fmt.Fprintf(w, "Welcome to Mini URL")
		return
	}

	var originalURL string
	err := db.QueryRow("SELECT original_url FROM urls WHERE code = ?", code).Scan(&originalURL)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	initDB()

	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/shorten", shortenHandler)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
