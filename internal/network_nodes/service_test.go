package network_nodes

import (
	"path/filepath"
	"testing"

	"github.com/sokol/system-control/internal/database"
)

func setupNodeTest(t *testing.T) *Service {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := NewRepository(db)
	return NewService(repo)
}

func TestCreateNode_Success(t *testing.T) {
	svc := setupNodeTest(t)

	node, err := svc.Create(CreateNodeRequest{Name: "server1", IP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if node.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if node.Name != "server1" {
		t.Errorf("expected name 'server1', got %q", node.Name)
	}
	if node.IP != "10.0.0.1" {
		t.Errorf("expected IP '10.0.0.1', got %q", node.IP)
	}
}

func TestCreateNode_EmptyName(t *testing.T) {
	svc := setupNodeTest(t)

	_, err := svc.Create(CreateNodeRequest{Name: "", IP: "10.0.0.1"})
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestCreateNode_InvalidIP(t *testing.T) {
	svc := setupNodeTest(t)

	_, err := svc.Create(CreateNodeRequest{Name: "server1", IP: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid IP, got nil")
	}
}

func TestGetAllNodes_Empty(t *testing.T) {
	svc := setupNodeTest(t)

	nodes, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestGetAllNodes_Multiple(t *testing.T) {
	svc := setupNodeTest(t)

	_, err := svc.Create(CreateNodeRequest{Name: "a-server", IP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	_, err = svc.Create(CreateNodeRequest{Name: "b-server", IP: "10.0.0.2"})
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}

	nodes, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestUpdateNode_Success(t *testing.T) {
	svc := setupNodeTest(t)

	node, err := svc.Create(CreateNodeRequest{Name: "server1", IP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.Update(node.ID, UpdateNodeRequest{Name: "server1-updated", IP: "10.0.0.99"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "server1-updated" {
		t.Errorf("expected name 'server1-updated', got %q", updated.Name)
	}
	if updated.IP != "10.0.0.99" {
		t.Errorf("expected IP '10.0.0.99', got %q", updated.IP)
	}
}

func TestUpdateNode_InvalidIP(t *testing.T) {
	svc := setupNodeTest(t)

	node, err := svc.Create(CreateNodeRequest{Name: "server1", IP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Update(node.ID, UpdateNodeRequest{Name: "server1", IP: "bad"})
	if err == nil {
		t.Fatal("expected error for invalid IP, got nil")
	}
}

func TestUpdateNode_EmptyName(t *testing.T) {
	svc := setupNodeTest(t)

	node, err := svc.Create(CreateNodeRequest{Name: "server1", IP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Update(node.ID, UpdateNodeRequest{Name: "", IP: "10.0.0.1"})
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestDeleteNode(t *testing.T) {
	svc := setupNodeTest(t)

	node, err := svc.Create(CreateNodeRequest{Name: "server1", IP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(node.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	nodes, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes after delete, got %d", len(nodes))
	}
}
