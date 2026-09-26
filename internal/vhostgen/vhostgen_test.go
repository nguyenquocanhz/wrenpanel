package vhostgen

import (
	"strings"
	"testing"
)

func TestGenerateNginxConfig_DirectPhp(t *testing.T) {
	opts := VhostConfigOptions{
		FQDN:       "example.com",
		Docroot:    "/home/user1/public_html",
		WebEngine:  "nginx",
		PhpVersion: "8.2",
	}

	conf, err := GenerateNginxConfig(opts)
	if err != nil {
		t.Fatalf("GenerateNginxConfig failed: %v", err)
	}

	if !strings.Contains(conf, "server_name example.com;") {
		t.Errorf("expected server_name example.com")
	}
	if !strings.Contains(conf, "root /home/user1/public_html;") {
		t.Errorf("expected docroot set")
	}
	if !strings.Contains(conf, "fastcgi_pass unix:/run/php/php8.2-fpm.sock;") {
		t.Errorf("expected php-fpm socket")
	}
}

func TestGenerateNginxConfig_WithApache(t *testing.T) {
	opts := VhostConfigOptions{
		FQDN:       "example.com",
		Docroot:    "/home/user1/public_html",
		WebEngine:  "nginx+apache",
		ApachePort: 8081,
	}

	conf, err := GenerateNginxConfig(opts)
	if err != nil {
		t.Fatalf("GenerateNginxConfig failed: %v", err)
	}

	if !strings.Contains(conf, "proxy_pass http://127.0.0.1:8081;") {
		t.Errorf("expected proxy_pass to apache on 127.0.0.1:8081")
	}

	apacheConf, err := GenerateApacheConfig(opts)
	if err != nil {
		t.Fatalf("GenerateApacheConfig failed: %v", err)
	}

	if !strings.Contains(apacheConf, "<VirtualHost 127.0.0.1:8081>") {
		t.Errorf("expected VirtualHost on 127.0.0.1:8081")
	}
}
