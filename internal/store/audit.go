package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type AuditLog struct {
	ID          int64   `json:"id"`
	Actor       string  `json:"actor"`  // admin username
	Action      string  `json:"action"` // vhost.create, php.remove_version...
	Target      *string `json:"target"`
	PayloadJSON *string `json:"payload_json"`
	Result      string  `json:"result"` // success | failed
	CreatedAt   string  `json:"created_at"`
}

var auditLogFilePath = "/var/log/wrenpanel/audit.log"

func SetAuditLogFilePath(path string) {
	auditLogFilePath = path
}

func (s *Store) RecordAudit(actor, action string, target, payloadJSON *string, result string) error {
	// 1. Insert into database
	res, err := s.db.Exec(`
		INSERT INTO audit_log (actor, action, target, payload_json, result)
		VALUES (?, ?, ?, ?, ?)
	`, actor, action, target, payloadJSON, result)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	// 2. Append-only to audit log file (as required by GEMINI.md)
	go func() {
		dir := filepath.Dir(auditLogFilePath)
		if dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0750)
		}
		f, err := os.OpenFile(auditLogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err == nil {
			defer f.Close()
			targetStr := ""
			if target != nil {
				targetStr = *target
			}
			payloadStr := ""
			if payloadJSON != nil {
				payloadStr = *payloadJSON
			}
			line := fmt.Sprintf("[%s] actor=%s action=%s target=%s result=%s payload=%s\n",
				time.Now().UTC().Format(time.RFC3339), actor, action, targetStr, result, payloadStr)
			_, _ = f.WriteString(line)
		}
	}()

	_ = res
	return nil
}

func (s *Store) ListAuditLogs(limit int) ([]AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`
		SELECT id, actor, action, target, payload_json, result, created_at
		FROM audit_log ORDER BY id DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AuditLog
	for rows.Next() {
		var a AuditLog
		if err := rows.Scan(&a.ID, &a.Actor, &a.Action, &a.Target, &a.PayloadJSON, &a.Result, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}
