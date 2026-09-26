package vhostgen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type VhostConfigOptions struct {
	FQDN         string
	Docroot      string
	WebEngine    string // "nginx" or "nginx+apache"
	PhpVersion   string // e.g. "8.2"
	PhpSocket    string // e.g. "/run/php/php8.2-fpm-sitename.sock"
	ApachePort   int    // e.g. 8081
	SSLEnabled   bool
	SSLCertPath  string
	SSLKeyPath   string
	AccessLog    string
	ErrorLog     string
}

const nginxTemplate = `# WrenPanel generated vhost for {{ .FQDN }}
# Source of truth: /etc/wrenpanel/nginx/vhosts/{{ .FQDN }}.conf
# DO NOT edit directly in /etc/nginx/sites-enabled/

{{- if .SSLEnabled }}
server {
    listen 80;
    server_name {{ .FQDN }};
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name {{ .FQDN }};

    ssl_certificate {{ .SSLCertPath }};
    ssl_certificate_key {{ .SSLKeyPath }};
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
{{- else }}
server {
    listen 80;
    server_name {{ .FQDN }};
{{- end }}

    root {{ .Docroot }};
    index index.php index.html index.htm;

    access_log {{ if .AccessLog }}{{ .AccessLog }}{{ else }}/var/log/nginx/{{ .FQDN }}_access.log{{ end }};
    error_log {{ if .ErrorLog }}{{ .ErrorLog }}{{ else }}/var/log/nginx/{{ .FQDN }}_error.log{{ end }};

{{- if eq .WebEngine "nginx+apache" }}
    # Reverse proxy to internal Apache backend
    location / {
        try_files $uri @apache;
    }

    location @apache {
        proxy_pass http://127.0.0.1:{{ .ApachePort }};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
{{- else }}
    # Direct Nginx + PHP-FPM
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        try_files $uri =404;
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass unix:{{ if .PhpSocket }}{{ .PhpSocket }}{{ else }}/run/php/php{{ .PhpVersion }}-fpm.sock{{ end }};
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param PATH_INFO $fastcgi_path_info;
    }
{{- end }}

    location ~ /\.ht {
        deny all;
    }

    location /.well-known/acme-challenge/ {
        root {{ .Docroot }};
    }
}
`

const apacheTemplate = `# WrenPanel generated Apache internal vhost for {{ .FQDN }}
# Binds strictly to 127.0.0.1
<VirtualHost 127.0.0.1:{{ .ApachePort }}>
    ServerName {{ .FQDN }}
    DocumentRoot {{ .Docroot }}

    <Directory {{ .Docroot }}>
        Options -Indexes +FollowSymLinks
        AllowOverride All
        Require all granted
    </Directory>

    ErrorLog ${APACHE_LOG_DIR}/{{ .FQDN }}_error.log
    CustomLog ${APACHE_LOG_DIR}/{{ .FQDN }}_access.log combined
</VirtualHost>
`

func GenerateNginxConfig(opts VhostConfigOptions) (string, error) {
	if strings.TrimSpace(opts.FQDN) == "" {
		return "", fmt.Errorf("fqdn cannot be empty")
	}
	if strings.TrimSpace(opts.Docroot) == "" {
		return "", fmt.Errorf("docroot cannot be empty")
	}

	tmpl, err := template.New("nginx").Parse(nginxTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse nginx template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("failed to execute nginx template: %w", err)
	}

	return buf.String(), nil
}

func GenerateApacheConfig(opts VhostConfigOptions) (string, error) {
	if opts.ApachePort <= 0 {
		return "", fmt.Errorf("invalid apache port %d", opts.ApachePort)
	}

	tmpl, err := template.New("apache").Parse(apacheTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse apache template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("failed to execute apache template: %w", err)
	}

	return buf.String(), nil
}
