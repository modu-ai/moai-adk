// Package runner provides registration functionality for GitHub Actions self-hosted runners.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GHExecutor is an interface for executing gh CLI commands (for testing).
type GHExecutor interface {
	RunGH(ctx context.Context, args ...string) error
	RunGHOutput(ctx context.Context, args ...string) (string, error)
}

// defaultGHExecutor implements GHExecutor using real gh CLI.
type defaultGHExecutor struct{}

// RunGH executes gh commands (implementation uses gh.go package).
func (d *defaultGHExecutor) RunGH(ctx context.Context, args ...string) error {
	// TODO T-03: Actual implementation uses RunGH from gh.go package
	return nil
}

// RunGHOutput executes gh commands and returns output.
func (d *defaultGHExecutor) RunGHOutput(ctx context.Context, args ...string) (string, error) {
	// TODO T-03: Actual implementation uses RunGHOutput from gh.go package
	return "", nil
}

// RegisterResult holds registration outcomes.
type RegisterResult struct {
	Success     bool     // Registration success
	RunnerName  string   // Runner name
	Labels      []string // List of labels
	SettingsURL string   // Configuration page URL (REQ-CI-004.3)
}

// tokenResponse represents registration token API response.
type tokenResponse struct {
	Token string `json:"token"`
}

// Registrar handles registering the runner with GitHub.
type Registrar struct {
	ghRunnerDir string
	executor    GHExecutor
}

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

// RegisterRunner registers the runner using gh api for token, then config.sh.
// Uses --replace flag by default (REQ-CI-004.1).
// Token is kept only in memory (REQ-CI-004.2).
// Outputs configuration URL (REQ-CI-004.3).
func (r *Registrar) RegisterRunner(ctx context.Context, repo string, labels []string) (*RegisterResult, error) {
	// Parse repo string (owner/repo format)
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: expected owner/repo, got %s", repo)
	}
	owner, name := parts[0], parts[1]

	// 1. Get registration token (REQ-CI-004.2)
	tokenPath := fmt.Sprintf("/repos/%s/%s/actions/runners/registration-token", owner, name)
	tokenOutput, err := r.executor.RunGHOutput(ctx, "api", tokenPath)
	if err != nil {
		return nil, fmt.Errorf("get registration token: %w", err)
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal([]byte(tokenOutput), &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	// Token is kept only in memory (REQ-CI-004.2)
	// Must be passed via stdin when calling config.sh
	// In mock, no token validation is performed
	_ = tokenResp.Token // TODO T-03: Need implementation for passing token via stdin

	// 2. Execute config.sh (pass token via stdin)
	// Use --replace flag (REQ-CI-004.1)
	labelsFlag := strings.Join(labels, ",")
	if err := r.executor.RunGH(ctx, "runner", "register", "--replace", "--labels", labelsFlag); err != nil {
		return nil, fmt.Errorf("register runner: %w", err)
	}

	// 3. Return result (REQ-CI-004.3)
	result := &RegisterResult{
		Success:     true,
		RunnerName:  "self-hosted-runner", // TODO: Get actual runner name
		Labels:      labels,
		SettingsURL: fmt.Sprintf("https://github.com/%s/%s/settings/actions/runners", owner, name),
	}

	return result, nil
}
