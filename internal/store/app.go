package store

import (
	"database/sql"
	"fmt"
)

type App struct {
	ID             int64   `json:"id"`
	AccountID      int64   `json:"account_id"`
	VhostID        *int64  `json:"vhost_id"`
	Kind           string  `json:"kind"` // node | python | docker
	Name           string  `json:"name"`
	SystemdUnit    *string `json:"systemd_unit"`
	RuntimeVersion *string `json:"runtime_version"`
	InternalPort   *int    `json:"internal_port"`
	Entrypoint     *string `json:"entrypoint"`
	Status         string  `json:"status"` // stopped | running | failed
	CreatedAt      string  `json:"created_at"`
}

func (s *Store) CreateApp(app *App) error {
	res, err := s.db.Exec(`
		INSERT INTO apps (account_id, vhost_id, kind, name, systemd_unit, runtime_version, internal_port, entrypoint, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, app.AccountID, app.VhostID, app.Kind, app.Name, app.SystemdUnit, app.RuntimeVersion, app.InternalPort, app.Entrypoint, app.Status)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		app.ID = id
	}
	return nil
}

func (s *Store) GetAppByID(id int64) (*App, error) {
	app := &App{}
	err := s.db.QueryRow(`
		SELECT id, account_id, vhost_id, kind, name, systemd_unit, runtime_version, internal_port, entrypoint, status, created_at
		FROM apps WHERE id = ?
	`, id).Scan(&app.ID, &app.AccountID, &app.VhostID, &app.Kind, &app.Name, &app.SystemdUnit, &app.RuntimeVersion, &app.InternalPort, &app.Entrypoint, &app.Status, &app.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Store) ListApps(accountID *int64, kind *string) ([]App, error) {
	query := `SELECT id, account_id, vhost_id, kind, name, systemd_unit, runtime_version, internal_port, entrypoint, status, created_at FROM apps WHERE 1=1`
	var args []interface{}
	if accountID != nil {
		query += " AND account_id = ?"
		args = append(args, *accountID)
	}
	if kind != nil && *kind != "" {
		query += " AND kind = ?"
		args = append(args, *kind)
	}
	query += " ORDER BY id DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.AccountID, &a.VhostID, &a.Kind, &a.Name, &a.SystemdUnit, &a.RuntimeVersion, &a.InternalPort, &a.Entrypoint, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) UpdateAppStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE apps SET status = ? WHERE id = ?`, status, id)
	return err
}

func (s *Store) UpdateApp(app *App) error {
	_, err := s.db.Exec(`
		UPDATE apps
		SET vhost_id = ?, name = ?, runtime_version = ?, internal_port = ?, entrypoint = ?, status = ?
		WHERE id = ?
	`, app.VhostID, app.Name, app.RuntimeVersion, app.InternalPort, app.Entrypoint, app.Status, app.ID)
	return err
}

func (s *Store) DeleteApp(id int64) error {
	_ = s.ReleasePortsForApp(id)
	_, err := s.db.Exec(`DELETE FROM apps WHERE id = ?`, id)
	return err
}
