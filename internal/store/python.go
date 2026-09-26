package store

import (
	"database/sql"
	"fmt"
)

type PythonVersion struct {
	ID              int64  `json:"id"`
	Version         string `json:"version"`          // "3.11"
	InterpreterPath string `json:"interpreter_path"` // /opt/wrenpanel/python/3.11/bin/python3
	InstalledAt     string `json:"installed_at"`
}

func (s *Store) ListPythonVersions() ([]PythonVersion, error) {
	rows, err := s.db.Query(`
		SELECT id, version, interpreter_path, installed_at
		FROM python_versions ORDER BY version DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PythonVersion
	for rows.Next() {
		var p PythonVersion
		if err := rows.Scan(&p.ID, &p.Version, &p.InterpreterPath, &p.InstalledAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (s *Store) GetPythonVersionByID(id int64) (*PythonVersion, error) {
	p := &PythonVersion{}
	err := s.db.QueryRow(`
		SELECT id, version, interpreter_path, installed_at
		FROM python_versions WHERE id = ?
	`, id).Scan(&p.ID, &p.Version, &p.InterpreterPath, &p.InstalledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) CreatePythonVersion(p *PythonVersion) error {
	res, err := s.db.Exec(`
		INSERT INTO python_versions (version, interpreter_path)
		VALUES (?, ?)
	`, p.Version, p.InterpreterPath)
	if err != nil {
		return fmt.Errorf("failed to register python version: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		p.ID = id
	}
	return nil
}

// CountAppsUsingPythonVersion checks how many python apps reference this version
func (s *Store) CountAppsUsingPythonVersion(version string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(1) FROM apps WHERE kind = 'python' AND runtime_version = ?
	`, version).Scan(&count)
	return count, err
}

// DeletePythonVersion warns if apps still reference it
func (s *Store) DeletePythonVersion(id int64) error {
	v, err := s.GetPythonVersionByID(id)
	if err != nil {
		return err
	}
	if v == nil {
		return fmt.Errorf("python version not found")
	}

	count, err := s.CountAppsUsingPythonVersion(v.Version)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot remove Python version %s: %d app(s) are currently referencing this version", v.Version, count)
	}

	_, err = s.db.Exec(`DELETE FROM python_versions WHERE id = ?`, id)
	return err
}
