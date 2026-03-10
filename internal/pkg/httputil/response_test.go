package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Error != "" {
		t.Errorf("expected empty error, got %q", resp.Error)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "something went wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "something went wrong" {
		t.Errorf("expected error 'something went wrong', got %q", resp.Error)
	}
}

func TestDecode(t *testing.T) {
	body := `{"name":"test","value":42}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	if err := Decode(req, &result); err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if result.Name != "test" || result.Value != 42 {
		t.Errorf("unexpected decoded result: %+v", result)
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	var result struct{}
	if err := Decode(req, &result); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestJSON_CustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusCreated, nil)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}
