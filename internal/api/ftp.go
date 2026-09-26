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

type CreateFtpRequest struct {
	AccountID int64  `json:"account_id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	RootDir   string `json:"root_dir"`
}

func (s *Server) handleListFtp(w http.ResponseWriter, r *http.Request) {
	var accountID *int64
	if accStr := r.URL.Query().Get("account_id"); accStr != "" {
		if id, err := strconv.ParseInt(accStr, 10, 64); err == nil {
			accountID = &id
		}
	}

	list, err := s.store.ListFtpAccounts(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.FtpAccount{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateFtp(w http.ResponseWriter, r *http.Request) {
	var req CreateFtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	acc, err := s.store.GetAccountByID(req.AccountID)
	if err != nil || acc == nil {
		writeError(w, http.StatusBadRequest, "valid account is required")
		return
	}

	// Prefix FTP user with account username: e.g. client1_dev
	if !strings.HasPrefix(req.Username, acc.Username+"_") {
		req.Username = fmt.Sprintf("%s_%s", acc.Username, req.Username)
	}

	if req.RootDir == "" {
		req.RootDir = fmt.Sprintf("%s/public_html", acc.HomeDir)
	}

	// Call worker to create FTP virtual account and chroot directory
	if s.workerClient != nil {
		_, _ = s.workerClient.Call(worker.ActionDirEnsure, map[string]string{
			"path": req.RootDir,
		})
		_, err := s.workerClient.Call(worker.ActionFtpUserCreate, map[string]string{
			"username": req.Username,
			"root_dir": req.RootDir,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("worker failed to create ftp user: %v", err))
			return
		}
	}

	ftp := &store.FtpAccount{
		AccountID: req.AccountID,
		Username:  req.Username,
		RootDir:   req.RootDir,
		Status:    "active",
	}

	if err := s.store.CreateFtpAccount(ftp); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := req.Username
	_ = s.store.RecordAudit("admin", "ftp.create_account", &target, nil, "success")
	writeJSON(w, http.StatusCreated, ftp)
}

func (s *Server) handleDeleteFtp(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ftp, err := s.store.GetFtpAccountByID(id)
	if err != nil || ftp == nil {
		writeError(w, http.StatusNotFound, "ftp account not found")
		return
	}

	if s.workerClient != nil {
		_, _ = s.workerClient.Call(worker.ActionFtpUserDelete, map[string]string{
			"username": ftp.Username,
		})
	}

	if err := s.store.DeleteFtpAccount(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := ftp.Username
	_ = s.store.RecordAudit("admin", "ftp.delete_account", &target, nil, "success")
	writeJSON(w, http.StatusOK, map[string]string{"message": "ftp account deleted"})
}
