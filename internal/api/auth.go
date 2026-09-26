package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"` // admin | tenant
	HomeDir  string `json:"home_dir,omitempty"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

type TokenPayload struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	HomeDir   string    `json:"home_dir,omitempty"`
	ExpiresAt time.Time `json:"exp"`
}

var serverSecret = []byte("wrenpanel-secret-session-key-2026")

func generateToken(user AuthUser) (string, error) {
	payload := TokenPayload{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		HomeDir:   user.HomeDir,
		ExpiresAt: time.Now().Add(24 * 7 * time.Hour), // 7 days
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(data)
	mac := hmac.New(sha256.New, serverSecret)
	mac.Write([]byte(payloadB64))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", payloadB64, sig), nil
}

func parseToken(tokenStr string) (*TokenPayload, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}
	payloadB64, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, serverSecret)
	mac.Write([]byte(payloadB64))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid signature")
	}

	data, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}
	var payload TokenPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if time.Now().After(payload.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}
	return &payload, nil
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	// 1. Check if admin
	admin, err := s.store.GetAdminByUsername(req.Username)
	if err == nil && admin != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err == nil {
			user := AuthUser{
				ID:       admin.ID,
				Username: admin.Username,
				Role:     "admin",
			}
			token, _ := generateToken(user)
			writeJSON(w, http.StatusOK, LoginResponse{Token: token, User: user})
			return
		}
	}

	// 2. Check if hosting account (tenant)
	acc, err := s.store.GetAccountByUsername(req.Username)
	if err == nil && acc != nil && acc.Status == "active" && acc.PasswordHash != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*acc.PasswordHash), []byte(req.Password)); err == nil {
			user := AuthUser{
				ID:       acc.ID,
				Username: acc.Username,
				Role:     "tenant",
				HomeDir:  acc.HomeDir,
			}
			token, _ := generateToken(user)
			writeJSON(w, http.StatusOK, LoginResponse{Token: token, User: user})
			return
		}
	}

	writeError(w, http.StatusUnauthorized, "Tên đăng nhập hoặc mật khẩu không chính xác")
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	payload, err := parseToken(tokenStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	writeJSON(w, http.StatusOK, AuthUser{
		ID:       payload.ID,
		Username: payload.Username,
		Role:     payload.Role,
		HomeDir:  payload.HomeDir,
	})
}
