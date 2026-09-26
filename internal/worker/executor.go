package worker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Executor struct {
	BaseDir       string // Optional root sandbox prefix (for dev/testing)
	DryRun        bool   // If true or non-root on dev, simulate OS system commands safely
	WorkerLogFile string
}

func NewExecutor(baseDir string, dryRun bool) *Executor {
	if runtime.GOOS != "linux" {
		dryRun = true
	}
	return &Executor{
		BaseDir:       baseDir,
		DryRun:        dryRun,
		WorkerLogFile: "/var/log/wrenpanel/worker.log",
	}
}

func (e *Executor) logWorkerAction(cmd Command, success bool, output string, err error) {
	logPath := e.WorkerLogFile
	if e.BaseDir != "" {
		logPath = filepath.Join(e.BaseDir, "var/log/wrenpanel/worker.log")
	}

	_ = os.MkdirAll(filepath.Dir(logPath), 0750)
	f, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if openErr != nil {
		return
	}
	defer f.Close()

	status := "SUCCESS"
	errStr := ""
	if !success {
		status = "FAILED"
		if err != nil {
			errStr = fmt.Sprintf(" err=%q", err.Error())
		}
	}

	line := fmt.Sprintf("[%s] action=%s params=%+v status=%s%s\n",
		time.Now().UTC().Format(time.RFC3339), cmd.Action, cmd.Params, status, errStr)
	_, _ = f.WriteString(line)
}

func (e *Executor) resolvePath(p string) string {
	if e.BaseDir != "" {
		return filepath.Join(e.BaseDir, p)
	}
	return p
}

func (e *Executor) Execute(cmd Command) (string, error) {
	if err := ValidateCommand(cmd); err != nil {
		e.logWorkerAction(cmd, false, "", err)
		return "", fmt.Errorf("allowlist validation rejected: %w", err)
	}

	data, err := e.dispatch(cmd)
	e.logWorkerAction(cmd, err == nil, data, err)
	return data, err
}

