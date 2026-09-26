package store

import (
	"database/sql"
	"fmt"
)

type Backup struct {
	ID           int64   `json:"id"`
	AccountID    int64   `json:"account_id"`
	Scope        string  `json:"scope"` // site | database | full
	VhostID      *int64  `json:"vhost_id"`
	FilePath     string  `json:"file_path"`
	Destination  string  `json:"destination"` // local | s3 | ftp | rsync
	ManifestJSON *string `json:"manifest_json"`
	CreatedAt    string  `json:"created_at"`
}

func (s *Store) CreateBackup(b *Backup) error {
	res, err := s.db.Exec(`
		INSERT INTO backups (account_id, scope, vhost_id, file_path, destination, manifest_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`, b.AccountID, b.Scope, b.VhostID, b.FilePath, b.Destination, b.ManifestJSON)
	if err != nil {
		return fmt.Errorf("failed to create backup record: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		b.ID = id
	}
	return nil
}

func (s *Store) GetBackupByID(id int64) (*Backup, error) {
	b := &Backup{}
	err := s.db.QueryRow(`
		SELECT id, account_id, scope, vhost_id, file_path, destination, manifest_json, created_at
		FROM backups WHERE id = ?
	`, id).Scan(&b.ID, &b.AccountID, &b.Scope, &b.VhostID, &b.FilePath, &b.Destination, &b.ManifestJSON, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Store) ListBackups(accountID *int64) ([]Backup, error) {
	query := `SELECT id, account_id, scope, vhost_id, file_path, destination, manifest_json, created_at FROM backups`
	var args []interface{}
	if accountID != nil {
		query += " WHERE account_id = ?"
		args = append(args, *accountID)
	}
	query += " ORDER BY id DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Backup
	for rows.Next() {
		var b Backup
		if err := rows.Scan(&b.ID, &b.AccountID, &b.Scope, &b.VhostID, &b.FilePath, &b.Destination, &b.ManifestJSON, &b.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

func (s *Store) DeleteBackup(id int64) error {
	_, err := s.db.Exec(`DELETE FROM backups WHERE id = ?`, id)
	return err
}
