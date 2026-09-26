package worker

import (
	"fmt"
	"regexp"
	"strings"
)

// AllowedActions list strictly allowed commands for the root worker
const (
	ActionVhostWriteNginxConfig  = "vhost.write_nginx_config"
	ActionVhostRemoveNginxConfig = "vhost.remove_nginx_config"
	ActionVhostWriteApacheConfig = "vhost.write_apache_config"
	ActionVhostRemoveApacheConfig = "vhost.remove_apache_config"
	ActionNginxTest              = "nginx.test"
	ActionNginxReload            = "nginx.reload"
	ActionApacheReload           = "apache.reload"
	ActionSystemdWriteUnit       = "systemd.write_unit"
	ActionSystemdRemoveUnit      = "systemd.remove_unit"
	ActionSystemdAction          = "systemd.action"
	ActionUserCreate             = "user.create"
	ActionUserDelete             = "user.delete"
	ActionUserSetPassword        = "user.set_password"
	ActionDirEnsure              = "dir.ensure"
	ActionDirRemove              = "dir.remove"
	ActionPhpReload              = "php.reload"
	ActionPhpWritePool           = "php.write_pool"
	ActionPhpRemovePool          = "php.remove_pool"
	ActionFtpUserCreate          = "ftp.user_create"
	ActionFtpUserDelete          = "ftp.user_delete"
)

var (
	validFQDNRegex     = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	validUsernameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{2,31}$`)
	validUnitNameRegex = regexp.MustCompile(`^wrenpanel-app-[a-z0-9-]+-[a-f0-9]{6}\.service$`)
)

type Command struct {
	ID     string            `json:"id"`
	Action string            `json:"action"`
	Params map[string]string `json:"params"`
}

type Response struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    string `json:"data,omitempty"`
}

// ValidateCommand verifies that the command is in the allowlist and has safe parameters
func ValidateCommand(cmd Command) error {
	switch cmd.Action {
	case ActionVhostWriteNginxConfig, ActionVhostRemoveNginxConfig, ActionVhostWriteApacheConfig, ActionVhostRemoveApacheConfig:
		fqdn := cmd.Params["fqdn"]
		if !validFQDNRegex.MatchString(fqdn) {
			return fmt.Errorf("invalid FQDN format: %s", fqdn)
		}
		if cmd.Action == ActionVhostWriteNginxConfig || cmd.Action == ActionVhostWriteApacheConfig {
			if strings.TrimSpace(cmd.Params["content"]) == "" {
				return fmt.Errorf("config content cannot be empty")
			}
		}

	case ActionNginxTest, ActionNginxReload, ActionApacheReload:
		// No special params required

	case ActionSystemdWriteUnit:
		unitName := cmd.Params["unit_name"]
		if !validUnitNameRegex.MatchString(unitName) {
			return fmt.Errorf("invalid systemd unit name '%s': must match pattern wrenpanel-app-<name>-<hash>.service", unitName)
		}
		if strings.TrimSpace(cmd.Params["content"]) == "" {
			return fmt.Errorf("unit content cannot be empty")
		}

	case ActionSystemdRemoveUnit:
		unitName := cmd.Params["unit_name"]
		if !validUnitNameRegex.MatchString(unitName) {
			return fmt.Errorf("invalid systemd unit name '%s'", unitName)
		}

	case ActionSystemdAction:
		unitName := cmd.Params["unit_name"]
		if !validUnitNameRegex.MatchString(unitName) {
			return fmt.Errorf("invalid systemd unit name '%s'", unitName)
		}
		act := cmd.Params["action"]
		if act != "start" && act != "stop" && act != "restart" && act != "status" && act != "reload" {
			return fmt.Errorf("invalid systemd action: %s", act)
		}

	case ActionUserCreate, ActionUserDelete:
		user := cmd.Params["username"]
		if !validUsernameRegex.MatchString(user) {
			return fmt.Errorf("invalid username '%s'", user)
		}
		if pass, ok := cmd.Params["password"]; ok && pass != "" {
			if strings.ContainsAny(pass, "\r\n") {
				return fmt.Errorf("password cannot contain newline characters")
			}
		}

	case ActionUserSetPassword:
		user := cmd.Params["username"]
		if !validUsernameRegex.MatchString(user) {
			return fmt.Errorf("invalid username '%s'", user)
		}
		pass := cmd.Params["password"]
		if pass == "" || strings.ContainsAny(pass, "\r\n") {
			return fmt.Errorf("valid password required without newlines")
		}

	case ActionDirEnsure, ActionDirRemove:
		path := cmd.Params["path"]
		// Path must be inside /home/ or /etc/wrenpanel/ or /var/log/
		if !strings.HasPrefix(path, "/home/") && !strings.HasPrefix(path, "/etc/wrenpanel/") && !strings.HasPrefix(path, "/var/log/") {
			return fmt.Errorf("directory path '%s' is not in an allowed directory hierarchy", path)
		}

	case ActionPhpReload, ActionPhpWritePool, ActionPhpRemovePool:
		version := cmd.Params["version"]
		if version == "" {
			return fmt.Errorf("php version is required")
		}

	case ActionFtpUserCreate, ActionFtpUserDelete:
		user := cmd.Params["username"]
		if !validUsernameRegex.MatchString(user) {
			return fmt.Errorf("invalid ftp username '%s'", user)
		}

	default:
		return fmt.Errorf("action '%s' is NOT permitted in root worker allowlist", cmd.Action)
	}

	return nil
}
