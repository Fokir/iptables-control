package database

import (
	"path/filepath"
	"testing"
)

func TestOpen_CreatesAndMigrates(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q) returned error: %v", dbPath, err)
	}
	defer db.Close()

	// Verify that tables were created by the migration
	tables := []string{"users", "sessions", "nat_groups", "nat_ports", "network_nodes", "nginx_domains", "audit_logs"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found after migration: %v", table, err)
		}
	}
}

func TestOpen_CanQuery(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer db.Close()

	// Insert and read back a user to verify DB is functional
	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", "testuser", "hash")
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("SELECT COUNT failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}
}

func TestOpen_IdempotentMigrations(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first Open returned error: %v", err)
	}
	db1.Close()

	// Opening again should not fail (migrations use IF NOT EXISTS)
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open returned error: %v", err)
	}
	db2.Close()
}

func TestOpen_ForeignKeysEnabled(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer db.Close()

	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("PRAGMA foreign_keys failed: %v", err)
	}
	if fk != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fk)
	}
}
