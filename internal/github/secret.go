package github

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// SecretManager는 GitHub 리포지토리 시크릿을 관리하는 인터페이스입니다.
type SecretManager interface {
	SetSecret(ctx context.Context, repo, name, value string) error
	GetSecret(ctx context.Context, repo, name string) (string, error)
	ListSecrets(ctx context.Context, repo string) ([]string, error)
	DeleteSecret(ctx context.Context, repo, name string) error
}

// GHSecretExecutor는 gh CLI 호출을 위한 인터페이스입니다 (테스트 가능).
type GHSecretExecutor interface {
	RunGH(ctx context.Context, args ...string) error
	RunGHOutput(ctx context.Context, args ...string) (string, error)
}

// secretSetterWithStdin는 stdin을 통한 secret 설정을 지원하는 인터페이스입니다.
type secretSetterWithStdin interface {
	RunGHWithStdin(ctx context.Context, stdin string, args ...string) error
}

// realGHExecutor는 gh CLI를 직접 실행하는 실제 구현입니다.
// gh.go의 execGH를 재사용합니다.
type realGHExecutor struct{}

// RunGH는 gh CLI 명령어를 실행합니다 (출력 없음).
func (r *realGHExecutor) RunGH(ctx context.Context, args ...string) error {
	// 빈 문자열 dir로 execGH 호출
	_, err := execGH(ctx, "", args...)
	return err
}

// RunGHOutput은 gh CLI 명령어를 실행하고 출력을 반환합니다.
func (r *realGHExecutor) RunGHOutput(ctx context.Context, args ...string) (string, error) {
	return execGH(ctx, "", args...)
}

// RunGHWithStdin은 stdin에 값을 전달하여 gh CLI 명령어를 실행합니다.
func (r *realGHExecutor) RunGHWithStdin(ctx context.Context, stdinValue string, args ...string) error {
	ghBinOnce.Do(func() {
		ghBinPath, ghBinErr = exec.LookPath("gh")
	})
	if ghBinErr != nil {
		return fmt.Errorf("gh lookup: %w", ghBinErr)
	}

	cmd := exec.CommandContext(ctx, ghBinPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// stdin 설정
	cmd.Stdin = bytes.NewBufferString(stdinValue)

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("gh %s: %s: %w", args[0], errMsg, err)
	}

	return nil
}

// GHSecretManager는 SecretManager의 gh CLI 기반 구현입니다.
type GHSecretManager struct {
	executor GHSecretExecutor
}

// NewGHSecretManager는 새로운 GHSecretManager를 생성합니다.
func NewGHSecretManager(executor GHSecretExecutor) *GHSecretManager {
	return &GHSecretManager{
		executor: executor,
	}
}

// NewRealGHSecretManager는 실제 gh CLI를 사용하는 GHSecretManager를 생성합니다.
func NewRealGHSecretManager() *GHSecretManager {
	return &GHSecretManager{
		executor: &realGHExecutor{},
	}
}

// SetSecret은 GitHub 리포지토리에 시크릿을 설정합니다.
// 절대 디스크에 쓰지 않고 stdin을 통해 전달합니다 (REQ-SEC-002).
func (m *GHSecretManager) SetSecret(ctx context.Context, repo, name, value string) error {
	// gh secret set NAME -R REPO
	args := []string{"secret", "set", name, "-R", repo}

	// 디버그 로그에서 값 마스킹
	maskedValue := maskSecret(value)
	fmt.Printf("[DEBUG] Setting secret %s=%s for repo %s\n", name, maskedValue, repo)

	// stdin을 통한 secret 설정
	if stdinExecutor, ok := m.executor.(secretSetterWithStdin); ok {
		return stdinExecutor.RunGHWithStdin(ctx, value, args...)
	}

	// Fallback: 값이 없는 경우 실행 (테스트용 MockExecutor)
	return m.executor.RunGH(ctx, args...)
}

// GetSecret은 GitHub 리포지토리에서 시크릿 값을 가져옵니다.
func (m *GHSecretManager) GetSecret(ctx context.Context, repo, name string) (string, error) {
	// gh secret set은 값을 가져오는 기능을 제공하지 않음
	// 이 메서드는 현재 구현되지 않음
	return "", fmt.Errorf("GetSecret: not implemented - gh CLI does not support retrieving secret values")
}

// ListSecrets은 GitHub 리포지토리의 모든 시크릿 이름을 나열합니다.
func (m *GHSecretManager) ListSecrets(ctx context.Context, repo string) ([]string, error) {
	// gh secret list -R REPO
	args := []string{"secret", "list", "-R", repo}

	output, err := m.executor.RunGHOutput(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("listing secrets: %w", err)
	}

	// 출력 파싱: 각 줄에서 시크릿 이름 추출
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var secrets []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		// gh secret list 출력 형식: "NAME\tUpdated at DATE"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) > 0 && parts[0] != "" {
			secrets = append(secrets, parts[0])
		}
	}

	return secrets, nil
}

// DeleteSecret은 GitHub 리포지토리에서 시크릿을 삭제합니다.
func (m *GHSecretManager) DeleteSecret(ctx context.Context, repo, name string) error {
	// gh secret delete NAME -R REPO
	args := []string{"secret", "delete", name, "-R", repo}

	fmt.Printf("[DEBUG] Deleting secret %s from repo %s\n", name, repo)

	return m.executor.RunGH(ctx, args...)
}

// maskSecret은 시크릿 값을 디버그용으로 마스킹합니다.
func maskSecret(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	// 첫 문자 + 마지막 4字符만 표시
	return value[:1] + "..." + value[len(value)-4:]
}
