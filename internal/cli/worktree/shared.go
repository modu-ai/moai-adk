package worktree

// Shared helpers for the worktree subcommands. These were extracted from the
// retired `worktree new` command, whose deletion would otherwise have taken
// their surviving consumers (clean, sync, done) with it.

import (
	"os/exec"
	"strings"
)

// gitWorktreeCmd executes a git subcommand for the worktree CLI.
// Overridable in tests so they never touch the network or a real repository.
var gitWorktreeCmd = func(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

// gitRepoRootFunc returns the absolute path of the git repository root via
// `git rev-parse --show-toplevel`. Callers must anchor on this rather than
// os.Getwd(): a test process runs with cwd=package-dir, so writes derived from
// the process cwd would land inside internal/cli/worktree/ instead of the repo
// root. Overridable in tests to inject a tempDir.
var gitRepoRootFunc = func() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	return strings.TrimSpace(string(out)), err
}

// resolveSpecBranch converts SPEC-ID patterns to branch names.
// e.g., "SPEC-AUTH-001" -> "feature/SPEC-AUTH-001"
// Regular branch names pass through unchanged.
func resolveSpecBranch(name string) string {
	if isSpecID(name) {
		return "feature/" + name
	}
	return name
}

// isSpecID reports whether name follows the SPEC-<DOMAIN>-<NNN> shape.
func isSpecID(name string) bool {
	if !strings.HasPrefix(name, "SPEC-") {
		return false
	}
	parts := strings.SplitN(name, "-", 3)
	return len(parts) >= 3 && parts[2] != ""
}
