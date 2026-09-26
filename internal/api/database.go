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

type CreateDatabaseRequest struct {
	AccountID int64  `json:"account_id"`
	VhostID   *int64 `json:"vhost_id"`
	Engine    string `json:"engine"` // mysql | postgresql | mongodb
	DBName    string `json:"db_name"`
	DBUser    string `json:"db_user"`
	DBPass    string `json:"db_password"`
}

func (s *Server) handleListDatabases(w http.ResponseWriter, r *http.Request) {
	var accountID *int64
	if accStr := r.URL.Query().Get("account_id"); accStr != "" {
		if id, err := strconv.ParseInt(accStr, 10, 64); err == nil {
			accountID = &id
		}
	}

	list, err := s.store.ListDatabases(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.Database{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req CreateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.DBName = strings.TrimSpace(req.DBName)
	req.DBUser = strings.TrimSpace(req.DBUser)
	if req.DBName == "" || req.DBUser == "" {
		writeError(w, http.StatusBadRequest, "db_name and db_user are required")
		return
	}
	if req.Engine == "" {
		req.Engine = "mysql"
	}

	acc, err := s.store.GetAccountByID(req.AccountID)
	if err != nil || acc == nil {
		writeError(w, http.StatusBadRequest, "valid account is required")
		return
	}

	// Prefix db name and db user with username_ as standard cPanel hosting does
	if !strings.HasPrefix(req.DBName, acc.Username+"_") {
		req.DBName = fmt.Sprintf("%s_%s", acc.Username, req.DBName)
	}
	if !strings.HasPrefix(req.DBUser, acc.Username+"_") {
		req.DBUser = fmt.Sprintf("%s_%s", acc.Username, req.DBUser)
	}

	db := &store.Database{
		AccountID: req.AccountID,
		VhostID:   req.VhostID,
		Engine:    req.Engine,
		DBName:    req.DBName,
		DBUser:    req.DBUser,
	}

	if err := s.store.CreateDatabase(db); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := req.DBName
	_ = s.store.RecordAudit("admin", "database.create", &target, nil, "success")
	writeJSON(w, http.StatusCreated, db)
}

func (s *Server) handleDeleteDatabase(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	db, err := s.store.GetDatabaseByID(id)
	if err != nil || db == nil {
		writeError(w, http.StatusNotFound, "database not found")
		return
	}

	if err := s.store.DeleteDatabase(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := db.DBName
	_ = s.store.RecordAudit("admin", "database.delete", &target, nil, "success")
	writeJSON(w, http.StatusOK, map[string]string{"message": "database deleted"})
}
