package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"wrenpanel/internal/store"
)

// --- Node.js Handlers ---

func (s *Server) handleListNodeVersions(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListNodeVersions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.NodeVersion{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateNodeVersion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version     string `json:"version"`      // e.g. "20.11.0"
		InstallPath string `json:"install_path"` // e.g. "/opt/wrenpanel/node/20.11.0"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Version = strings.TrimSpace(req.Version)
	if req.Version == "" {
		writeError(w, http.StatusBadRequest, "version is required")
		return
	}
	if req.InstallPath == "" {
		req.InstallPath = fmt.Sprintf("/opt/wrenpanel/node/%s", req.Version)
	}

	n := &store.NodeVersion{
		Version:     req.Version,
		InstallPath: req.InstallPath,
	}

	if err := s.store.CreateNodeVersion(n); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := req.Version
	_ = s.store.RecordAudit("admin", "node.create_version", &target, nil, "success")
	writeJSON(w, http.StatusCreated, n)
}

func (s *Server) handleDeleteNodeVersion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := s.store.DeleteNodeVersion(id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "node version removed"})
}

// --- Python Handlers ---

func (s *Server) handleListPythonVersions(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListPythonVersions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.PythonVersion{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreatePythonVersion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version         string `json:"version"`          // e.g. "3.11"
		InterpreterPath string `json:"interpreter_path"` // e.g. "/opt/wrenpanel/python/3.11/bin/python3"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Version = strings.TrimSpace(req.Version)
	if req.Version == "" {
		writeError(w, http.StatusBadRequest, "version is required")
		return
	}
	if req.InterpreterPath == "" {
		req.InterpreterPath = fmt.Sprintf("/opt/wrenpanel/python/%s/bin/python3", req.Version)
	}

	p := &store.PythonVersion{
		Version:         req.Version,
		InterpreterPath: req.InterpreterPath,
	}

	if err := s.store.CreatePythonVersion(p); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := req.Version
	_ = s.store.RecordAudit("admin", "python.create_version", &target, nil, "success")
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleDeletePythonVersion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := s.store.DeletePythonVersion(id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "python version removed"})
}
