package nftables

import (
	"path/filepath"
	"testing"

	"github.com/sokol/system-control/internal/database"
)

func setupNftablesTest(t *testing.T) *Service {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := NewRepository(db)
	engine := NewEngine()
	return NewService(repo, engine)
}

func validCreateRequest() CreateGroupRequest {
	return CreateGroupRequest{
		Name:            "test-group",
		Enabled:         false,
		TargetIP:        "10.0.0.1",
		TargetReverseIP: "10.0.0.2",
		DestinationIP:   "192.168.1.1",
		Ports: []CreatePortRequest{
			{ExternalPort: 80, InternalPort: 8080, ProtocolTCP: true, ProtocolUDP: false},
		},
	}
}

func TestCreate_Success(t *testing.T) {
	svc := setupNftablesTest(t)
	req := validCreateRequest()

	group, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if group.ID == 0 {
		t.Error("expected non-zero group ID")
	}
	if group.Name != "test-group" {
		t.Errorf("expected name 'test-group', got %q", group.Name)
	}
	if len(group.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(group.Ports))
	}
	if group.Ports[0].ExternalPort != 80 {
		t.Errorf("expected external port 80, got %d", group.Ports[0].ExternalPort)
	}
}

func TestCreate_InvalidName(t *testing.T) {
	svc := setupNftablesTest(t)
	req := validCreateRequest()
	req.Name = ""

	_, err := svc.Create(req)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestCreate_InvalidTargetIP(t *testing.T) {
	svc := setupNftablesTest(t)
	req := validCreateRequest()
	req.TargetIP = "invalid"

	_, err := svc.Create(req)
	if err == nil {
		t.Fatal("expected error for invalid targetIP, got nil")
	}
}

func TestCreate_InvalidPort(t *testing.T) {
	svc := setupNftablesTest(t)
	req := validCreateRequest()
	req.Ports = []CreatePortRequest{
		{ExternalPort: 0, InternalPort: 80, ProtocolTCP: true},
	}

	_, err := svc.Create(req)
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
}

func TestCreate_NoProtocol(t *testing.T) {
	svc := setupNftablesTest(t)
	req := validCreateRequest()
	req.Ports = []CreatePortRequest{
		{ExternalPort: 80, InternalPort: 80, ProtocolTCP: false, ProtocolUDP: false},
	}

	_, err := svc.Create(req)
	if err == nil {
		t.Fatal("expected error when no protocol selected, got nil")
	}
}

func TestGetAll_Empty(t *testing.T) {
	svc := setupNftablesTest(t)

	groups, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if groups != nil && len(groups) != 0 {
		t.Errorf("expected empty list, got %d groups", len(groups))
	}
}

func TestGetAll_WithGroups(t *testing.T) {
	svc := setupNftablesTest(t)

	_, err := svc.Create(validCreateRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	groups, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
}

func TestGetByID(t *testing.T) {
	svc := setupNftablesTest(t)
	created, err := svc.Create(validCreateRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	group, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if group.Name != "test-group" {
		t.Errorf("expected name 'test-group', got %q", group.Name)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := setupNftablesTest(t)

	_, err := svc.GetByID(999)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestUpdate_Success(t *testing.T) {
	svc := setupNftablesTest(t)
	created, err := svc.Create(validCreateRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updateReq := UpdateGroupRequest{
		Name:            "updated-group",
		Enabled:         true,
		TargetIP:        "10.0.0.10",
		TargetReverseIP: "10.0.0.20",
		DestinationIP:   "192.168.2.1",
		Ports: []CreatePortRequest{
			{ExternalPort: 443, InternalPort: 8443, ProtocolTCP: true, ProtocolUDP: false},
			{ExternalPort: 53, InternalPort: 53, ProtocolTCP: true, ProtocolUDP: true},
		},
	}

	updated, err := svc.Update(created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "updated-group" {
		t.Errorf("expected name 'updated-group', got %q", updated.Name)
	}
	if len(updated.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(updated.Ports))
	}
}

func TestUpdate_InvalidData(t *testing.T) {
	svc := setupNftablesTest(t)
	created, err := svc.Create(validCreateRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updateReq := UpdateGroupRequest{
		Name:            "",
		TargetIP:        "10.0.0.1",
		TargetReverseIP: "10.0.0.2",
		DestinationIP:   "192.168.1.1",
	}

	_, err = svc.Update(created.ID, updateReq)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestDelete_Success(t *testing.T) {
	svc := setupNftablesTest(t)
	created, err := svc.Create(validCreateRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = svc.GetByID(created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestSetEnabled(t *testing.T) {
	svc := setupNftablesTest(t)
	req := validCreateRequest()
	req.Enabled = false

	created, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	enabled, err := svc.SetEnabled(created.ID, true)
	if err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	if !enabled.Enabled {
		t.Error("expected Enabled=true")
	}

	disabled, err := svc.SetEnabled(created.ID, false)
	if err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	if disabled.Enabled {
		t.Error("expected Enabled=false")
	}
}

func TestSetEnabled_NotFound(t *testing.T) {
	svc := setupNftablesTest(t)

	_, err := svc.SetEnabled(999, true)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestCreate_NoPorts(t *testing.T) {
	svc := setupNftablesTest(t)
	req := validCreateRequest()
	req.Ports = nil

	group, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Create with no ports: %v", err)
	}
	if len(group.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(group.Ports))
	}
}
