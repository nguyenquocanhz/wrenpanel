package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"wrenpanel/internal/store"
	"wrenpanel/internal/worker"
)

type CreateAccountRequest struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"` // Optional: if empty, will auto-generate
	Email       *string `json:"email"`
	DiskQuotaMB int64   `json:"disk_quota_mb"`
}

type CreateAccountResponse struct {
	store.Account
	InitialPassword string `json:"initial_password"`
}

type ResetPasswordRequest struct {
	Password string `json:"password"` // Optional: if empty, will auto-generate
}

func generateRandomPassword(length int) string {
	const charset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "WrenPanel@2026!"
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
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

	plainPassword := strings.TrimSpace(req.Password)
	if plainPassword == "" {
		plainPassword = generateRandomPassword(16)
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	hashStr := string(hashBytes)

	homeDir := fmt.Sprintf("/home/%s", req.Username)

	// Call worker to create system user and standard cPanel-style directories
	if s.workerClient != nil {
		_, err := s.workerClient.Call(worker.ActionUserCreate, map[string]string{
			"username": req.Username,
			"home_dir": homeDir,
			"password": plainPassword,
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
		Username:     req.Username,
		HomeDir:      homeDir,
		Email:        req.Email,
		DiskQuotaMB:  req.DiskQuotaMB,
		Status:       "active",
		PasswordHash: &hashStr,
	}

	if err := s.store.CreateAccount(acc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	target := req.Username
	_ = s.store.RecordAudit("admin", "account.create", &target, nil, "success")

	resp := CreateAccountResponse{
		Account:         *acc,
		InitialPassword: plainPassword,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleResetAccountPassword(w http.ResponseWriter, r *http.Request) {
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

	var req ResetPasswordRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	newPassword := strings.TrimSpace(req.Password)
	if newPassword == "" {
		newPassword = generateRandomPassword(16)
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := s.store.UpdateAccountPassword(id, string(hashBytes)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if s.workerClient != nil {
		_, err := s.workerClient.Call(worker.ActionUserSetPassword, map[string]string{
			"username": acc.Username,
			"password": newPassword,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update linux password: %v", err))
			return
		}
	}

	target := acc.Username
	_ = s.store.RecordAudit("admin", "account.reset_password", &target, nil, "success")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Password reset successfully",
		"username":     acc.Username,
		"new_password": newPassword,
	})
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
