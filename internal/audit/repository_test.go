package audit

import (
	"path/filepath"
	"testing"

	"github.com/sokol/system-control/internal/database"
)

func setupAuditTest(t *testing.T) *Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewRepository(db)
}

func int64Ptr(v int64) *int64 { return &v }

func TestCreate_AndList(t *testing.T) {
	repo := setupAuditTest(t)

	// Create a user first to satisfy FK constraint
	db := repo.db
	db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", "admin", "hash")

	entry := &LogEntry{
		UserID:     int64Ptr(1),
		Username:   "admin",
		Action:     "create",
		EntityType: "nftables",
		EntityID:   int64Ptr(42),
		Details:    `{"name":"test"}`,
		IPAddress:  "127.0.0.1",
	}

	if err := repo.Create(entry); err != nil {
		t.Fatalf("Create: %v", err)
	}

	result, err := repo.List(ListParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total=1, got %d", result.Total)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	if result.Entries[0].Username != "admin" {
		t.Errorf("expected username 'admin', got %q", result.Entries[0].Username)
	}
	if result.Entries[0].Action != "create" {
		t.Errorf("expected action 'create', got %q", result.Entries[0].Action)
	}
}

func TestCreate_NilUserID(t *testing.T) {
	repo := setupAuditTest(t)

	entry := &LogEntry{
		UserID:     nil,
		Username:   "system",
		Action:     "startup",
		EntityType: "system",
		IPAddress:  "",
	}

	if err := repo.Create(entry); err != nil {
		t.Fatalf("Create with nil UserID: %v", err)
	}

	result, err := repo.List(ListParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total=1, got %d", result.Total)
	}
}

func TestList_FilterByAction(t *testing.T) {
	repo := setupAuditTest(t)

	for _, action := range []string{"create", "update", "create", "delete"} {
		repo.Create(&LogEntry{
			Username:   "admin",
			Action:     action,
			EntityType: "nftables",
			IPAddress:  "127.0.0.1",
		})
	}

	result, err := repo.List(ListParams{Page: 1, Limit: 10, Action: "create"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total=2 for action=create, got %d", result.Total)
	}
}

func TestList_FilterByEntityType(t *testing.T) {
	repo := setupAuditTest(t)

	repo.Create(&LogEntry{Username: "admin", Action: "create", EntityType: "nftables", IPAddress: "127.0.0.1"})
	repo.Create(&LogEntry{Username: "admin", Action: "create", EntityType: "nginx", IPAddress: "127.0.0.1"})
	repo.Create(&LogEntry{Username: "admin", Action: "create", EntityType: "nftables", IPAddress: "127.0.0.1"})

	result, err := repo.List(ListParams{Page: 1, Limit: 10, EntityType: "nginx"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total=1 for entityType=nginx, got %d", result.Total)
	}
}

func TestList_Pagination(t *testing.T) {
	repo := setupAuditTest(t)

	for i := 0; i < 5; i++ {
		repo.Create(&LogEntry{Username: "admin", Action: "create", EntityType: "test", IPAddress: "127.0.0.1"})
	}

	// Page 1, limit 2
	result, err := repo.List(ListParams{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("expected total=5, got %d", result.Total)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries on page 1, got %d", len(result.Entries))
	}

	// Page 3, limit 2 => 1 entry
	result, err = repo.List(ListParams{Page: 3, Limit: 2})
	if err != nil {
		t.Fatalf("List page 3: %v", err)
	}
	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry on page 3, got %d", len(result.Entries))
	}
}

func TestList_Empty(t *testing.T) {
	repo := setupAuditTest(t)

	result, err := repo.List(ListParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total=0, got %d", result.Total)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestCleanup(t *testing.T) {
	repo := setupAuditTest(t)

	// Insert an entry with a timestamp in the past
	_, err := repo.db.Exec(
		"INSERT INTO audit_logs (username, action, entity_type, ip_address, created_at) VALUES (?, ?, ?, ?, datetime('now', '-10 days'))",
		"admin", "create", "test", "127.0.0.1",
	)
	if err != nil {
		t.Fatalf("insert old entry: %v", err)
	}

	// Cleanup entries older than 5 days
	if err := repo.Cleanup(5); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	result, err := repo.List(ListParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total=0 after cleanup, got %d", result.Total)
	}
}
