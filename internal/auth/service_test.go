package auth

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sokol/system-control/internal/database"
)

func setupAuthTest(t *testing.T) *Service {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := NewRepository(db)
	return NewService(repo, 3600)
}

func TestEnsureAdminUser_CreatesWhenNone(t *testing.T) {
	svc := setupAuthTest(t)

	if err := svc.EnsureAdminUser("admin", "secret"); err != nil {
		t.Fatalf("EnsureAdminUser: %v", err)
	}

	// Verify user was created
	user, err := svc.repo.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("expected username 'admin', got %q", user.Username)
	}
}

func TestEnsureAdminUser_SkipsWhenExists(t *testing.T) {
	svc := setupAuthTest(t)

	if err := svc.EnsureAdminUser("admin", "secret"); err != nil {
		t.Fatalf("first EnsureAdminUser: %v", err)
	}

	// Should not create another user
	if err := svc.EnsureAdminUser("admin2", "secret2"); err != nil {
		t.Fatalf("second EnsureAdminUser: %v", err)
	}

	count, err := svc.repo.UserCount()
	if err != nil {
		t.Fatalf("UserCount: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := setupAuthTest(t)
	if err := svc.EnsureAdminUser("admin", "password123"); err != nil {
		t.Fatalf("EnsureAdminUser: %v", err)
	}

	session, user, err := svc.Login("admin", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.Username != "admin" {
		t.Errorf("expected username 'admin', got %q", user.Username)
	}
	if session.ID == "" {
		t.Error("expected non-empty session ID")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := setupAuthTest(t)
	if err := svc.EnsureAdminUser("admin", "password123"); err != nil {
		t.Fatalf("EnsureAdminUser: %v", err)
	}

	_, _, err := svc.Login("admin", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	svc := setupAuthTest(t)

	_, _, err := svc.Login("nobody", "password")
	if err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestValidateSession_Valid(t *testing.T) {
	svc := setupAuthTest(t)
	if err := svc.EnsureAdminUser("admin", "pass"); err != nil {
		t.Fatalf("EnsureAdminUser: %v", err)
	}

	session, _, err := svc.Login("admin", "pass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	user, err := svc.ValidateSession(session.ID)
	if err != nil {
		t.Fatalf("ValidateSession: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("expected username 'admin', got %q", user.Username)
	}
}

func TestValidateSession_InvalidID(t *testing.T) {
	svc := setupAuthTest(t)

	_, err := svc.ValidateSession("nonexistent-session-id")
	if err == nil {
		t.Fatal("expected error for invalid session ID, got nil")
	}
}

func TestValidateSession_Expired(t *testing.T) {
	// Create a service with a very short session max age
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo, 1) // 1 second session

	if err := svc.EnsureAdminUser("admin", "pass"); err != nil {
		t.Fatalf("EnsureAdminUser: %v", err)
	}

	session, _, err := svc.Login("admin", "pass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// Manually expire the session
	_ = repo.ExtendSession(session.ID, time.Now().Add(-time.Hour))

	_, err = svc.ValidateSession(session.ID)
	if err == nil {
		t.Fatal("expected error for expired session, got nil")
	}
}

func TestLogout(t *testing.T) {
	svc := setupAuthTest(t)
	if err := svc.EnsureAdminUser("admin", "pass"); err != nil {
		t.Fatalf("EnsureAdminUser: %v", err)
	}

	session, _, err := svc.Login("admin", "pass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if err := svc.Logout(session.ID); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// Session should be invalid after logout
	_, err = svc.ValidateSession(session.ID)
	if err == nil {
		t.Fatal("expected error after logout, got nil")
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	svc := setupAuthTest(t)
	if err := svc.EnsureAdminUser("admin", "pass"); err != nil {
		t.Fatalf("EnsureAdminUser: %v", err)
	}

	session, _, err := svc.Login("admin", "pass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// Expire the session manually
	_ = svc.repo.ExtendSession(session.ID, time.Now().Add(-time.Hour))

	if err := svc.CleanupExpiredSessions(); err != nil {
		t.Fatalf("CleanupExpiredSessions: %v", err)
	}

	// Session should be gone
	_, err = svc.ValidateSession(session.ID)
	if err == nil {
		t.Fatal("expected error after cleanup, got nil")
	}
}
