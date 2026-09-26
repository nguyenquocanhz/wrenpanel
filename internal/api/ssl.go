package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wrenpanel/internal/store"
	"wrenpanel/internal/vhostgen"
	"wrenpanel/internal/worker"
)

type IssueCertRequest struct {
	VhostID int64  `json:"vhost_id"`
	Email   string `json:"email"`
}

func (s *Server) handleListCerts(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListCerts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.Cert{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleIssueCert(w http.ResponseWriter, r *http.Request) {
	var req IssueCertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required for Let's Encrypt")
		return
	}

	vhost, err := s.store.GetVhostByID(req.VhostID)
	if err != nil || vhost == nil {
		writeError(w, http.StatusNotFound, "vhost not found")
		return
	}

	// 1. Verify DNS A record as mandated by GEMINI.md
	if s.acmeManager != nil {
		ok, ips, err := s.acmeManager.VerifyDNSResolution(vhost.FQDN)
		if err != nil || !ok {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("DNS verification failed: domain %s does not resolve to this server IP (%v)", vhost.FQDN, err))
			return
		}
		_ = ips
	}

	// 2. Issue certificate via Lego HTTP-01
	var certPath, keyPath string
	var issuedAtStr, expiresAtStr string

	if s.acmeManager != nil {
		res, err := s.acmeManager.IssueCertificateHTTP01(vhost.FQDN, req.Email, vhost.Docroot)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("ACME issuance failed: %v", err))
			return
		}
		certPath = res.CertPath
		keyPath = res.KeyPath
		issuedAtStr = res.IssuedAt.Format(time.RFC3339)
		expiresAtStr = res.ExpiresAt.Format(time.RFC3339)
	} else {
		// Mock paths if acmeManager is disabled
		certPath = fmt.Sprintf("/etc/wrenpanel/ssl/%s/cert.pem", vhost.FQDN)
		keyPath = fmt.Sprintf("/etc/wrenpanel/ssl/%s/privkey.pem", vhost.FQDN)
		now := time.Now().UTC()
		issuedAtStr = now.Format(time.RFC3339)
		expiresAtStr = now.Add(90 * 24 * time.Hour).Format(time.RFC3339)
	}

	// 3. Save to Store
	cert := &store.Cert{
		FQDN:      vhost.FQDN,
		CertPath:  certPath,
		KeyPath:   keyPath,
		Issuer:    "lets-encrypt",
		IssuedAt:  &issuedAtStr,
		ExpiresAt: &expiresAtStr,
		AutoRenew: 1,
	}

	if err := s.store.CreateCert(cert); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save cert record: %v", err))
		return
	}

	// 4. Update Vhost with ssl_cert_id
	vhost.SslCertID = &cert.ID
	_ = s.store.UpdateVhost(vhost)

	// 5. Regenerate Nginx Vhost with SSL configuration
	phpVerStr := "8.2"
	if vhost.PhpVersionID != nil {
		p, _ := s.store.GetPhpVersionByID(*vhost.PhpVersionID)
		if p != nil {
			phpVerStr = p.Version
		}
	}

	nginxConf, err := vhostgen.GenerateNginxConfig(vhostgen.VhostConfigOptions{
		FQDN:        vhost.FQDN,
		Docroot:     vhost.Docroot,
		WebEngine:   vhost.WebEngine,
		PhpVersion:  phpVerStr,
		SSLEnabled:  true,
		SSLCertPath: certPath,
		SSLKeyPath:  keyPath,
	})
	if err == nil && s.workerClient != nil {
		_, _ = s.workerClient.Call(worker.ActionVhostWriteNginxConfig, map[string]string{
			"fqdn":    vhost.FQDN,
			"content": nginxConf,
		})
		_, _ = s.workerClient.Call(worker.ActionNginxReload, nil)
	}

	target := vhost.FQDN
	_ = s.store.RecordAudit("admin", "ssl.issue", &target, nil, "success")
	writeJSON(w, http.StatusCreated, cert)
}
