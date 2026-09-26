package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"wrenpanel/internal/acme"
	"wrenpanel/internal/store"
	"wrenpanel/internal/worker"
)

type Server struct {
	store        *store.Store
	workerClient *worker.Client
	acmeManager  *acme.Manager
	router       *chi.Mux
}

func NewServer(s *store.Store, wc *worker.Client, acmeMgr *acme.Manager) *Server {
	srv := &Server{
		store:        s,
		workerClient: wc,
		acmeManager:  acmeMgr,
		router:       chi.NewRouter(),
	}

	srv.setupMiddleware()
	srv.setupRoutes()

	return srv
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Permissive CORS for development, local API access
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	s.router.Route("/api", func(r chi.Router) {
		// Auth
		r.Post("/auth/login", s.handleLogin)
		r.Get("/auth/me", s.handleMe)

		// System
		r.Get("/system/status", s.handleSystemStatus)

		// Accounts
		r.Get("/accounts", s.handleListAccounts)
		r.Post("/accounts", s.handleCreateAccount)
		r.Post("/accounts/{id}/reset-password", s.handleResetAccountPassword)
		r.Delete("/accounts/{id}", s.handleDeleteAccount)

		// Vhosts & Domains
		r.Get("/vhosts", s.handleListVhosts)
		r.Post("/vhosts", s.handleCreateVhost)
		r.Get("/vhosts/{id}", s.handleGetVhost)
		r.Get("/vhosts/{id}/deletion-impact", s.handleVhostDeletionImpact)
		r.Delete("/vhosts/{id}", s.handleDeleteVhost)

		// PHP Manager
		r.Get("/php/versions", s.handleListPhpVersions)
		r.Post("/php/versions", s.handleCreatePhpVersion)
		r.Delete("/php/versions/{id}", s.handleDeletePhpVersion)
		r.Post("/php/reload", s.handleReloadPhp)

		// Node & Python
		r.Get("/node/versions", s.handleListNodeVersions)
		r.Post("/node/versions", s.handleCreateNodeVersion)
		r.Delete("/node/versions/{id}", s.handleDeleteNodeVersion)

		r.Get("/python/versions", s.handleListPythonVersions)
		r.Post("/python/versions", s.handleCreatePythonVersion)
		r.Delete("/python/versions/{id}", s.handleDeletePythonVersion)

		// Apps (Node/Python/Docker)
		r.Get("/apps", s.handleListApps)
		r.Post("/apps", s.handleCreateApp)
		r.Get("/apps/{id}", s.handleGetApp)
		r.Post("/apps/{id}/action", s.handleAppAction)
		r.Delete("/apps/{id}", s.handleDeleteApp)

		// Databases & phpMyAdmin
		r.Get("/databases", s.handleListDatabases)
		r.Post("/databases", s.handleCreateDatabase)
		r.Delete("/databases/{id}", s.handleDeleteDatabase)
		r.Get("/db-manager/status", s.handleGetPmaStatus)
		r.Post("/db-manager/launch", s.handleLaunchPma)

		// FTP Accounts
		r.Get("/ftp", s.handleListFtp)
		r.Post("/ftp", s.handleCreateFtp)
		r.Delete("/ftp/{id}", s.handleDeleteFtp)

		// SSL
		r.Get("/ssl/certs", s.handleListCerts)
		r.Post("/ssl/issue", s.handleIssueCert)

		// Backups
		r.Get("/backups", s.handleListBackups)
		r.Post("/backups", s.handleCreateBackup)

		// Audit Log
		r.Get("/audit-logs", s.handleListAuditLogs)
	})
}

// Helpers
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
