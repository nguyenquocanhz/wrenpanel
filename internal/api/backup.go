package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"wrenpanel/internal/store"
)

type CreateBackupRequest struct {
	AccountID   int64  `json:"account_id"`
	Scope       string `json:"scope"` // site | database | full
	VhostID     *int64 `json:"vhost_id"`
	Destination string `json:"destination"` // local | s3 | ftp | rsync
}

func (s *Server) handleListBackups(w http.ResponseWriter, r *http.Request) {
	var accountID *int64
	if accStr := r.URL.Query().Get("account_id"); accStr != "" {
		if id, err := strconv.ParseInt(accStr, 10, 64); err == nil {
			accountID = &id
		}
	}

	list, err := s.store.ListBackups(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.Backup{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateBackup(w http.ResponseWriter, r *http.Request) {
	var req CreateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Scope == "" {
		req.Scope = "site"
	}
	if req.Destination == "" {
		req.Destination = "local"
	}

	acc, err := s.store.GetAccountByID(req.AccountID)
	if err != nil || acc == nil {
		writeError(w, http.StatusBadRequest, "valid account is required")
		return
	}

	// Format: /var/lib/wrenpanel/backups/staging/<account>_<scope>_<timestamp>.tar.zst
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("/var/lib/wrenpanel/backups/staging/%s_%s_%s.tar.zst", acc.Username, req.Scope, timestamp)
	manifest := fmt.Sprintf(`{"account": "%s", "scope": "%s", "created_at": "%s"}`, acc.Username, req.Scope, time.Now().UTC().Format(time.RFC3339))

	backup := &store.Backup{
		AccountID:    req.AccountID,
		Scope:        req.Scope,
		VhostID:      req.VhostID,
		FilePath:     backupFile,
		Destination:  req.Destination,
		ManifestJSON: &manifest,
	}

	if err := s.store.CreateBackup(backup); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := backupFile
	_ = s.store.RecordAudit("admin", "backup.create", &target, nil, "success")
	writeJSON(w, http.StatusCreated, backup)
}
