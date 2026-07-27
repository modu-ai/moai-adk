package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/hook/quality"
)

// gateCmd exposes the heavy commit-quality gate as a standalone CLI verb so
// the git pre-commit hook (internal/cli/hook_install_precommit.go) can shell
// out to it without being subject to Claude Code's 5s PreToolUse hook budget.
// This is the relocation surface for SPEC-PRETOOL-GATE-MOVE-001: the same
// QualityGate.Run path that used to silently drop its deny verdict after 5s
// now runs in the user's shell via `moai gate`, with the full 210s step
// budget (vet 30s + lint 60s + test 120s) available.
//
// @MX:ANCHOR: [AUTO] heavy commit-quality gate CLI entry point; invoked by the git pre-commit hook
// @MX:REASON: this verb is the relocation target for the PreToolUse 5s budget defect (census C-2); the pre-commit hook shells out to it
var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "Run the heavy commit-quality gate (vet + lint + test, 16-language detection)",
	Long: `Run the MoAI-ADK quality gate outside the PreToolUse 5s budget.

Invoked by the git pre-commit hook (installed via 'moai init' / 'moai update')
to enforce vet / lint / test at commit time. Runs in the user's shell, so the
full step budget (vet 30s + lint 60s + test 120s per .moai/config/sections/gate.yaml)
is available — unlike the PreToolUse hook path which is silently dropped after 5s.

Exit code 0: gate passed (or no recognized language toolchain detected).
Exit code 1: gate failed; diagnostic output on stderr.

Project directory resolution priority:
  1. $CLAUDE_PROJECT_DIR (when invoked from inside Claude Code)
  2. current working directory

Go build tags are read from .moai/config/build-tags (first non-comment non-blank
line) when present, mirroring the PreToolUse gate path.`,
	RunE: runGate,
}

func init() {
	rootCmd.AddCommand(gateCmd)
}

// runGate executes the quality gate and propagates pass/fail as the process
// exit code. Failure output goes to stderr so git's pre-commit hook surfaces
// it to the caller (verified by M1.c — git pre-commit stderr reaches the Bash
// tool result).
func runGate(cmd *cobra.Command, _ []string) error {
	cfg := quality.DefaultGateConfig()
	cfg.ProjectDir = resolveGateProjectDir()
	cfg.GoBuildTags = readGateGoBuildTags(cfg.ProjectDir)

	// Overall deadline: vet (30s) + lint (60s) + test (120s) + ast-grep + slack.
	// Per-step timeouts inside the gate enforce the real ceilings; this outer
	// deadline is a safety net against a runaway step hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	gate := quality.NewQualityGate(cfg)
	passed, output := gate.Run(ctx)
	if !passed {
		if output != "" {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output)
		}
		return fmt.Errorf("quality gate failed")
	}
	// Pass path: a non-empty output carries a non-blocking notice (e.g. an
	// ast-grep step that skipped because sg is absent). Surface it on stderr
	// without failing the gate.
	if output != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output)
	}
	return nil
}

// resolveGateProjectDir returns the project directory the gate should run
// against, honouring the same $CLAUDE_PROJECT_DIR priority the PreToolUse
// hook path uses (CLAUDE.local.md B7). Falls back to os.Getwd() when the env
// var is unset or empty.
func resolveGateProjectDir() string {
	if dir := strings.TrimSpace(os.Getenv("CLAUDE_PROJECT_DIR")); dir != "" {
		if abs, err := filepath.Abs(dir); err == nil {
			return abs
		}
		return dir
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

// readGateGoBuildTags reads the project Go build tags from
// <dir>/.moai/config/build-tags (first non-empty, non-comment line). Mirrors
// the unexported readGoBuildTags helper in internal/hook/pre_tool.go so the
// standalone gate CLI honours the same build-tag file a project configures
// for the PreToolUse path. Returns "" when the file is absent or unreadable.
//
// The duplication is deliberate: pre_tool.go's helper is an unexported method
// on a handler struct, unreachable from the cli package without a wider
// refactor. Consolidating into a shared internal/hook/quality helper is a
// follow-up; both readers are leaf file reads with no side effects.
func readGateGoBuildTags(dir string) string {
	if dir == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(dir, ".moai", "config", "build-tags"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}
