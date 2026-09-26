package store

import (
	"path/filepath"
	"testing"
)

func TestStore_EndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_wrenpanel.db")

	s, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer s.DB().Close()

	// 1. Account
	acc := &Account{
		Username:    "user1",
		HomeDir:     "/home/user1",
		DiskQuotaMB: 5000,
		Status:      "active",
	}
	if err := s.CreateAccount(acc); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}
	if acc.ID == 0 {
		t.Fatalf("Expected non-zero account ID")
	}

	// 2. PHP Version
	php := &PhpVersion{
		Version:    "8.2",
		FpmService: "php8.2-fpm",
		BinaryPath: "/usr/bin/php8.2",
	}
	if err := s.CreatePhpVersion(php); err != nil {
		t.Fatalf("CreatePhpVersion failed: %v", err)
	}

	// 3. Vhost: primary domain
	vhostPrimary := &Vhost{
		AccountID:    acc.ID,
		Type:         "primary",
		FQDN:         "example.com",
		Docroot:      "/home/user1/public_html",
		WebEngine:    "nginx",
		PhpVersionID: &php.ID,
		Status:       "active",
	}
	if err := s.CreateVhost(vhostPrimary); err != nil {
		t.Fatalf("CreateVhost primary failed: %v", err)
	}
	if vhostPrimary.DocrootOwnerID == nil || *vhostPrimary.DocrootOwnerID != vhostPrimary.ID {
		t.Fatalf("Expected vhost to own its docroot")
	}

	// 4. Vhost: subdomain with shared docroot
	vhostSub := &Vhost{
		AccountID:      acc.ID,
		ParentVhostID:  &vhostPrimary.ID,
		Type:           "subdomain",
		FQDN:           "app.example.com",
		Docroot:        "/home/user1/public_html", // sharing docroot
		DocrootOwnerID: &vhostPrimary.ID,          // owned by primary
		WebEngine:      "nginx",
		PhpVersionID:   &php.ID,
		Status:         "active",
	}
	if err := s.CreateVhost(vhostSub); err != nil {
		t.Fatalf("CreateVhost sub failed: %v", err)
	}

	// 5. Test Deletion Impact: Cannot delete primary while child exists
	impactPrimary, err := s.AnalyzeDeletionImpact(vhostPrimary.ID)
	if err != nil {
		t.Fatalf("AnalyzeDeletionImpact primary failed: %v", err)
	}
	if impactPrimary.AllowDelete {
		t.Fatalf("Primary vhost should NOT be allowed to delete while child vhosts exist")
	}

	// 6. Test Deletion Impact for Subdomain: Shares docroot with primary, so must keep files!
	impactSub, err := s.AnalyzeDeletionImpact(vhostSub.ID)
	if err != nil {
		t.Fatalf("AnalyzeDeletionImpact sub failed: %v", err)
	}
	if !impactSub.AllowDelete {
		t.Fatalf("Sub vhost should be allowed to delete")
	}
	if impactSub.CanDeleteFiles {
		t.Fatalf("Sub vhost shares docroot and is not owner, MUST NOT be allowed to delete files")
	}

	// 7. Test PHP version deletion protection: Cannot delete PHP version when in use
	err = s.DeletePhpVersion(php.ID)
	if err == nil {
		t.Fatalf("Expected error when deleting PHP version in use, got nil")
	}

	// 8. Test Port Allocation
	app := &App{
		AccountID: acc.ID,
		Kind:      "node",
		Name:      "test-node-app",
		Status:    "stopped",
	}
	if err := s.CreateApp(app); err != nil {
		t.Fatalf("CreateApp failed: %v", err)
	}

	port, err := s.AllocatePort(app.ID, 15000, 15010)
	if err != nil {
		t.Fatalf("AllocatePort failed: %v", err)
	}
	if port < 15000 || port > 15010 {
		t.Fatalf("Allocated port %d out of range", port)
	}
	_ = s.ReleasePort(port)
}