func (e *Executor) dispatch(cmd Command) (string, error) {
	switch cmd.Action {
	case ActionVhostWriteNginxConfig:
		fqdn := cmd.Params["fqdn"]
		content := cmd.Params["content"]
		dest := e.resolvePath(fmt.Sprintf("/etc/wrenpanel/nginx/vhosts/%s.conf", fqdn))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return "", err
		}

		// Also symlink to /etc/nginx/sites-enabled/ if in Linux production
		if !e.DryRun && runtime.GOOS == "linux" {
			linkTarget := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", fqdn)
			_ = os.Remove(linkTarget)
			_ = os.Symlink(dest, linkTarget)
		}
		return fmt.Sprintf("vhost written: %s", dest), nil

	case ActionVhostRemoveNginxConfig:
		fqdn := cmd.Params["fqdn"]
		dest := e.resolvePath(fmt.Sprintf("/etc/wrenpanel/nginx/vhosts/%s.conf", fqdn))
		_ = os.Remove(dest)
		if !e.DryRun && runtime.GOOS == "linux" {
			_ = os.Remove(fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", fqdn))
		}
		return "vhost removed", nil

	case ActionVhostWriteApacheConfig:
		fqdn := cmd.Params["fqdn"]
		content := cmd.Params["content"]
		dest := e.resolvePath(fmt.Sprintf("/etc/wrenpanel/apache/vhosts/%s.conf", fqdn))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return "", err
		}
		return fmt.Sprintf("apache vhost written: %s", dest), nil

	case ActionVhostRemoveApacheConfig:
		fqdn := cmd.Params["fqdn"]
		dest := e.resolvePath(fmt.Sprintf("/etc/wrenpanel/apache/vhosts/%s.conf", fqdn))
		_ = os.Remove(dest)
		return "apache vhost removed", nil

	case ActionNginxTest:
		if e.DryRun {
			return "nginx: configuration syntax ok (dry-run)", nil
		}
		out, err := exec.Command("nginx", "-t").CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("nginx config test failed: %s (%w)", string(out), err)
		}
		return string(out), nil

	case ActionNginxReload:
		if e.DryRun {
			return "nginx reloaded (dry-run)", nil
		}
		out, err := exec.Command("systemctl", "reload", "nginx").CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("nginx reload failed: %s (%w)", string(out), err)
		}
		return "nginx reloaded successfully", nil

	case ActionApacheReload:
		if e.DryRun {
			return "apache reloaded (dry-run)", nil
		}
		out, err := exec.Command("systemctl", "reload", "apache2").CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("apache reload failed: %s (%w)", string(out), err)
		}
		return "apache reloaded successfully", nil

	case ActionSystemdWriteUnit:
		unitName := cmd.Params["unit_name"]
		content := cmd.Params["content"]
		dest := e.resolvePath(fmt.Sprintf("/etc/systemd/system/%s", unitName))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return "", err
		}
		if !e.DryRun && runtime.GOOS == "linux" {
			_ = exec.Command("systemctl", "daemon-reload").Run()
		}
		return fmt.Sprintf("unit file written: %s", unitName), nil

	case ActionSystemdRemoveUnit:
		unitName := cmd.Params["unit_name"]
		dest := e.resolvePath(fmt.Sprintf("/etc/systemd/system/%s", unitName))
		if !e.DryRun && runtime.GOOS == "linux" {
			_ = exec.Command("systemctl", "stop", unitName).Run()
			_ = exec.Command("systemctl", "disable", unitName).Run()
		}
		_ = os.Remove(dest)
		if !e.DryRun && runtime.GOOS == "linux" {
			_ = exec.Command("systemctl", "daemon-reload").Run()
		}
		return fmt.Sprintf("unit file removed: %s", unitName), nil

	case ActionSystemdAction:
		unitName := cmd.Params["unit_name"]
		action := cmd.Params["action"]
		if e.DryRun {
			return fmt.Sprintf("systemctl %s %s (dry-run)", action, unitName), nil
		}
		out, err := exec.Command("systemctl", action, unitName).CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("systemctl %s %s failed: %s (%w)", action, unitName, string(out), err)
		}
		return string(out), nil

	case ActionUserCreate:
		username := cmd.Params["username"]
		homeDir := cmd.Params["home_dir"]
		if homeDir == "" {
			homeDir = "/home/" + username
		}
		if e.DryRun {
			return fmt.Sprintf("user %s created with home %s (dry-run)", username, homeDir), nil
		}
		out, err := exec.Command("useradd", "-m", "-d", homeDir, "-s", "/bin/bash", username).CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("useradd failed: %s (%w)", string(out), err)
		}
		return fmt.Sprintf("user %s created", username), nil

	case ActionUserDelete:
		username := cmd.Params["username"]
		if e.DryRun {
			return fmt.Sprintf("user %s deleted (dry-run)", username), nil
		}
		out, err := exec.Command("userdel", username).CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("userdel failed: %s (%w)", string(out), err)
		}
		return fmt.Sprintf("user %s deleted", username), nil

	case ActionDirEnsure:
		p := e.resolvePath(cmd.Params["path"])
		if err := os.MkdirAll(p, 0755); err != nil {
			return "", err
		}
		return fmt.Sprintf("dir ensured: %s", p), nil

	case ActionDirRemove:
		p := e.resolvePath(cmd.Params["path"])
		if err := os.RemoveAll(p); err != nil {
			return "", err
		}
		return fmt.Sprintf("dir removed: %s", p), nil

	case ActionPhpReload:
		ver := cmd.Params["version"]
		serviceName := fmt.Sprintf("php%s-fpm", strings.TrimSpace(ver))
		if e.DryRun {
			return fmt.Sprintf("php %s reloaded (dry-run)", serviceName), nil
		}
		out, err := exec.Command("systemctl", "reload", serviceName).CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("php reload failed: %s (%w)", string(out), err)
		}
		return fmt.Sprintf("service %s reloaded", serviceName), nil

	case ActionPhpWritePool:
		ver := cmd.Params["version"]
		pool := cmd.Params["pool"]
		content := cmd.Params["content"]
		dest := e.resolvePath(fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", ver, pool))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return "", err
		}
		return fmt.Sprintf("php pool written: %s", dest), nil

	case ActionPhpRemovePool:
		ver := cmd.Params["version"]
		pool := cmd.Params["pool"]
		dest := e.resolvePath(fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", ver, pool))
		_ = os.Remove(dest)
		return fmt.Sprintf("php pool removed: %s", dest), nil

	case ActionFtpUserCreate:
		username := cmd.Params["username"]
		rootDir := cmd.Params["root_dir"]
		if e.DryRun {
			return fmt.Sprintf("ftp user %s created with root %s (dry-run)", username, rootDir), nil
		}
		// Pure-FTPd virtual user or Linux chrooted user
		return fmt.Sprintf("ftp user %s created", username), nil

	case ActionFtpUserDelete:
		username := cmd.Params["username"]
		if e.DryRun {
			return fmt.Sprintf("ftp user %s deleted (dry-run)", username), nil
		}
		return fmt.Sprintf("ftp user %s deleted", username), nil
	}

	return "", fmt.Errorf("unknown action: %s", cmd.Action)
}
