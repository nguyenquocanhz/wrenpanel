package store

import (
	"fmt"
	"net"
	"time"
)

type PortAllocation struct {
	Port       int       `json:"port"`
	AppID      int64     `json:"app_id"`
	ReservedAt time.Time `json:"reserved_at"`
}

// IsPortAvailable checks if a port can be bound on 127.0.0.1
func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

// AllocatePort finds an available port between minPort and maxPort and reserves it
func (s *Store) AllocatePort(appID int64, minPort, maxPort int) (int, error) {
	if minPort <= 0 {
		minPort = 10000
	}
	if maxPort <= 0 {
		maxPort = 20000
	}

	for port := minPort; port <= maxPort; port++ {
		// 1. Check if already reserved in database
		var exists int
		err := s.db.QueryRow(`SELECT COUNT(1) FROM port_allocations WHERE port = ?`, port).Scan(&exists)
		if err != nil {
			return 0, err
		}
		if exists > 0 {
			continue
		}

		// 2. Try actual socket bind to verify it's really free
		if !IsPortAvailable(port) {
			continue
		}

		// 3. Reserve it in DB
		_, err = s.db.Exec(`
			INSERT INTO port_allocations (port, app_id)
			VALUES (?, ?)
		`, port, appID)
		if err == nil {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range [%d, %d]", minPort, maxPort)
}

func (s *Store) ReleasePort(port int) error {
	_, err := s.db.Exec(`DELETE FROM port_allocations WHERE port = ?`, port)
	return err
}

func (s *Store) ReleasePortsForApp(appID int64) error {
	_, err := s.db.Exec(`DELETE FROM port_allocations WHERE app_id = ?`, appID)
	return err
}
