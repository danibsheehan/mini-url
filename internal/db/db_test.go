package db

import "testing"

func TestInit_MemoryDBAndSchema(t *testing.T) {
	if Conn != nil {
		_ = Conn.Close()
		Conn = nil
	}

	if err := Init(":memory:"); err != nil {
		t.Fatalf("Init(:memory:) error = %v", err)
	}
	defer func() {
		_ = Conn.Close()
		Conn = nil
	}()

	_, err := Conn.Exec(`INSERT INTO urls(code, original_url) VALUES(?, ?)`, "abc123", "https://example.com")
	if err != nil {
		t.Fatalf("insert into urls failed, schema may be missing: %v", err)
	}
}

func TestInit_InvalidPath(t *testing.T) {
	if Conn != nil {
		_ = Conn.Close()
		Conn = nil
	}

	err := Init("/definitely/not/a/real/path/urls.db")
	if err == nil {
		t.Fatalf("expected error for invalid path")
	}
}
