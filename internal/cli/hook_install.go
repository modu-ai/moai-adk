package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// moaiPrePushMarker is the identifier written on line 2 of the hook file.
// Its presence signals that the hook was installed by MoAI-ADK and can be
// safely overwritten on subsequent installs.
const moaiPrePushMarker = "# MoAI-ADK pre-push hook"

// ErrUserHookExists is returned when a pre-existing hook without the MoAI
// marker is found. The caller should inform the user and skip installation.
var ErrUserHookExists = errors.New("pre-existing user hook found without MoAI-ADK marker")

// prePushHookContent is the canonical content of the pre-push hook.
// Kept as a constant so the installer and tests share a single source of truth.
//
// MUST stay byte-identical with internal/template/templates/.git_hooks/pre-push.
// TestPrePushTemplateMatchesConstant enforces this; do not edit one without the other.
const prePushHookContent = `#!/bin/sh
# MoAI-ADK pre-push hook — runs make ci-local; logs invocation outcome
# Bypass via: SKIP_MOAI_PREPUSH=1 git push   (logged on next invocation)
# To disable permanently: remove this file or pass --no-hooks to moai update
set -eu

if [ "${SKIP_MOAI_PREPUSH:-0}" = "1" ]; then
    printf '[pre-push] SKIP_MOAI_PREPUSH=1 -- bypass requested\n' >&2
    exit 0
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
LOG_DIR="$REPO_ROOT/.moai/logs"
LOG_FILE="$LOG_DIR/prepush-bypass.log"
mkdir -p "$LOG_DIR" 2>/dev/null || true

START_TS="$(date +%s)"

if make -C "$REPO_ROOT" -s ci-local >/dev/null; then
    OUTCOME="pass"
    EXIT_CODE=0
else
    EXIT_CODE=$?
    OUTCOME="fail"
fi

END_TS="$(date +%s)"
DURATION=$((END_TS - START_TS))
USER_NAME="${USER:-$(id -un 2>/dev/null || printf 'unknown')}"
BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || printf 'unknown')"

printf '%s\t%s\t%s\t%s\t%ds\n' "$END_TS" "$USER_NAME" "$BRANCH" "$OUTCOME" "$DURATION" >> "$LOG_FILE" 2>/dev/null || true

if [ "$OUTCOME" = "fail" ]; then
    printf '\n[pre-push] FAILED: local CI mirror reported errors.\n' >&2
    printf '[pre-push] Hint: make fmt && make lint && make test\n' >&2
    printf '[pre-push] Override (logged on next invocation): SKIP_MOAI_PREPUSH=1 git push\n' >&2
    exit "$EXIT_CODE"
fi

exit 0
`

// PrePushInstaller installs the MoAI-ADK pre-push git hook.
type PrePushInstaller struct {
	// repoRoot is the root of the git repository.
	repoRoot string
}

// NewPrePushInstaller creates a PrePushInstaller for the given repository root.
func NewPrePushInstaller(repoRoot string) *PrePushInstaller {
	return &PrePushInstaller{repoRoot: repoRoot}
}

// InstallPrePushHook writes the pre-push hook to .git/hooks/pre-push with mode 0755.
//
// Behaviour:
//   - skip=true: no-op, returns nil.
//   - File does not exist: creates it.
//   - File exists with MoAI-ADK marker (first 3 lines): overwrites safely.
//   - File exists WITHOUT marker: returns ErrUserHookExists without modifying the file.
func (p *PrePushInstaller) InstallPrePushHook(skip bool) error {
	if skip {
		return nil
	}

	hookDir := filepath.Join(p.repoRoot, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return fmt.Errorf("create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hookDir, "pre-push")

	// Check if the file already exists.
	if _, err := os.Stat(hookPath); err == nil {
		// File exists — inspect first 3 lines for MoAI marker.
		hasMoaiMarker, err := fileHasMoaiMarker(hookPath)
		if err != nil {
			return fmt.Errorf("read existing hook: %w", err)
		}
		if !hasMoaiMarker {
			return ErrUserHookExists
		}
		// MoAI hook found — overwrite.
	}

	if err := os.WriteFile(hookPath, []byte(prePushHookContent), 0o755); err != nil {
		return fmt.Errorf("write pre-push hook: %w", err)
	}

	return nil
}

// installPrePushHookOptional installs the pre-push hook into projectRoot's .git/hooks/
// unless skip is true. Friendly, non-fatal: prints status to out, returns nothing.
// Used by `moai init` and `moai update` to install the hook consistently
// (REQ-CIAUT-002).
//
// If a non-MoAI user hook is present, this function preserves it and prints a
// note. Other errors are reported as warnings; project init/update is never
// blocked by hook installation failures.
func installPrePushHookOptional(projectRoot string, skip bool, out io.Writer) {
	installer := NewPrePushInstaller(projectRoot)
	err := installer.InstallPrePushHook(skip)
	switch {
	case err == nil:
		if !skip {
			_, _ = fmt.Fprintln(out, "  Pre-push hook installed (.git/hooks/pre-push)")
		}
	case errors.Is(err, ErrUserHookExists):
		_, _ = fmt.Fprintln(out, "  Note: existing pre-push hook preserved (no MoAI-ADK marker found)")
	default:
		_, _ = fmt.Fprintf(out, "  Warning: pre-push hook install failed: %v\n", err)
	}
}

// fileHasMoaiMarker reads the first 3 lines of the given file and returns true
// if any of them contain the MoAI-ADK marker string.
func fileHasMoaiMarker(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() && lineCount < 3 {
		if strings.Contains(scanner.Text(), moaiPrePushMarker) {
			return true, nil
		}
		lineCount++
	}
	return false, scanner.Err()
}
