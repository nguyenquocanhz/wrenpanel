package store

import (
	"database/sql"
	"fmt"
	"strings"
)

type Vhost struct {
	ID             int64   `json:"id"`
	AccountID      int64   `json:"account_id"`
	ParentVhostID  *int64  `json:"parent_vhost_id"`
	Type           string  `json:"type"` // primary | subdomain | addon | alias | system
	FQDN           string  `json:"fqdn"`
	Docroot        string  `json:"docroot"`
	DocrootOwnerID *int64  `json:"docroot_owner_id"`
	WebEngine      string  `json:"web_engine"` // nginx | nginx+apache
	PhpVersionID   *int64  `json:"php_version_id"`
	SslCertID      *int64  `json:"ssl_cert_id"`
	Status         string  `json:"status"` // active | disabled
	CreatedAt      string  `json:"created_at"`
}

// ReservedPrefixes defines service prefixes that users cannot use for subdomains
var ReservedPrefixes = []string{
	"mail", "webmail", "wrenpanel", "autoconfig", "autodiscover", "ftp", "ns1", "ns2",
}

func IsReservedPrefix(subdomainPrefix string) bool {
	lower := strings.ToLower(strings.TrimSpace(subdomainPrefix))
	for _, p := range ReservedPrefixes {
		if lower == p {
			return true
		}
	}
	return false
}

func (s *Store) CreateVhost(v *Vhost) error {
	res, err := s.db.Exec(`
		INSERT INTO vhosts (
			account_id, parent_vhost_id, type, fqdn, docroot,
			docroot_owner_id, web_engine, php_version_id, ssl_cert_id, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, v.AccountID, v.ParentVhostID, v.Type, v.FQDN, v.Docroot,
		v.DocrootOwnerID, v.WebEngine, v.PhpVersionID, v.SslCertID, v.Status)
	if err != nil {
		return fmt.Errorf("failed to create vhost: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		v.ID = id
		// If docroot_owner_id was not set, this vhost created the folder, so it owns it
		if v.DocrootOwnerID == nil {
			v.DocrootOwnerID = &id
			_, _ = s.db.Exec(`UPDATE vhosts SET docroot_owner_id = ? WHERE id = ?`, id, id)
		}
	}
	return nil
}

func (s *Store) GetVhostByID(id int64) (*Vhost, error) {
	v := &Vhost{}
	err := s.db.QueryRow(`
		SELECT id, account_id, parent_vhost_id, type, fqdn, docroot,
		       docroot_owner_id, web_engine, php_version_id, ssl_cert_id, status, created_at
		FROM vhosts WHERE id = ?
	`, id).Scan(
		&v.ID, &v.AccountID, &v.ParentVhostID, &v.Type, &v.FQDN, &v.Docroot,
		&v.DocrootOwnerID, &v.WebEngine, &v.PhpVersionID, &v.SslCertID, &v.Status, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Store) GetVhostByFQDN(fqdn string) (*Vhost, error) {
	v := &Vhost{}
	err := s.db.QueryRow(`
		SELECT id, account_id, parent_vhost_id, type, fqdn, docroot,
		       docroot_owner_id, web_engine, php_version_id, ssl_cert_id, status, created_at
		FROM vhosts WHERE fqdn = ?
	`, fqdn).Scan(
		&v.ID, &v.AccountID, &v.ParentVhostID, &v.Type, &v.FQDN, &v.Docroot,
		&v.DocrootOwnerID, &v.WebEngine, &v.PhpVersionID, &v.SslCertID, &v.Status, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Store) ListVhosts(accountID *int64) ([]Vhost, error) {
	query := `
		SELECT id, account_id, parent_vhost_id, type, fqdn, docroot,
		       docroot_owner_id, web_engine, php_version_id, ssl_cert_id, status, created_at
		FROM vhosts
	`
	var args []interface{}
	if accountID != nil {
		query += " WHERE account_id = ?"
		args = append(args, *accountID)
	}
	query += " ORDER BY id ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Vhost
	for rows.Next() {
		var v Vhost
		if err := rows.Scan(
			&v.ID, &v.AccountID, &v.ParentVhostID, &v.Type, &v.FQDN, &v.Docroot,
			&v.DocrootOwnerID, &v.WebEngine, &v.PhpVersionID, &v.SslCertID, &v.Status, &v.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (s *Store) UpdateVhost(v *Vhost) error {
	_, err := s.db.Exec(`
		UPDATE vhosts
		SET web_engine = ?, php_version_id = ?, ssl_cert_id = ?, status = ?
		WHERE id = ?
	`, v.WebEngine, v.PhpVersionID, v.SslCertID, v.Status, v.ID)
	return err
}

func (s *Store) DeleteVhost(id int64) error {
	_, err := s.db.Exec(`DELETE FROM vhosts WHERE id = ?`, id)
	return err
}

// CountChildVhosts checks if this vhost has child subdomains/addons
func (s *Store) CountChildVhosts(vhostID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(1) FROM vhosts WHERE parent_vhost_id = ?
	`, vhostID).Scan(&count)
	return count, err
}

