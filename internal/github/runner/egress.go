// Package runner provides network egress validation and audit logging.
package runner

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// EgressValidator validates network access.
// REQ-SEC-007, REQ-CI-007.1, REQ-CI-007.2
type EgressValidator interface {
	ValidateGitHubAPI(ctx context.Context) error
	ValidateRunnerDownload(ctx context.Context) error
}

// egressValidator implements EgressValidator.
type egressValidator struct {
	dialer *net.Dialer
}

// NewEgressValidator creates a new EgressValidator.
func NewEgressValidator() EgressValidator {
	return &egressValidator{
		dialer: &net.Dialer{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateGitHubAPI validates GitHub API access.
func (v *egressValidator) ValidateGitHubAPI(ctx context.Context) error {
	// GitHub API domain list
	domains := []string{
		"api.github.com:443",
		"github.com:443",
	}

	for _, domain := range domains {
		conn, err := v.dialer.DialContext(ctx, "tcp", domain)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", domain, err)
		}
		_ = conn.Close()
	}

	return nil
}

// ValidateRunnerDownload validates runner download access.
func (v *egressValidator) ValidateRunnerDownload(ctx context.Context) error {
	// GitHub runner download domains
	domains := []string{
		"github.com:443",
		"objects.githubusercontent.com:443",
		"githubusercontent.com:443",
	}

	for _, domain := range domains {
		conn, err := v.dialer.DialContext(ctx, "tcp", domain)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", domain, err)
		}
		_ = conn.Close()
	}

	return nil
}

// AuditLogger logs runner operations.
// REQ-SEC-008
type AuditLogger interface {
	LogInstall(ctx context.Context, version string) error
	LogRegister(ctx context.Context, runnerID string) error
}

// auditLogger implements AuditLogger.
type auditLogger struct {
	auditDir string
}

// NewAuditLogger creates a new AuditLogger.
func NewAuditLogger(auditDir string) AuditLogger {
	if auditDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			auditDir = "/tmp/moai-audit"
		} else {
			auditDir = filepath.Join(homeDir, ".moai", "audit")
		}
	}
	return &auditLogger{
		auditDir: auditDir,
	}
}

// LogInstall logs installation operations.
func (l *auditLogger) LogInstall(ctx context.Context, version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// Create audit directory
	if err := os.MkdirAll(l.auditDir, 0755); err != nil {
		return fmt.Errorf("create audit directory: %w", err)
	}

	// Create audit log file
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	logPath := filepath.Join(l.auditDir, fmt.Sprintf("install_%s.log", timestamp))

	entry := fmt.Sprintf("[%s] INSTALL version=%s\n", timestamp, version)

	if err := os.WriteFile(logPath, []byte(entry), 0644); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}

// LogRegister logs registration operations.
func (l *auditLogger) LogRegister(ctx context.Context, runnerID string) error {
	if runnerID == "" {
		return fmt.Errorf("runner ID cannot be empty")
	}

	// Create audit directory
	if err := os.MkdirAll(l.auditDir, 0755); err != nil {
		return fmt.Errorf("create audit directory: %w", err)
	}

	// Create audit log file
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	logPath := filepath.Join(l.auditDir, fmt.Sprintf("register_%s.log", timestamp))

	entry := fmt.Sprintf("[%s] REGISTER runner_id=%s\n", timestamp, runnerID)

	if err := os.WriteFile(logPath, []byte(entry), 0644); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}
