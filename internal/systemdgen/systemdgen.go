package systemdgen

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"text/template"
)

type AppUnitOptions struct {
	Kind             string // node | python | docker
	Name             string
	User             string
	Group            string
	WorkingDirectory string
	ExecStart        string
	Environment      map[string]string
	InternalPort     int
}

const unitTemplate = `[Unit]
Description=WrenPanel managed {{ .Kind }} app: {{ .Name }}
After=network.target

[Service]
Type=simple
User={{ .User }}
Group={{ .Group }}
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .ExecStart }}
Restart=always
RestartSec=5s
Environment=PORT={{ .InternalPort }}
{{- range $key, $val := .Environment }}
Environment={{ $key }}={{ $val }}
{{- end }}

# Security hardening
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
`

// GenerateUnitName creates a collision-resistant unique unit name: wrenpanel-app-<name>-<short-hash>.service
func GenerateUnitName(kind, name, workingDir string) string {
	cleanName := strings.ToLower(strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, name))

	h := sha256.New()
	h.Write([]byte(kind + ":" + name + ":" + workingDir))
	hashStr := hex.EncodeToString(h.Sum(nil))[:6]

	return fmt.Sprintf("wrenpanel-app-%s-%s.service", cleanName, hashStr)
}

func GenerateUnitFile(opts AppUnitOptions) (string, error) {
	if opts.Name == "" {
		return "", fmt.Errorf("app name cannot be empty")
	}
	if opts.ExecStart == "" {
		return "", fmt.Errorf("ExecStart cannot be empty")
	}
	if opts.User == "" {
		opts.User = "nobody"
	}
	if opts.Group == "" {
		opts.Group = opts.User
	}

	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse unit template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("failed to execute unit template: %w", err)
	}

	return buf.String(), nil
}
