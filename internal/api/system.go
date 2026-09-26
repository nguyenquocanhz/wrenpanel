package api

import (
	"net/http"
	"runtime"
)

type SystemStatusResponse struct {
	Version      string `json:"version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	GoVersion    string `json:"go_version"`
	WorkerStatus string `json:"worker_status"`
	NumCPU       int    `json:"num_cpu"`
}

func (s *Server) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	workerStatus := "connected"
	if s.workerClient == nil {
		workerStatus = "standalone/mock"
	}

	resp := SystemStatusResponse{
		Version:      "0.1.0",
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		WorkerStatus: workerStatus,
		NumCPU:       runtime.NumCPU(),
	}

	writeJSON(w, http.StatusOK, resp)
}
