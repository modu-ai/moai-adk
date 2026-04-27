package github

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// SecretManager manages GitHub repository secrets.
type SecretManager interface {
	SetSecret(ctx context.Context, repo, name, value string) error
	GetSecret(ctx context.Context, repo, name string) (string, error)
	ListSecrets(ctx context.Context, repo string) ([]string, error)
	DeleteSecret(ctx context.Context, repo, name string) error
}

// GHSecretExecutor is an interface for gh CLI invocation (testable).
type GHSecretExecutor interface {
	RunGH(ctx context.Context, args ...string) error
	RunGHOutput(ctx context.Context, args ...string) (string, error)
}

// secretSetterWithStdin supports setting secrets via stdin.
type secretSetterWithStdin interface {
	RunGHWithStdin(ctx context.Context, stdin string, args ...string) error
}

// realGHExecutor executes gh CLI directly. Reuses execGH from gh.go.
type realGHExecutor struct{}

func (r *realGHExecutor) RunGH(ctx context.Context, args ...string) error {
	_, err := execGH(ctx, "", args...)
	return err
}

func (r *realGHExecutor) RunGHOutput(ctx context.Context, args ...string) (string, error) {
	return execGH(ctx, "", args...)
}

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

// GHSecretManager is a gh CLI-based implementation of SecretManager.
type GHSecretManager struct {
	executor GHSecretExecutor
}

// NewGHSecretManager creates a new GHSecretManager with the given executor.
func NewGHSecretManager(executor GHSecretExecutor) *GHSecretManager {
	return &GHSecretManager{
		executor: executor,
	}
}

// NewRealGHSecretManager creates a GHSecretManager using the real gh CLI.
func NewRealGHSecretManager() *GHSecretManager {
	return &GHSecretManager{
		executor: &realGHExecutor{},
	}
}

// SetSecret sets a GitHub repository secret via stdin (REQ-SEC-002: never write to disk).
func (m *GHSecretManager) SetSecret(ctx context.Context, repo, name, value string) error {
	args := []string{"secret", "set", name, "-R", repo}

	maskedValue := MaskSecret(value)
	fmt.Printf("[DEBUG] Setting secret %s=%s for repo %s\n", name, maskedValue, repo)

	if stdinExecutor, ok := m.executor.(secretSetterWithStdin); ok {
		return stdinExecutor.RunGHWithStdin(ctx, value, args...)
	}

	return m.executor.RunGH(ctx, args...)
}

// GetSecret returns an error — gh CLI does not support retrieving secret values.
func (m *GHSecretManager) GetSecret(_ context.Context, _ string, _ string) (string, error) {
	return "", fmt.Errorf("GetSecret: not implemented - gh CLI does not support retrieving secret values")
}

// ListSecrets lists all secret names in a GitHub repository.
func (m *GHSecretManager) ListSecrets(ctx context.Context, repo string) ([]string, error) {
	args := []string{"secret", "list", "-R", repo}

	output, err := m.executor.RunGHOutput(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("listing secrets: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var secrets []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		// gh secret list output format: "NAME\tUpdated at DATE"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) > 0 && parts[0] != "" {
			secrets = append(secrets, parts[0])
		}
	}

	return secrets, nil
}

// DeleteSecret deletes a secret from a GitHub repository.
func (m *GHSecretManager) DeleteSecret(ctx context.Context, repo, name string) error {
	args := []string{"secret", "delete", name, "-R", repo}

	fmt.Printf("[DEBUG] Deleting secret %s from repo %s\n", name, repo)

	return m.executor.RunGH(ctx, args...)
}

// MaskSecret masks a secret value for debug output.
func MaskSecret(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	return value[:1] + "..." + value[len(value)-4:]
}
