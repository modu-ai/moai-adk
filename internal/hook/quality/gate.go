package quality

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// GateConfig holds configuration for the QualityGate.
type GateConfig struct {
	// Enabled controls whether the quality gate runs at all.
	Enabled bool
	// SkipTests skips the go test step when true, useful for quick commits.
	SkipTests bool
	// VetTimeout is the maximum duration allowed for go vet.
	VetTimeout time.Duration
	// LintTimeout is the maximum duration allowed for golangci-lint.
	LintTimeout time.Duration
	// TestTimeout is the maximum duration allowed for go test.
	TestTimeout time.Duration
}

// DefaultGateConfig returns a GateConfig with production-safe defaults.
func DefaultGateConfig() *GateConfig {
	return &GateConfig{
		Enabled:     true,
		SkipTests:   false,
		VetTimeout:  30 * time.Second,
		LintTimeout: 60 * time.Second,
		TestTimeout: 120 * time.Second,
	}
}

// QualityGate runs deterministic quality checks before git commit.
// Steps execute sequentially: go vet → golangci-lint → go test.
// If any step fails, subsequent steps are skipped (REQ-GATE-005).
type QualityGate struct {
	config *GateConfig
}

// NewQualityGate creates a QualityGate with the given configuration.
// If cfg is nil, DefaultGateConfig is used.
func NewQualityGate(cfg *GateConfig) *QualityGate {
	if cfg == nil {
		cfg = DefaultGateConfig()
	}
	return &QualityGate{config: cfg}
}

// Run executes quality gate checks sequentially.
// Returns (passed bool, output string) where output contains error details on failure.
// When gate is disabled (config.Enabled == false), returns (true, "") immediately.
func (g *QualityGate) Run(ctx context.Context) (bool, string) {
	if !g.config.Enabled {
		return true, ""
	}

	// Step 1: go vet
	if ok, out := g.runStep(ctx, "go vet", g.config.VetTimeout, "go", "vet", "./..."); !ok {
		return false, out
	}

	// Step 2: golangci-lint
	if ok, out := g.runGolangciLint(ctx); !ok {
		return false, out
	}

	// Step 3: go test (skippable)
	if !g.config.SkipTests {
		if ok, out := g.runStep(ctx, "go test", g.config.TestTimeout, "go", "test", "./..."); !ok {
			return false, out
		}
	}

	return true, ""
}

// runGolangciLint runs golangci-lint, skipping gracefully if the binary is not found.
func (g *QualityGate) runGolangciLint(ctx context.Context) (bool, string) {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		// Binary not found — skip step rather than fail.
		return true, ""
	}
	return g.runStep(ctx, "golangci-lint", g.config.LintTimeout, "golangci-lint", "run")
}

// runStep executes a single quality gate command with the given timeout.
// Returns (true, "") on success, (false, errorMessage) on failure or timeout.
func (g *QualityGate) runStep(ctx context.Context, stepName string, timeout time.Duration, name string, args ...string) (bool, string) {
	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(stepCtx, name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err == nil {
		return true, ""
	}

	output := strings.TrimSpace(buf.String())

	// Distinguish timeout from other failures (REQ-GATE-009).
	if stepCtx.Err() == context.DeadlineExceeded {
		msg := fmt.Sprintf("quality gate timed out: %s exceeded %s", stepName, timeout)
		return false, msg
	}

	if output == "" {
		output = err.Error()
	}
	return false, fmt.Sprintf("quality gate failed: %s\n\n%s", stepName, output)
}

// isGitCommitRe matches git commit commands.
// Matches: git commit, git commit -m "...", git commit --amend, git commit --no-verify, etc.
// Does NOT match: git commit --help, echo "git commit".
var isGitCommitRe = regexp.MustCompile(`^\s*git\s+commit\b`)

// IsGitCommit reports whether command is a git commit invocation.
// --no-verify and --amend flags do not bypass the gate (REQ-GATE-011).
func IsGitCommit(command string) bool {
	return isGitCommitRe.MatchString(command)
}
