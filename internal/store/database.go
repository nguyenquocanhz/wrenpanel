package store

import (
	"database/sql"
	"fmt"
)

type Database struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id"`
	VhostID   *int64 `json:"vhost_id"`
	Engine    string `json:"engine"` // mysql | postgresql | mongodb
	DBName    string `json:"db_name"`
	DBUser    string `json:"db_user"`
	CreatedAt string `json:"created_at"`
}

func (s *Store) CreateDatabase(db *Database) error {
	res, err := s.db.Exec(`
		INSERT INTO databases (account_id, vhost_id, engine, db_name, db_user)
		VALUES (?, ?, ?, ?, ?)
	`, db.AccountID, db.VhostID, db.Engine, db.DBName, db.DBUser)
	if err != nil {
		return fmt.Errorf("failed to create database record: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		db.ID = id
	}
	return nil
}

func (s *Store) GetDatabaseByID(id int64) (*Database, error) {
	d := &Database{}
	err := s.db.QueryRow(`
		SELECT id, account_id, vhost_id, engine, db_name, db_user, created_at
		FROM databases WHERE id = ?
	`, id).Scan(&d.ID, &d.AccountID, &d.VhostID, &d.Engine, &d.DBName, &d.DBUser, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Store) ListDatabases(accountID *int64) ([]Database, error) {
	query := `SELECT id, account_id, vhost_id, engine, db_name, db_user, created_at FROM databases`
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

	var list []Database
	for rows.Next() {
		var d Database
		if err := rows.Scan(&d.ID, &d.AccountID, &d.VhostID, &d.Engine, &d.DBName, &d.DBUser, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (s *Store) DeleteDatabase(id int64) error {
	_, err := s.db.Exec(`DELETE FROM databases WHERE id = ?`, id)
	return err
}
