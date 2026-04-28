// Package runner는 네트워크 egress 검증 및 감사 로깅을 제공합니다.
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

// EgressValidator는 네트워크 액세스를 검증합니다.
// REQ-SEC-007, REQ-CI-007.1, REQ-CI-007.2
type EgressValidator interface {
	ValidateGitHubAPI(ctx context.Context) error
	ValidateRunnerDownload(ctx context.Context) error
}

// egressValidator는 EgressValidator의 구현체입니다.
type egressValidator struct {
	dialer *net.Dialer
}

// NewEgressValidator는 새로운 EgressValidator를 생성합니다.
func NewEgressValidator() EgressValidator {
	return &egressValidator{
		dialer: &net.Dialer{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateGitHubAPI는 GitHub API 액세스를 검증합니다.
func (v *egressValidator) ValidateGitHubAPI(ctx context.Context) error {
	// GitHub API 도메인 목록
	domains := []string{
		"api.github.com:443",
		"github.com:443",
	}

	for _, domain := range domains {
		conn, err := v.dialer.DialContext(ctx, "tcp", domain)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", domain, err)
		}
		conn.Close()
	}

	return nil
}

// ValidateRunnerDownload는 runner 다운로드 액세스를 검증합니다.
func (v *egressValidator) ValidateRunnerDownload(ctx context.Context) error {
	// GitHub runner 다운로드 도메인
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
		conn.Close()
	}

	return nil
}

// AuditLogger는 runner 작업을 기록합니다.
// REQ-SEC-008
type AuditLogger interface {
	LogInstall(ctx context.Context, version string) error
	LogRegister(ctx context.Context, runnerID string) error
}

// auditLogger는 AuditLogger의 구현체입니다.
type auditLogger struct {
	auditDir string
}

// NewAuditLogger는 새로운 AuditLogger를 생성합니다.
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

// LogInstall은 설치 작업을 기록합니다.
func (l *auditLogger) LogInstall(ctx context.Context, version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// 감사 디렉토리 생성
	if err := os.MkdirAll(l.auditDir, 0755); err != nil {
		return fmt.Errorf("create audit directory: %w", err)
	}

	// 감사 로그 파일 생성
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	logPath := filepath.Join(l.auditDir, fmt.Sprintf("install_%s.log", timestamp))

	entry := fmt.Sprintf("[%s] INSTALL version=%s\n", timestamp, version)

	if err := os.WriteFile(logPath, []byte(entry), 0644); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}

// LogRegister는 등록 작업을 기록합니다.
func (l *auditLogger) LogRegister(ctx context.Context, runnerID string) error {
	if runnerID == "" {
		return fmt.Errorf("runner ID cannot be empty")
	}

	// 감사 디렉토리 생성
	if err := os.MkdirAll(l.auditDir, 0755); err != nil {
		return fmt.Errorf("create audit directory: %w", err)
	}

	// 감사 로그 파일 생성
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	logPath := filepath.Join(l.auditDir, fmt.Sprintf("register_%s.log", timestamp))

	entry := fmt.Sprintf("[%s] REGISTER runner_id=%s\n", timestamp, runnerID)

	if err := os.WriteFile(logPath, []byte(entry), 0644); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}
