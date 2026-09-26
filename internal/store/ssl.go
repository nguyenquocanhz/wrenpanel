package store

import (
	"database/sql"
	"fmt"
)

type Cert struct {
	ID        int64   `json:"id"`
	FQDN      string  `json:"fqdn"`
	CertPath  string  `json:"cert_path"`
	KeyPath   string  `json:"key_path"`
	Issuer    string  `json:"issuer"` // lets-encrypt | uploaded
	IssuedAt  *string `json:"issued_at"`
	ExpiresAt *string `json:"expires_at"`
	AutoRenew int     `json:"auto_renew"` // 1 | 0
}

func (s *Store) CreateCert(c *Cert) error {
	res, err := s.db.Exec(`
		INSERT INTO certs (fqdn, cert_path, key_path, issuer, issued_at, expires_at, auto_renew)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, c.FQDN, c.CertPath, c.KeyPath, c.Issuer, c.IssuedAt, c.ExpiresAt, c.AutoRenew)
	if err != nil {
		return fmt.Errorf("failed to save cert: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		c.ID = id
	}
	return nil
}

func (s *Store) GetCertByID(id int64) (*Cert, error) {
	c := &Cert{}
	err := s.db.QueryRow(`
		SELECT id, fqdn, cert_path, key_path, issuer, issued_at, expires_at, auto_renew
		FROM certs WHERE id = ?
	`, id).Scan(&c.ID, &c.FQDN, &c.CertPath, &c.KeyPath, &c.Issuer, &c.IssuedAt, &c.ExpiresAt, &c.AutoRenew)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) GetCertByFQDN(fqdn string) (*Cert, error) {
	c := &Cert{}
	err := s.db.QueryRow(`
		SELECT id, fqdn, cert_path, key_path, issuer, issued_at, expires_at, auto_renew
		FROM certs WHERE fqdn = ?
	`, fqdn).Scan(&c.ID, &c.FQDN, &c.CertPath, &c.KeyPath, &c.Issuer, &c.IssuedAt, &c.ExpiresAt, &c.AutoRenew)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) ListCerts() ([]Cert, error) {
	rows, err := s.db.Query(`
		SELECT id, fqdn, cert_path, key_path, issuer, issued_at, expires_at, auto_renew
		FROM certs ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Cert
	for rows.Next() {
		var c Cert
		if err := rows.Scan(&c.ID, &c.FQDN, &c.CertPath, &c.KeyPath, &c.Issuer, &c.IssuedAt, &c.ExpiresAt, &c.AutoRenew); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (s *Store) UpdateCert(c *Cert) error {
	_, err := s.db.Exec(`
		UPDATE certs
		SET cert_path = ?, key_path = ?, issuer = ?, issued_at = ?, expires_at = ?, auto_renew = ?
		WHERE id = ?
	`, c.CertPath, c.KeyPath, c.Issuer, c.IssuedAt, c.ExpiresAt, c.AutoRenew, c.ID)
	return err
}

func (s *Store) DeleteCert(id int64) error {
	_, err := s.db.Exec(`DELETE FROM certs WHERE id = ?`, id)
	return err
}
