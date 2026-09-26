package store

import (
	"database/sql"
	"fmt"
)

type PhpVersion struct {
	ID          int64  `json:"id"`
	Version     string `json:"version"`     // "8.2"
	FpmService  string `json:"fpm_service"`  // "php8.2-fpm"
	BinaryPath  string `json:"binary_path"`
	InstalledAt string `json:"installed_at"`
}

func (s *Store) ListPhpVersions() ([]PhpVersion, error) {
	rows, err := s.db.Query(`
		SELECT id, version, fpm_service, binary_path, installed_at
		FROM php_versions ORDER BY version DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PhpVersion
	for rows.Next() {
		var p PhpVersion
		if err := rows.Scan(&p.ID, &p.Version, &p.FpmService, &p.BinaryPath, &p.InstalledAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (s *Store) GetPhpVersionByID(id int64) (*PhpVersion, error) {
	p := &PhpVersion{}
	err := s.db.QueryRow(`
		SELECT id, version, fpm_service, binary_path, installed_at
		FROM php_versions WHERE id = ?
	`, id).Scan(&p.ID, &p.Version, &p.FpmService, &p.BinaryPath, &p.InstalledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) CreatePhpVersion(p *PhpVersion) error {
	res, err := s.db.Exec(`
		INSERT INTO php_versions (version, fpm_service, binary_path)
		VALUES (?, ?, ?)
	`, p.Version, p.FpmService, p.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to register php version: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		p.ID = id
	}
	return nil
}

// CountVhostsUsingPhpVersion counts how many vhosts reference this PHP version
func (s *Store) CountVhostsUsingPhpVersion(id int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM vhosts WHERE php_version_id = ?`, id).Scan(&count)
	return count, err
}

// DeletePhpVersion ensures rule: Gỡ version: chặn nếu còn site đang gán version đó, chỉ cho gỡ khi 0 tham chiếu
func (s *Store) DeletePhpVersion(id int64) error {
	count, err := s.CountVhostsUsingPhpVersion(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot remove PHP version: %d vhost(s) are currently referencing this version", count)
	}

	_, err = s.db.Exec(`DELETE FROM php_versions WHERE id = ?`, id)
	return err
}
