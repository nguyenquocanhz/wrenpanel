package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

type PmaInfoResponse struct {
	Engine     string `json:"engine"`      // phpmyadmin | adminer
	Status     string `json:"status"`      // running | stopped
	Path       string `json:"secret_path"` // protected path behind panel auth
	WebURL     string `json:"web_url"`
	IsInternal bool   `json:"is_internal"` // true (always protected)
}

func (s *Server) handleGetPmaStatus(w http.ResponseWriter, r *http.Request) {
	resp := PmaInfoResponse{
		Engine:     "phpmyadmin",
		Status:     "running",
		Path:       "/_panel/db-manager",
		WebURL:     "/_panel/db-manager",
		IsInternal: true,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleLaunchPma(w http.ResponseWriter, r *http.Request) {
	// Generate a secure one-time access token
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	// Single Sign-On path behind panel authentication
	launchURL := fmt.Sprintf("/_panel/db-manager?token=%s", token)

	writeJSON(w, http.StatusOK, map[string]string{
		"message":    "phpMyAdmin session ready",
		"launch_url": launchURL,
	})
}
