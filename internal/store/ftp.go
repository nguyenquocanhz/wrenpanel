package store

import (
	"database/sql"
	"fmt"
)

type FtpAccount struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id"`
	Username  string `json:"username"`
	RootDir   string `json:"root_dir"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (s *Store) CreateFtpAccount(ftp *FtpAccount) error {
	res, err := s.db.Exec(`
		INSERT INTO ftp_accounts (account_id, username, root_dir, status)
		VALUES (?, ?, ?, ?)
	`, ftp.AccountID, ftp.Username, ftp.RootDir, ftp.Status)
	if err != nil {
		return fmt.Errorf("failed to create ftp account: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		ftp.ID = id
	}
	return nil
}

func (s *Store) GetFtpAccountByID(id int64) (*FtpAccount, error) {
	f := &FtpAccount{}
	err := s.db.QueryRow(`
		SELECT id, account_id, username, root_dir, status, created_at
		FROM ftp_accounts WHERE id = ?
	`, id).Scan(&f.ID, &f.AccountID, &f.Username, &f.RootDir, &f.Status, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Store) ListFtpAccounts(accountID *int64) ([]FtpAccount, error) {
	query := `SELECT id, account_id, username, root_dir, status, created_at FROM ftp_accounts`
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

	var list []FtpAccount
	for rows.Next() {
		var f FtpAccount
		if err := rows.Scan(&f.ID, &f.AccountID, &f.Username, &f.RootDir, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, nil
}

func (s *Store) DeleteFtpAccount(id int64) error {
	_, err := s.db.Exec(`DELETE FROM ftp_accounts WHERE id = ?`, id)
	return err
}
