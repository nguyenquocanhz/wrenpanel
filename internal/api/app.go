package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"wrenpanel/internal/store"
	"wrenpanel/internal/systemdgen"
	"wrenpanel/internal/worker"
)

type CreateAppRequest struct {
	AccountID        int64             `json:"account_id"`
	VhostID          *int64            `json:"vhost_id"`
	Kind             string            `json:"kind"` // node | python | docker
	Name             string            `json:"name"`
	RuntimeVersion   *string           `json:"runtime_version"`
	Entrypoint       *string           `json:"entrypoint"`
	WorkingDirectory string            `json:"working_directory"`
	Environment      map[string]string `json:"environment"`
}

func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	var accountID *int64
	if accStr := r.URL.Query().Get("account_id"); accStr != "" {
		if id, err := strconv.ParseInt(accStr, 10, 64); err == nil {
			accountID = &id
		}
	}
	var kind *string
	if k := r.URL.Query().Get("kind"); k != "" {
		kind = &k
	}

	list, err := s.store.ListApps(accountID, kind)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.App{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetApp(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid app id")
		return
	}

	app, err := s.store.GetAppByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	var req CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Kind != "node" && req.Kind != "python" && req.Kind != "docker" {
		writeError(w, http.StatusBadRequest, "kind must be 'node', 'python', or 'docker'")
		return
	}

	acc, err := s.store.GetAccountByID(req.AccountID)
	if err != nil || acc == nil {
		writeError(w, http.StatusBadRequest, "valid account is required")
		return
	}

	if req.WorkingDirectory == "" {
		req.WorkingDirectory = fmt.Sprintf("%s/apps/%s", acc.HomeDir, req.Name)
	}

	// 1. First insert App record so we have an AppID for port reservation foreign key
	app := &store.App{
		AccountID:      req.AccountID,
		VhostID:        req.VhostID,
		Kind:           req.Kind,
		Name:           req.Name,
		RuntimeVersion: req.RuntimeVersion,
		Entrypoint:     req.Entrypoint,
		Status:         "stopped",
	}

	if err := s.store.CreateApp(app); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save app: %v", err))
		return
	}

	// 2. Allocate internal port automatically
	port, err := s.store.AllocatePort(app.ID, 10000, 20000)
	if err != nil {
		_ = s.store.DeleteApp(app.ID)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to allocate internal port: %v", err))
		return
	}
	app.InternalPort = &port

	// 3. Generate ExecStart depending on runtime
	entry := "index.js"
	if req.Entrypoint != nil && *req.Entrypoint != "" {
		entry = *req.Entrypoint
	}

	var execStart string
	switch req.Kind {
	case "node":
		nodeBin := "node"
		if req.RuntimeVersion != nil && *req.RuntimeVersion != "" {
			nodeBin = fmt.Sprintf("/opt/wrenpanel/node/%s/bin/node", *req.RuntimeVersion)
		}
		execStart = fmt.Sprintf("%s %s", nodeBin, entry)
	case "python":
		pyBin := "python3"
		if req.RuntimeVersion != nil && *req.RuntimeVersion != "" {
			pyBin = fmt.Sprintf("/opt/wrenpanel/python/%s/bin/python3", *req.RuntimeVersion)
		}
		execStart = fmt.Sprintf("%s %s", pyBin, entry)
	case "docker":
		execStart = fmt.Sprintf("/usr/bin/docker compose -f %s/compose.yml up", req.WorkingDirectory)
	}

	// 4. Generate systemd unit file
	unitName := systemdgen.GenerateUnitName(req.Kind, req.Name, req.WorkingDirectory)
	app.SystemdUnit = &unitName

	unitContent, err := systemdgen.GenerateUnitFile(systemdgen.AppUnitOptions{
		Kind:             req.Kind,
		Name:             req.Name,
		User:             acc.Username,
		Group:            acc.Username,
		WorkingDirectory: req.WorkingDirectory,
		ExecStart:        execStart,
		Environment:      req.Environment,
		InternalPort:     port,
	})
	if err != nil {
		_ = s.store.DeleteApp(app.ID)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate systemd unit: %v", err))
		return
	}

	// 5. Send unit file to worker and start
	if s.workerClient != nil {
		_, err = s.workerClient.Call(worker.ActionSystemdWriteUnit, map[string]string{
			"unit_name": unitName,
			"content":   unitContent,
		})
		if err != nil {
			_ = s.store.DeleteApp(app.ID)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("worker failed to write unit file: %v", err))
			return
		}

		_, _ = s.workerClient.Call(worker.ActionSystemdAction, map[string]string{
			"unit_name": unitName,
			"action":    "start",
		})
		app.Status = "running"
	}

	_ = s.store.UpdateApp(app)

	target := unitName
	_ = s.store.RecordAudit("admin", "app.create", &target, nil, "success")
	writeJSON(w, http.StatusCreated, app)
}

func (s *Server) handleAppAction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var body struct {
		Action string `json:"action"` // start | stop | restart
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	app, err := s.store.GetAppByID(id)
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}

	if app.SystemdUnit == nil {
		writeError(w, http.StatusBadRequest, "app has no associated systemd unit")
		return
	}

	if s.workerClient != nil {
		_, err := s.workerClient.Call(worker.ActionSystemdAction, map[string]string{
			"unit_name": *app.SystemdUnit,
			"action":    body.Action,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("action failed: %v", err))
			return
		}
	}

	if body.Action == "stop" {
		app.Status = "stopped"
	} else {
		app.Status = "running"
	}
	_ = s.store.UpdateAppStatus(app.ID, app.Status)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Action '%s' performed successfully", body.Action),
		"status":  app.Status,
	})
}

func (s *Server) handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	app, err := s.store.GetAppByID(id)
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}

	if app.SystemdUnit != nil && s.workerClient != nil {
		_, _ = s.workerClient.Call(worker.ActionSystemdRemoveUnit, map[string]string{
			"unit_name": *app.SystemdUnit,
		})
	}

	if err := s.store.DeleteApp(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := app.Name
	_ = s.store.RecordAudit("admin", "app.delete", &target, nil, "success")
	writeJSON(w, http.StatusOK, map[string]string{"message": "app deleted"})
}
