package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"wrenpanel/internal/store"
	"wrenpanel/internal/worker"
)

type CreatePhpRequest struct {
	Version    string `json:"version"`     // "8.2"
	FpmService string `json:"fpm_service"`  // "php8.2-fpm"
	BinaryPath string `json:"binary_path"`
}

func (s *Server) handleListPhpVersions(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListPhpVersions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.PhpVersion{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreatePhpVersion(w http.ResponseWriter, r *http.Request) {
	var req CreatePhpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Version = strings.TrimSpace(req.Version)
	if req.Version == "" {
		writeError(w, http.StatusBadRequest, "version is required (e.g., 8.2)")
		return
	}
	if req.FpmService == "" {
		req.FpmService = fmt.Sprintf("php%s-fpm", req.Version)
	}
	if req.BinaryPath == "" {
		req.BinaryPath = fmt.Sprintf("/usr/bin/php%s", req.Version)
	}

	php := &store.PhpVersion{
		Version:    req.Version,
		FpmService: req.FpmService,
		BinaryPath: req.BinaryPath,
	}

	if err := s.store.CreatePhpVersion(php); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := req.Version
	_ = s.store.RecordAudit("admin", "php.create_version", &target, nil, "success")
	writeJSON(w, http.StatusCreated, php)
}

func (s *Server) handleDeletePhpVersion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid php version id")
		return
	}

	php, err := s.store.GetPhpVersionByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if php == nil {
		writeError(w, http.StatusNotFound, "php version not found")
		return
	}

	// Enforce rule: "Gỡ version: chặn nếu còn site đang gán version đó, chỉ cho gỡ khi 0 tham chiếu"
	if err := s.store.DeletePhpVersion(id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	target := php.Version
	_ = s.store.RecordAudit("admin", "php.delete_version", &target, nil, "success")
	writeJSON(w, http.StatusOK, map[string]string{"message": "php version removed"})
}

func (s *Server) handleReloadPhp(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version string `json:"version"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.Version == "" {
		writeError(w, http.StatusBadRequest, "version is required")
		return
	}

	if s.workerClient != nil {
		_, err := s.workerClient.Call(worker.ActionPhpReload, map[string]string{
			"version": body.Version,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to reload php: %v", err))
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("PHP %s reloaded", body.Version),
	})
}
