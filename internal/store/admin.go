package store

import (
	"database/sql"
	"fmt"
)

type Admin struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"-"`
	Role         string  `json:"role"`
	TotpSecret   *string `json:"-"`
	TotpEnabled  int     `json:"totp_enabled"`
	Status       string  `json:"status"`
	LastLoginAt  *string `json:"last_login_at"`
	CreatedAt    string  `json:"created_at"`
}

func (s *Store) CreateAdmin(admin *Admin) error {
	res, err := s.db.Exec(`
		INSERT INTO admins (username, password_hash, role, totp_secret, totp_enabled, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, admin.Username, admin.PasswordHash, admin.Role, admin.TotpSecret, admin.TotpEnabled, admin.Status)
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		admin.ID = id
	}
	return nil
}

func (s *Store) GetAdminByUsername(username string) (*Admin, error) {
	a := &Admin{}
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, role, totp_secret, totp_enabled, status, last_login_at, created_at
		FROM admins WHERE username = ?
	`, username).Scan(&a.ID, &a.Username, &a.PasswordHash, &a.Role, &a.TotpSecret, &a.TotpEnabled, &a.Status, &a.LastLoginAt, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) CountAdmins() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM admins`).Scan(&count)
	return count, err
}