// CountOtherVhostsSharingDocroot checks if other active vhosts are sharing the same docroot
func (s *Store) CountOtherVhostsSharingDocroot(docroot string, excludeVhostID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(1) FROM vhosts
		WHERE docroot = ? AND id != ? AND status = 'active'
	`, docroot, excludeVhostID).Scan(&count)
	return count, err
}

// VhostDeletionAnalysis provides details for the safe deletion modal/logic
type VhostDeletionImpact struct {
	VhostID         int64  `json:"vhost_id"`
	FQDN            string `json:"fqdn"`
	Docroot         string `json:"docroot"`
	HasChildren     bool   `json:"has_children"`
	ChildCount      int    `json:"child_count"`
	SharedDocroot   bool   `json:"shared_docroot"`
	SharedCount     int    `json:"shared_count"`
	IsDocrootOwner  bool   `json:"is_docroot_owner"`
	CanDeleteFiles  bool   `json:"can_delete_files"`
	MustKeepFiles   bool   `json:"must_keep_files"`
	AllowDelete     bool   `json:"allow_delete"`
	ReasonIfBlocked string `json:"reason_if_blocked"`
}

// AnalyzeDeletionImpact calculates the exact impact before deletion based on docroot_owner rules
func (s *Store) AnalyzeDeletionImpact(vhostID int64) (*VhostDeletionImpact, error) {
	v, err := s.GetVhostByID(vhostID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, fmt.Errorf("vhost with ID %d not found", vhostID)
	}

	childCount, err := s.CountChildVhosts(vhostID)
	if err != nil {
		return nil, err
	}

	sharedCount, err := s.CountOtherVhostsSharingDocroot(v.Docroot, vhostID)
	if err != nil {
		return nil, err
	}

	impact := &VhostDeletionImpact{
		VhostID:        v.ID,
		FQDN:           v.FQDN,
		Docroot:        v.Docroot,
		HasChildren:    childCount > 0,
		ChildCount:     childCount,
		SharedDocroot:  sharedCount > 0,
		SharedCount:    sharedCount,
		IsDocrootOwner: v.DocrootOwnerID != nil && *v.DocrootOwnerID == v.ID,
	}

	// Rule 1: Cannot delete primary domain with existing subdomains/addons
	if impact.HasChildren {
		impact.AllowDelete = false
		impact.ReasonIfBlocked = fmt.Sprintf("Domain có %d subdomain/addon con. Bạn phải xóa hết domain con trước khi xóa domain chính.", childCount)
		return impact, nil
	}

	impact.AllowDelete = true

	// Rule 2: Docroot file deletion logic
	if impact.SharedDocroot {
		// Other sites share this folder -> Must strictly keep files
		impact.CanDeleteFiles = false
		impact.MustKeepFiles = true
	} else if impact.IsDocrootOwner {
		// No other sites, and is docroot owner -> can ask user whether to delete or keep
		impact.CanDeleteFiles = true
		impact.MustKeepFiles = false
	} else {
		// No other sites, but NOT docroot owner (e.g. shared folder from base domain) -> must keep files
		impact.CanDeleteFiles = false
		impact.MustKeepFiles = true
	}

	return impact, nil
}
