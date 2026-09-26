package worker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkerAllowlistValidation(t *testing.T) {
	// 1. Invalid action
	err := ValidateCommand(Command{Action: "rm -rf /"})
	if err == nil {
		t.Errorf("expected error for non-allowlisted action")
	}

	// 2. Invalid FQDN
	err = ValidateCommand(Command{
		Action: ActionVhostWriteNginxConfig,
		Params: map[string]string{
			"fqdn":    "bad;name&&rm",
			"content": "server {}",
		},
	})
	if err == nil {
		t.Errorf("expected error for malicious FQDN")
	}

	// 3. Valid FQDN
	err = ValidateCommand(Command{
		Action: ActionVhostWriteNginxConfig,
		Params: map[string]string{
			"fqdn":    "mysite.com",
			"content": "server {}",
		},
	})
	if err != nil {
		t.Errorf("unexpected error for valid FQDN: %v", err)
	}

	// 4. Invalid unit name
	err = ValidateCommand(Command{
		Action: ActionSystemdWriteUnit,
		Params: map[string]string{
			"unit_name": "malicious.service",
			"content":   "[Unit]",
		},
	})
	if err == nil {
		t.Errorf("expected error for non-wrenpanel unit name")
	}

	// 5. Valid unit name
	err = ValidateCommand(Command{
		Action: ActionSystemdWriteUnit,
		Params: map[string]string{
			"unit_name": "wrenpanel-app-nodeapp-a1b2c3.service",
			"content":   "[Unit]",
		},
	})
	if err != nil {
		t.Errorf("unexpected error for valid unit name: %v", err)
	}
}

func TestWorkerServerAndClient(t *testing.T) {
	tempDir := t.TempDir()
	executor := NewExecutor(tempDir, true)

	// Use TCP for portable testing across all OSes (Windows, Linux, macOS)
	addr := "127.0.0.1:29876"
	server := NewServer("tcp", addr, executor)

	go func() {
		_ = server.Start()
	}()
	defer server.Close()

	// Wait for server to listen
	time.Sleep(100 * time.Millisecond)

	client := NewClient("tcp", addr)
	defer client.Close()

	// 1. Write Nginx Config
	res, err := client.Call(ActionVhostWriteNginxConfig, map[string]string{
		"fqdn":    "test.example.com",
		"content": "server { listen 80; }",
	})
	if err != nil {
		t.Fatalf("client.Call failed: %v", err)
	}
	if !strings.Contains(res, "vhost written") {
		t.Errorf("unexpected data response: %s", res)
	}

	// Verify file was written inside sandbox
	expectedFile := filepath.Join(tempDir, "etc/wrenpanel/nginx/vhosts/test.example.com.conf")
	if !fileExists(expectedFile) {
		t.Errorf("expected vhost file to exist at %s", expectedFile)
	}

	// 2. Remove Nginx Config
	_, err = client.Call(ActionVhostRemoveNginxConfig, map[string]string{
		"fqdn": "test.example.com",
	})
	if err != nil {
		t.Fatalf("failed to remove config: %v", err)
	}
	if fileExists(expectedFile) {
		t.Errorf("expected vhost file to be removed at %s", expectedFile)
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
