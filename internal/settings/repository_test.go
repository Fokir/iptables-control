package settings

import (
	"path/filepath"
	"testing"

	"github.com/sokol/system-control/internal/database"
)

func setupRepo(t *testing.T) *Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepository(db)
}

func TestGet_MissingKey(t *testing.T) {
	repo := setupRepo(t)
	v, err := repo.Get("nope")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v != "" {
		t.Errorf("expected empty string for missing key, got %q", v)
	}
}

func TestSetGet_RoundTrip(t *testing.T) {
	repo := setupRepo(t)
	if err := repo.Set("force_tls12", "true"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	v, err := repo.Get("force_tls12")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v != "true" {
		t.Errorf("expected \"true\", got %q", v)
	}
	if err := repo.Set("force_tls12", "false"); err != nil {
		t.Fatalf("Set overwrite: %v", err)
	}
	v, _ = repo.Get("force_tls12")
	if v != "false" {
		t.Errorf("expected \"false\" after overwrite, got %q", v)
	}
}
