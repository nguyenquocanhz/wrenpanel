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

type CreateAccountRequest struct {
	Username    string  `json:"username"`
	Email       *string `json:"email"`
	DiskQuotaMB int64   `json:"disk_quota_mb"`
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListAccounts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.Account{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	// Verify not existing
	existing, _ := s.store.GetAccountByUsername(req.Username)
	if existing != nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("Account '%s' already exists", req.Username))
		return
	}

	homeDir := fmt.Sprintf("/home/%s", req.Username)

	// Call worker to create system user and standard cPanel-style directories
	if s.workerClient != nil {
		_, err := s.workerClient.Call(worker.ActionUserCreate, map[string]string{
			"username": req.Username,
			"home_dir": homeDir,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("worker failed to create user: %v", err))
			return
		}

		// Ensure standard cPanel directories as specified in GEMINI.md:
		// public_html, mail, logs, tmp
		dirs := []string{"public_html", "mail", "logs", "ssl", "tmp"}
		for _, d := range dirs {
			_, _ = s.workerClient.Call(worker.ActionDirEnsure, map[string]string{
				"path": fmt.Sprintf("%s/%s", homeDir, d),
			})
		}
	}

	acc := &store.Account{
		Username:    req.Username,
		HomeDir:     homeDir,
		Email:       req.Email,
		DiskQuotaMB: req.DiskQuotaMB,
		Status:      "active",
	}

	if err := s.store.CreateAccount(acc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := req.Username
	_ = s.store.RecordAudit("admin", "account.create", &target, nil, "success")
	writeJSON(w, http.StatusCreated, acc)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	acc, err := s.store.GetAccountByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if acc == nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	// Delete user via worker
	if s.workerClient != nil {
		_, _ = s.workerClient.Call(worker.ActionUserDelete, map[string]string{
			"username": acc.Username,
		})
	}

	if err := s.store.DeleteAccount(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := acc.Username
	_ = s.store.RecordAudit("admin", "account.delete", &target, nil, "success")
	writeJSON(w, http.StatusOK, map[string]string{"message": "account deleted"})
}
