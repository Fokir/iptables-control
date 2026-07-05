package nginx

import (
	"path/filepath"
	"testing"

	"github.com/sokol/system-control/internal/database"
)

type stubSettings struct{ forceTLS12 bool }

func (s stubSettings) ForceTLS12() bool { return s.forceTLS12 }

func setupNginxTest(t *testing.T) *Service {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sitesDir := t.TempDir()
	repo := NewRepository(db)
	return NewService(repo, sitesDir, 8080, stubSettings{})
}

// mustCreate creates a domain, ignoring nginx config errors (expected in test env without root).
func mustCreate(t *testing.T, svc *Service, req CreateDomainRequest) *Domain {
	t.Helper()
	domain, err := svc.Create(req)
	if err != nil && domain == nil {
		t.Fatalf("Create: %v", err)
	}
	return domain
}

func TestRebuildTargets(t *testing.T) {
	domains := []Domain{
		{Domain: "enabled-ssl.example.com", Enabled: true, SSLEnabled: true},
		{Domain: "enabled-nossl.example.com", Enabled: true, SSLEnabled: false},
		{Domain: "disabled-ssl.example.com", Enabled: false, SSLEnabled: true},
	}

	targets := rebuildTargets(domains)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Domain != "enabled-ssl.example.com" {
		t.Errorf("expected 'enabled-ssl.example.com', got %q", targets[0].Domain)
	}
}

func TestCreateDomain_Success(t *testing.T) {
	svc := setupNginxTest(t)

	domain := mustCreate(t, svc, CreateDomainRequest{
		Domain:       "example.com",
		UpstreamIP:   "10.0.0.1",
		UpstreamPort: 8080,
	})
	if domain.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if domain.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", domain.Domain)
	}
	if domain.UpstreamPort != 8080 {
		t.Errorf("expected upstream port 8080, got %d", domain.UpstreamPort)
	}
	if !domain.Enabled {
		t.Error("expected Enabled=true for new domain")
	}
}

func TestCreateDomain_DefaultPort(t *testing.T) {
	svc := setupNginxTest(t)

	domain := mustCreate(t, svc, CreateDomainRequest{
		Domain:       "example.com",
		UpstreamIP:   "10.0.0.1",
		UpstreamPort: 0, // should default to 80
	})
	if domain.UpstreamPort != 80 {
		t.Errorf("expected default upstream port 80, got %d", domain.UpstreamPort)
	}
}

func TestCreateDomain_InvalidDomain(t *testing.T) {
	svc := setupNginxTest(t)

	_, err := svc.Create(CreateDomainRequest{
		Domain:       "invalid",
		UpstreamIP:   "10.0.0.1",
		UpstreamPort: 80,
	})
	if err == nil {
		t.Fatal("expected error for invalid domain, got nil")
	}
}

func TestCreateDomain_InvalidUpstreamIP(t *testing.T) {
	svc := setupNginxTest(t)

	_, err := svc.Create(CreateDomainRequest{
		Domain:       "example.com",
		UpstreamIP:   "bad-ip",
		UpstreamPort: 80,
	})
	if err == nil {
		t.Fatal("expected error for invalid upstream IP, got nil")
	}
}

func TestGetAllDomains_Empty(t *testing.T) {
	svc := setupNginxTest(t)

	domains, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}

func TestGetAllDomains_Multiple(t *testing.T) {
	svc := setupNginxTest(t)

	mustCreate(t, svc, CreateDomainRequest{Domain: "a.example.com", UpstreamIP: "10.0.0.1", UpstreamPort: 80})
	mustCreate(t, svc, CreateDomainRequest{Domain: "b.example.com", UpstreamIP: "10.0.0.2", UpstreamPort: 80})

	domains, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(domains))
	}
}

func TestUpdateDomain_Success(t *testing.T) {
	svc := setupNginxTest(t)

	created := mustCreate(t, svc, CreateDomainRequest{Domain: "old.example.com", UpstreamIP: "10.0.0.1", UpstreamPort: 80})

	updated, err := svc.Update(created.ID, UpdateDomainRequest{
		Domain:       "new.example.com",
		UpstreamIP:   "10.0.0.99",
		UpstreamPort: 9090,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Domain != "new.example.com" {
		t.Errorf("expected domain 'new.example.com', got %q", updated.Domain)
	}
	if updated.UpstreamIP != "10.0.0.99" {
		t.Errorf("expected upstream IP '10.0.0.99', got %q", updated.UpstreamIP)
	}
}

func TestUpdateDomain_InvalidDomain(t *testing.T) {
	svc := setupNginxTest(t)

	created := mustCreate(t, svc, CreateDomainRequest{Domain: "example.com", UpstreamIP: "10.0.0.1", UpstreamPort: 80})

	_, err :=  svc.Update(created.ID, UpdateDomainRequest{Domain: "invalid", UpstreamIP: "10.0.0.1", UpstreamPort: 80})
	if err == nil {
		t.Fatal("expected error for invalid domain, got nil")
	}
}

func TestUpdateDomain_NotFound(t *testing.T) {
	svc := setupNginxTest(t)

	_, err := svc.Update(999, UpdateDomainRequest{Domain: "example.com", UpstreamIP: "10.0.0.1", UpstreamPort: 80})
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestDeleteDomain(t *testing.T) {
	svc := setupNginxTest(t)

	created := mustCreate(t, svc, CreateDomainRequest{Domain: "example.com", UpstreamIP: "10.0.0.1", UpstreamPort: 80})

	if err := svc.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	domains, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains after delete, got %d", len(domains))
	}
}

func TestDeleteDomain_NotFound(t *testing.T) {
	svc := setupNginxTest(t)

	err := svc.Delete(999)
	if err == nil {
		t.Fatal("expected error for non-existent domain, got nil")
	}
}

func TestSetEnabled_Enable(t *testing.T) {
	svc := setupNginxTest(t)

	created := mustCreate(t, svc, CreateDomainRequest{Domain: "example.com", UpstreamIP: "10.0.0.1", UpstreamPort: 80})

	// Disable first
	disabled, err := svc.SetEnabled(created.ID, false)
	if err != nil {
		t.Fatalf("SetEnabled(false): %v", err)
	}
	if disabled.Enabled {
		t.Error("expected Enabled=false")
	}

	// Re-enable
	enabled, err := svc.SetEnabled(created.ID, true)
	if err != nil {
		t.Fatalf("SetEnabled(true): %v", err)
	}
	if !enabled.Enabled {
		t.Error("expected Enabled=true")
	}
}

func TestSetEnabled_NotFound(t *testing.T) {
	svc := setupNginxTest(t)

	_, err := svc.SetEnabled(999, true)
	if err == nil {
		t.Fatal("expected error for non-existent domain, got nil")
	}
}
