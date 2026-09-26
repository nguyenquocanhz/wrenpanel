package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"wrenpanel/internal/store"
	"wrenpanel/internal/worker"
)

func setupTestServer(t *testing.T) (*Server, *store.Store, func()) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "api_test.db")

	s, err := store.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}

	executor := worker.NewExecutor(tempDir, true)
	addr := "127.0.0.1:29888"
	wServer := worker.NewServer("tcp", addr, executor)

	go func() {
		_ = wServer.Start()
	}()

	client := worker.NewClient("tcp", addr)
	srv := NewServer(s, client, nil)

	cleanup := func() {
		client.Close()
		wServer.Close()
		s.DB().Close()
	}

	return srv, s, cleanup
}

func TestAPI_VhostAndSafeDeletion(t *testing.T) {
	srv, s, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Create Account
	acc := &store.Account{
		Username:    "alice",
		HomeDir:     "/home/alice",
		DiskQuotaMB: 5000,
		Status:      "active",
	}
	if err := s.CreateAccount(acc); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	// 2. Reject Reserved Prefix (e.g. mail.alice.com)
	reservedReq := CreateVhostRequest{
		AccountID: acc.ID,
		Type:      "subdomain",
		FQDN:      "mail.alice.com",
	}
	body, _ := json.Marshal(reservedReq)
	req := httptest.NewRequest("POST", "/api/vhosts", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for reserved prefix, got %d", w.Code)
	}

	// 3. Create Primary Domain (example.com)
	primaryReq := CreateVhostRequest{
		AccountID: acc.ID,
		Type:      "primary",
		FQDN:      "example.com",
		Docroot:   "/home/alice/public_html",
		WebEngine: "nginx",
	}
	body, _ = json.Marshal(primaryReq)
	req = httptest.NewRequest("POST", "/api/vhosts", bytes.NewReader(body))
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for primary domain, got %d: %s", w.Code, w.Body.String())
	}

	var createdPrimary store.Vhost
	_ = json.Unmarshal(w.Body.Bytes(), &createdPrimary)

	// 4. Create Subdomain sharing docroot (app.example.com)
	subReq := CreateVhostRequest{
		AccountID:      acc.ID,
		ParentVhostID:  &createdPrimary.ID,
		Type:           "subdomain",
		FQDN:           "app.example.com",
		Docroot:        "/home/alice/public_html",
		DocrootOwnerID: &createdPrimary.ID,
		WebEngine:      "nginx",
	}
	body, _ = json.Marshal(subReq)
	req = httptest.NewRequest("POST", "/api/vhosts", bytes.NewReader(body))
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for subdomain, got %d: %s", w.Code, w.Body.String())
	}

	var createdSub store.Vhost
	_ = json.Unmarshal(w.Body.Bytes(), &createdSub)

	// 5. Try deleting primary domain while subdomain exists -> MUST BE BLOCKED
	req = httptest.NewRequest("DELETE", "/api/vhosts/1", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden when deleting primary domain with existing child, got %d", w.Code)
	}

	// 6. Delete Subdomain -> Permitted, but must NOT delete files
	req = httptest.NewRequest("DELETE", "/api/vhosts/2?delete_files=true", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK deleting subdomain, got %d", w.Code)
	}

	var delResult map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &delResult)
	if delResult["deleted_files"] == true {
		t.Errorf("Subdomain sharing docroot must not delete shared files!")
	}

	// 7. Verify Audit Log was recorded
	req = httptest.NewRequest("GET", "/api/audit-logs", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	var logs []store.AuditLog
	_ = json.Unmarshal(w.Body.Bytes(), &logs)
	if len(logs) == 0 {
		t.Errorf("Expected audit logs to be recorded")
	}
}
