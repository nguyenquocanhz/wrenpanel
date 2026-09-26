package store

import (
	"fmt"
)

type CronJob struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id"`
	Schedule  string `json:"schedule"` // cron expression
	Command   string `json:"command"`
	Status    string `json:"status"` // active | disabled
}

func (s *Store) CreateCronJob(c *CronJob) error {
	res, err := s.db.Exec(`
		INSERT INTO cron_jobs (account_id, schedule, command, status)
		VALUES (?, ?, ?, ?)
	`, c.AccountID, c.Schedule, c.Command, c.Status)
	if err != nil {
		return fmt.Errorf("failed to create cron job: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		c.ID = id
	}
	return nil
}

func (s *Store) ListCronJobs(accountID *int64) ([]CronJob, error) {
	query := `SELECT id, account_id, schedule, command, status FROM cron_jobs`
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

	var list []CronJob
	for rows.Next() {
		var c CronJob
		if err := rows.Scan(&c.ID, &c.AccountID, &c.Schedule, &c.Command, &c.Status); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (s *Store) DeleteCronJob(id int64) error {
	_, err := s.db.Exec(`DELETE FROM cron_jobs WHERE id = ?`, id)
	return err
}
