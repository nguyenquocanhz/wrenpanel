package store

import (
	"database/sql"
	"fmt"
)

type NodeVersion struct {
	ID          int64  `json:"id"`
	Version     string `json:"version"`      // "20.11.0"
	InstallPath string `json:"install_path"` // /opt/wrenpanel/node/20.11.0
	InstalledAt string `json:"installed_at"`
}

func (s *Store) ListNodeVersions() ([]NodeVersion, error) {
	rows, err := s.db.Query(`
		SELECT id, version, install_path, installed_at
		FROM node_versions ORDER BY version DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []NodeVersion
	for rows.Next() {
		var n NodeVersion
		if err := rows.Scan(&n.ID, &n.Version, &n.InstallPath, &n.InstalledAt); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

func (s *Store) GetNodeVersionByID(id int64) (*NodeVersion, error) {
	n := &NodeVersion{}
	err := s.db.QueryRow(`
		SELECT id, version, install_path, installed_at
		FROM node_versions WHERE id = ?
	`, id).Scan(&n.ID, &n.Version, &n.InstallPath, &n.InstalledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Store) CreateNodeVersion(n *NodeVersion) error {
	res, err := s.db.Exec(`
		INSERT INTO node_versions (version, install_path)
		VALUES (?, ?)
	`, n.Version, n.InstallPath)
	if err != nil {
		return fmt.Errorf("failed to register node version: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		n.ID = id
	}
	return nil
}

// CountAppsUsingNodeVersion checks how many node apps reference this version
func (s *Store) CountAppsUsingNodeVersion(version string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(1) FROM apps WHERE kind = 'node' AND runtime_version = ?
	`, version).Scan(&count)
	return count, err
}

// DeleteNodeVersion checks reference before deletion
func (s *Store) DeleteNodeVersion(id int64) error {
	v, err := s.GetNodeVersionByID(id)
	if err != nil {
		return err
	}
	if v == nil {
		return fmt.Errorf("node version not found")
	}

	count, err := s.CountAppsUsingNodeVersion(v.Version)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot remove Node.js version %s: %d app(s) are currently referencing this version", v.Version, count)
	}

	_, err = s.db.Exec(`DELETE FROM node_versions WHERE id = ?`, id)
	return err
}
