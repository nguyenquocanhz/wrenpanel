package systemdgen

import (
	"strings"
	"testing"
)

func TestGenerateUnitFile(t *testing.T) {
	opts := AppUnitOptions{
		Kind:             "node",
		Name:             "my-blog",
		User:             "john",
		Group:            "john",
		WorkingDirectory: "/home/john/apps/blog",
		ExecStart:        "/opt/wrenpanel/node/20.11.0/bin/node server.js",
		InternalPort:     3000,
		Environment: map[string]string{
			"NODE_ENV": "production",
		},
	}

	content, err := GenerateUnitFile(opts)
	if err != nil {
		t.Fatalf("GenerateUnitFile failed: %v", err)
	}

	if !strings.Contains(content, "ExecStart=/opt/wrenpanel/node/20.11.0/bin/node server.js") {
		t.Errorf("ExecStart missing or incorrect")
	}
	if !strings.Contains(content, "Environment=PORT=3000") {
		t.Errorf("PORT env missing")
	}
	if !strings.Contains(content, "Environment=NODE_ENV=production") {
		t.Errorf("NODE_ENV missing")
	}

	unitName := GenerateUnitName(opts.Kind, opts.Name, opts.WorkingDirectory)
	if !strings.HasPrefix(unitName, "wrenpanel-app-my-blog-") || !strings.HasSuffix(unitName, ".service") {
		t.Errorf("Unexpected unit name format: %s", unitName)
	}
}
