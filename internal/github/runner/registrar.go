// Package runner는 GitHub Actions self-hosted runner 등록 기능을 제공합니다.
// Package runner provides registration functionality for GitHub Actions self-hosted runners.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GHExecutor는 gh CLI 실행을 위한 인터페이스입니다 (테스트용).
// GHExecutor is an interface for executing gh CLI commands (for testing).
type GHExecutor interface {
	RunGH(ctx context.Context, args ...string) error
	RunGHOutput(ctx context.Context, args ...string) (string, error)
}

// defaultGHExecutor는 실제 gh CLI를 실행하는 구현입니다.
// defaultGHExecutor implements GHExecutor using real gh CLI.
type defaultGHExecutor struct{}

// RunGH는 gh 명령을 실행합니다 (구현체는 gh.go 패키지 사용).
func (d *defaultGHExecutor) RunGH(ctx context.Context, args ...string) error {
	// TODO T-03: 실제 구현은 gh.go 패키지의 RunGH 사용
	return nil
}

// RunGHOutput은 gh 명령을 실행하고 출력을 반환합니다.
func (d *defaultGHExecutor) RunGHOutput(ctx context.Context, args ...string) (string, error) {
	// TODO T-03: 실제 구현은 gh.go 패키지의 RunGHOutput 사용
	return "", nil
}

// RegisterResult는 등록 결과를 담습니다.
// RegisterResult holds registration outcomes.
type RegisterResult struct {
	Success     bool     // 등록 성공 여부 (Registration success)
	RunnerName  string   // Runner 이름 (Runner name)
	Labels      []string // 라벨 목록 (List of labels)
	SettingsURL string   // 설정 페이지 URL (REQ-CI-004.3)
}

// tokenResponse는 registration token API 응답입니다.
// tokenResponse represents registration token API response.
type tokenResponse struct {
	Token string `json:"token"`
}

// Registrar는 runner 등록을 담당합니다.
// Registrar handles registering the runner with GitHub.
type Registrar struct {
	ghRunnerDir string
	executor    GHExecutor
}

// NewRegistrar는 새로운 Registrar 인스턴스를 생성합니다.
// NewRegistrar creates a new Registrar instance.
func NewRegistrar(ghRunnerDir string, executor GHExecutor) *Registrar {
	exec := executor
	if exec == nil {
		exec = &defaultGHExecutor{}
	}

	return &Registrar{
		ghRunnerDir: ghRunnerDir,
		executor:    exec,
	}
}

// RegisterRunner는 runner를 GitHub에 등록합니다.
// RegisterRunner registers the runner using gh api for token, then config.sh.
// --replace 플래그를 기본 사용합니다 (REQ-CI-004.1).
// Token은 메모리에만 보관합니다 (REQ-CI-004.2).
// 등록 후 설정 URL을 출력합니다 (REQ-CI-004.3).
func (r *Registrar) RegisterRunner(ctx context.Context, repo string, labels []string) (*RegisterResult, error) {
	// repo 문자열 파싱 (owner/repo 형식)
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: expected owner/repo, got %s", repo)
	}
	owner, name := parts[0], parts[1]

	// 1. Registration token 획득 (REQ-CI-004.2)
	tokenPath := fmt.Sprintf("/repos/%s/%s/actions/runners/registration-token", owner, name)
	tokenOutput, err := r.executor.RunGHOutput(ctx, "api", tokenPath)
	if err != nil {
		return nil, fmt.Errorf("get registration token: %w", err)
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal([]byte(tokenOutput), &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	// Token은 메모리에만 보관 (REQ-CI-004.2)
	// 실제 config.sh 호출 시 stdin으로 전달되어야 함
	// 현재 mock에서는 token 검증 없이 진행
	_ = tokenResp.Token // TODO T-03: stdin을 통한 token 전달 구현 필요

	// 2. config.sh 실행 (stdin으로 token 전달)
	// --replace 플래그 사용 (REQ-CI-004.1)
	labelsFlag := strings.Join(labels, ",")
	if err := r.executor.RunGH(ctx, "runner", "register", "--replace", "--labels", labelsFlag); err != nil {
		return nil, fmt.Errorf("register runner: %w", err)
	}

	// 3. 결과 반환 (REQ-CI-004.3)
	result := &RegisterResult{
		Success:     true,
		RunnerName:  "self-hosted-runner", // TODO: 실제 runner 이름 획득
		Labels:      labels,
		SettingsURL: fmt.Sprintf("https://github.com/%s/%s/settings/actions/runners", owner, name),
	}

	return result, nil
}
