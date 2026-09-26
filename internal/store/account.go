package store

import (
	"database/sql"
	"fmt"
)

type Account struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username"`
	HomeDir      string  `json:"home_dir"`
	Email        *string `json:"email"`
	DiskQuotaMB  int64   `json:"disk_quota_mb"`
	Status       string  `json:"status"` // active | suspended
	PasswordHash *string `json:"-"`
	CreatedAt    string  `json:"created_at"`
}

func (s *Store) CreateAccount(acc *Account) error {
	res, err := s.db.Exec(`
		INSERT INTO accounts (username, home_dir, email, disk_quota_mb, status, password_hash)
		VALUES (?, ?, ?, ?, ?, ?)
	`, acc.Username, acc.HomeDir, acc.Email, acc.DiskQuotaMB, acc.Status, acc.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		acc.ID = id
	}
	return nil
}

func (s *Store) GetAccountByID(id int64) (*Account, error) {
	acc := &Account{}
	err := s.db.QueryRow(`
		SELECT id, username, home_dir, email, disk_quota_mb, status, password_hash, created_at
		FROM accounts WHERE id = ?
	`, id).Scan(&acc.ID, &acc.Username, &acc.HomeDir, &acc.Email, &acc.DiskQuotaMB, &acc.Status, &acc.PasswordHash, &acc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *Store) GetAccountByUsername(username string) (*Account, error) {
	acc := &Account{}
	err := s.db.QueryRow(`
		SELECT id, username, home_dir, email, disk_quota_mb, status, password_hash, created_at
		FROM accounts WHERE username = ?
	`, username).Scan(&acc.ID, &acc.Username, &acc.HomeDir, &acc.Email, &acc.DiskQuotaMB, &acc.Status, &acc.PasswordHash, &acc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *Store) ListAccounts() ([]Account, error) {
	rows, err := s.db.Query(`
		SELECT id, username, home_dir, email, disk_quota_mb, status, password_hash, created_at
		FROM accounts ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Username, &a.HomeDir, &a.Email, &a.DiskQuotaMB, &a.Status, &a.PasswordHash, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) UpdateAccount(acc *Account) error {
	_, err := s.db.Exec(`
		UPDATE accounts
		SET email = ?, disk_quota_mb = ?, status = ?
		WHERE id = ?
	`, acc.Email, acc.DiskQuotaMB, acc.Status, acc.ID)
	return err
}

func (s *Store) UpdateAccountPassword(id int64, passwordHash string) error {
	_, err := s.db.Exec(`UPDATE accounts SET password_hash = ? WHERE id = ?`, passwordHash, id)
	return err
}

func (s *Store) DeleteAccount(id int64) error {
	_, err := s.db.Exec(`DELETE FROM accounts WHERE id = ?`, id)
	return err
}
