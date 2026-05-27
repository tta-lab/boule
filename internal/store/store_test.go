package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Fatal("DefaultPath returned empty string")
	}
	if filepath.Ext(path) != ".db" {
		t.Fatalf("expected .db extension, got %s", filepath.Ext(path))
	}
}

func TestOpen(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestMigrate(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='messages'").Scan(&tableName)
	if err != nil {
		t.Fatalf("messages table not found: %v", err)
	}
	if tableName != "messages" {
		t.Fatalf("expected table 'messages', got '%s'", tableName)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	for i := 0; i < 3; i++ {
		if err := Migrate(db); err != nil {
			t.Fatalf("Migrate iteration %d failed: %v", i, err)
		}
	}
}
