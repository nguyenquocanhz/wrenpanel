package api

import (
	"net/http"
	"strconv"

	"wrenpanel/internal/store"
)

func (s *Server) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	list, err := s.store.ListAuditLogs(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.AuditLog{}
	}
	writeJSON(w, http.StatusOK, list)
}
