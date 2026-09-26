package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"wrenpanel/internal/store"
	"wrenpanel/internal/vhostgen"
	"wrenpanel/internal/worker"
)

type CreateVhostRequest struct {
	AccountID      int64   `json:"account_id"`
	ParentVhostID  *int64  `json:"parent_vhost_id"`
	Type           string  `json:"type"` // primary | subdomain | addon
	FQDN           string  `json:"fqdn"`
	Docroot        string  `json:"docroot"`
	DocrootOwnerID *int64  `json:"docroot_owner_id"`
	WebEngine      string  `json:"web_engine"` // nginx | nginx+apache
	PhpVersionID   *int64  `json:"php_version_id"`
}

func (s *Server) handleListVhosts(w http.ResponseWriter, r *http.Request) {
	var accountID *int64
	if accStr := r.URL.Query().Get("account_id"); accStr != "" {
		if id, err := strconv.ParseInt(accStr, 10, 64); err == nil {
			accountID = &id
		}
	}

	list, err := s.store.ListVhosts(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.Vhost{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetVhost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vhost id")
		return
	}

	v, err := s.store.GetVhostByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if v == nil {
		writeError(w, http.StatusNotFound, "vhost not found")
		return
	}

	writeJSON(w, http.StatusOK, v)
}

func (s *Server) handleCreateVhost(w http.ResponseWriter, r *http.Request) {
	var req CreateVhostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	req.FQDN = strings.TrimSpace(strings.ToLower(req.FQDN))
	if req.FQDN == "" {
		writeError(w, http.StatusBadRequest, "fqdn is required")
		return
	}

	// 1. Reserved prefix validation
	parts := strings.Split(req.FQDN, ".")
	if len(parts) > 2 {
		prefix := parts[0]
		if store.IsReservedPrefix(prefix) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Prefix '%s' is reserved for system services (mail, ftp, wrenpanel, etc.) and cannot be used.", prefix))
			return
		}
	}

	// 2. Check duplicate FQDN
	existing, err := s.store.GetVhostByFQDN(req.FQDN)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing != nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("Domain/vhost '%s' already exists", req.FQDN))
		return
	}

	// 3. Verify Account
	acc, err := s.store.GetAccountByID(req.AccountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if acc == nil {
		writeError(w, http.StatusBadRequest, "account does not exist")
		return
	}

	// 4. Default docroot if empty
	if req.Docroot == "" {
		if req.Type == "primary" {
			req.Docroot = fmt.Sprintf("%s/public_html", acc.HomeDir)
		} else {
			req.Docroot = fmt.Sprintf("%s/public_html/%s", acc.HomeDir, parts[0])
		}
	}

	if req.WebEngine == "" {
		req.WebEngine = "nginx"
	}

	// 5. Lookup PHP version info if selected
	phpVerStr := "8.2"
	if req.PhpVersionID != nil {
		phpVer, _ := s.store.GetPhpVersionByID(*req.PhpVersionID)
		if phpVer != nil {
			phpVerStr = phpVer.Version
		}
	}

	// 6. Generate Vhost config
	apachePort := 8080
	if req.WebEngine == "nginx+apache" {
		apachePort = 8081 // or allocated port
	}

	nginxConf, err := vhostgen.GenerateNginxConfig(vhostgen.VhostConfigOptions{
		FQDN:       req.FQDN,
		Docroot:    req.Docroot,
		WebEngine:  req.WebEngine,
		PhpVersion: phpVerStr,
		ApachePort: apachePort,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate nginx config: %v", err))
		return
	}

	// 7. Execute worker commands
	if s.workerClient != nil {
		_, _ = s.workerClient.Call(worker.ActionDirEnsure, map[string]string{
			"path": req.Docroot,
		})
		_, err = s.workerClient.Call(worker.ActionVhostWriteNginxConfig, map[string]string{
			"fqdn":    req.FQDN,
			"content": nginxConf,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("worker failed to write config: %v", err))
			return
		}
		_, _ = s.workerClient.Call(worker.ActionNginxReload, nil)
	}

	// 8. Insert into Store
	vhost := &store.Vhost{
		AccountID:      req.AccountID,
		ParentVhostID:  req.ParentVhostID,
		Type:           req.Type,
		FQDN:           req.FQDN,
		Docroot:        req.Docroot,
		DocrootOwnerID: req.DocrootOwnerID,
		WebEngine:      req.WebEngine,
		PhpVersionID:   req.PhpVersionID,
		Status:         "active",
	}

	if err := s.store.CreateVhost(vhost); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		return
	}

	// 9. Audit log
	target := req.FQDN
	payload := fmt.Sprintf(`{"fqdn": "%s", "docroot": "%s", "web_engine": "%s"}`, req.FQDN, req.Docroot, req.WebEngine)
	_ = s.store.RecordAudit("admin", "vhost.create", &target, &payload, "success")

	writeJSON(w, http.StatusCreated, vhost)
}

func (s *Server) handleVhostDeletionImpact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vhost id")
		return
	}

	impact, err := s.store.AnalyzeDeletionImpact(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, impact)
}

func (s *Server) handleDeleteVhost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vhost id")
		return
	}

	deleteFiles := r.URL.Query().Get("delete_files") == "true"

	v, err := s.store.GetVhostByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if v == nil {
		writeError(w, http.StatusNotFound, "vhost not found")
		return
	}

	// 1. Analyze safe deletion impact
	impact, err := s.store.AnalyzeDeletionImpact(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !impact.AllowDelete {
		writeError(w, http.StatusForbidden, impact.ReasonIfBlocked)
		return
	}

	// 2. Remove Nginx config via worker
	if s.workerClient != nil {
		_, _ = s.workerClient.Call(worker.ActionVhostRemoveNginxConfig, map[string]string{
			"fqdn": v.FQDN,
		})
		_, _ = s.workerClient.Call(worker.ActionNginxReload, nil)

		// 3. Remove files ONLY IF allowed and user requested
		if deleteFiles && impact.CanDeleteFiles {
			_, _ = s.workerClient.Call(worker.ActionDirRemove, map[string]string{
				"path": v.Docroot,
			})
		}
	}

	// 4. Delete DB record
	if err := s.store.DeleteVhost(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 5. Audit log
	target := v.FQDN
	payload := fmt.Sprintf(`{"fqdn": "%s", "deleted_files": %t}`, v.FQDN, deleteFiles && impact.CanDeleteFiles)
	_ = s.store.RecordAudit("admin", "vhost.delete", &target, &payload, "success")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"deleted_files": deleteFiles && impact.CanDeleteFiles,
		"message":       fmt.Sprintf("Vhost %s deleted successfully", v.FQDN),
	})
}
