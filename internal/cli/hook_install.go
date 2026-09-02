package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// moaiPrePushMarker is the identifier written on line 2 of the hook file.
// Its presence signals that the hook was installed by MoAI-ADK and can be
// safely overwritten on subsequent installs.
const moaiPrePushMarker = "# MoAI-ADK pre-push hook"

// moaiPrePushProvenanceName is the pre-push provenance sidecar basename; see
// hookTier.provenanceName (hook_install_shared.go) for the rationale.
const moaiPrePushProvenanceName = ".moai-pre-push.sha256"

// prePushBackupPrefix is the pre-push backup basename prefix; see
// hookTier.backupPrefix (hook_install_shared.go) for the rationale.
const prePushBackupPrefix = "pre-push.bak."

// errPrePushBackupFailed wraps a pre-push backup-write failure so the optional
// wrapper can turn it into a warning while leaving the hook untouched (the
// REQ-PCP-010 sub-case (a) equivalent for the push tier, t257). It never
// reaches the caller as a failure.
var errPrePushBackupFailed = errors.New("pre-push backup failed")

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

# Capture git's ref-update stdin once, before any later step consumes it. git
# passes lines of "<local ref> <local oid> <remote ref> <remote oid>"; cat exits
# 0 on empty input, so this is safe under "set -eu".
REFS="$(cat)"

REPO_ROOT="$(git rev-parse --show-toplevel)"
LOG_DIR="$REPO_ROOT/.moai/logs"
LOG_FILE="$LOG_DIR/prepush-bypass.log"
mkdir -p "$LOG_DIR" 2>/dev/null || true

START_TS="$(date +%s)"

if [ -f "$REPO_ROOT/Makefile" ]; then
    if make -C "$REPO_ROOT" -s ci-local >/dev/null; then
        OUTCOME="pass"
        EXIT_CODE=0
    else
        EXIT_CODE=$?
        OUTCOME="fail"
    fi
else
    # No Makefile — skip ci-local (end-user project without a CI mirror).
    OUTCOME="skip (no Makefile)"
    EXIT_CODE=0
    printf '[pre-push] No Makefile found — skipping ci-local\n' >&2
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

# Commit-message convention validation. Runs only after ci-local passes (the
# fail branch above exits first). Skipped when moai is not on PATH so projects
# without moai installed are unaffected. The convention engine self-gates on its
# own enforce_on_push config, so this is a no-op unless enforcement is enabled.
if command -v moai >/dev/null 2>&1; then
    ZERO="0000000000000000000000000000000000000000"
    SUBJECTS="$(
        printf '%s\n' "$REFS" | while read -r local_ref local_oid remote_ref remote_oid; do
            [ -z "$local_oid" ] && continue
            if [ "$local_oid" = "$ZERO" ]; then
                # Branch deletion: nothing is being pushed for this ref.
                continue
            elif [ "$remote_oid" = "$ZERO" ]; then
                # New remote ref: enumerate commits not already on any remote.
                git log --format=%s "$local_oid" --not --remotes
            else
                git log --format=%s "$remote_oid".."$local_oid"
            fi
        done
    )"
    if [ -n "$SUBJECTS" ]; then
        printf '%s\n' "$SUBJECTS" | moai hook pre-push
    fi
fi

exit 0
`

// PrePushInstaller installs the MoAI-ADK pre-push git hook. It shares the
// hook-preservation machinery with the pre-commit installer
// (hook_install_shared.go): an existing marker-bearing hook is classified by
// provenance before any overwrite, and a user-modified hook is backed up and
// disclosed rather than silently replaced (t257).
type PrePushInstaller struct {
	// repoRoot is the root of the git repository.
	repoRoot string

	// content is the hook body this installer writes — the "incoming" operand
	// of the attribution. Defaults to prePushHookContent; overridden in tests
	// to construct a version bump without editing the shipped constant.
	content string

	// lastAttribution is the classification of the most recent install run
	// against an existing marker-bearing hook, or nil when the run classified
	// nothing (no existing hook, a non-MoAI hook, or skip=true).
	lastAttribution *hookAttribution

	// lastProvenanceErr holds a provenance-write failure from the most recent
	// run. The write happens after a hook write that already succeeded, so it
	// must not fail the caller; it is recorded rather than discarded so the
	// wrapper can warn about it.
	lastProvenanceErr error

	// lastBackupPath is the backup copy written for a user-modified hook in the
	// most recent run, or "" when no backup was taken. The wrapper reads it to
	// emit the disclosure notice.
	lastBackupPath string

	// now supplies the timestamp for backup names. A seam so tests can force
	// the same stamp twice and construct the occupied-path case.
	now func() time.Time
}

// NewPrePushInstaller creates a PrePushInstaller for the given repository root.
func NewPrePushInstaller(repoRoot string) *PrePushInstaller {
	return &PrePushInstaller{repoRoot: repoRoot, content: prePushHookContent, now: time.Now}
}

// InstallPrePushHook writes the pre-push hook to .git/hooks/pre-push with mode 0755.
//
// Behaviour:
//   - skip=true: no-op, returns nil.
//   - File does not exist: creates it, plus its provenance record.
//   - File exists with MoAI-ADK marker (first 3 lines): classified by the
//     shared three-way attribution — an unmodified hook is replaced quietly;
//     a user-modified (or undecidable-legacy, differing) hook is backed up to
//     `pre-push.bak.<stamp>` BEFORE the replacement.
//   - File exists WITHOUT marker: returns ErrUserHookExists without modifying the file.
func (p *PrePushInstaller) InstallPrePushHook(skip bool) error {
	p.lastAttribution = nil
	p.lastProvenanceErr = nil
	p.lastBackupPath = ""

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
		// MoAI hook found — classify it before overwriting, exactly as the
		// pre-commit installer does: the verdict decides backup + disclosure.
		installed, err := os.ReadFile(hookPath)
		if err != nil {
			return fmt.Errorf("read existing hook: %w", err)
		}
		attribution := classifyHook(installed, readHookProvenance(hookDir, prePushTier), []byte(p.content))
		p.lastAttribution = &attribution

		// A user-modified hook is copied aside BEFORE the replacement is
		// written, so the patch is recoverable the moment the loss-bearing
		// overwrite happens (REQ-PCP-003 equivalent for the push tier).
		if attribution.Class == hookUserModified {
			backupPath, err := backupHook(hookDir, installed, p.now(), prePushTier)
			if err != nil {
				// The backup sits before the hook write, so a failure here
				// means the patch cannot be made recoverable — the hook is
				// left untouched and the caller is not failed; the wrapper
				// turns this into a warning. Overwriting anyway would destroy
				// the patch with no recoverable copy.
				return fmt.Errorf("%w: %w", errPrePushBackupFailed, err)
			}
			p.lastBackupPath = backupPath
		}
	}

	if err := os.WriteFile(hookPath, []byte(p.content), 0o755); err != nil {
		return fmt.Errorf("write pre-push hook: %w", err)
	}

	// Record what was just written, so the next run can attribute a
	// difference. A failure here follows a hook write that already succeeded,
	// so it must not fail the caller: the missing record self-heals on the
	// next run, which finds installed == incoming and re-stamps.
	p.lastProvenanceErr = writeHookProvenance(hookDir, p.content, prePushTier)

	return nil
}

// installPrePushHookOptional installs the pre-push hook into projectRoot's .git/hooks/
// unless skip is true. Friendly, non-fatal: prints progress to out, warnings to
// warn (a writer distinct from out — callers bind it to the command's stderr,
// the REQ-PCP-004 wiring the pre-commit wrapper uses), returns nothing. Used by
// `moai init` and `moai update` to install the hook consistently
// (REQ-CIAUT-002).
//
// If a non-MoAI user hook is present, this function preserves it and prints a
// note. Other errors are reported as warnings; project init/update is never
// blocked by hook installation failures.
func installPrePushHookOptional(projectRoot string, skip bool, out, warn io.Writer) {
	installer := NewPrePushInstaller(projectRoot)
	err := installer.InstallPrePushHook(skip)
	switch {
	case err == nil:
		if !skip {
			_, _ = fmt.Fprintln(out, "  Pre-push hook installed (.git/hooks/pre-push)")
			if installer.lastBackupPath != "" {
				// The notice carries exactly two elements — the backup path and
				// the fact of the replacement — on the warning writer, never
				// the progress writer (which `moai update` binds to stdout,
				// where a redirected run swallows a data-loss notice whole).
				_, _ = fmt.Fprintf(warn, "  Warning: user-modified pre-push hook was replaced; previous hook backed up at %s\n", installer.lastBackupPath)
			}
			if installer.lastProvenanceErr != nil {
				// The hook write already succeeded, so the replacement stands;
				// the missing record self-heals on the next run.
				_, _ = fmt.Fprintf(warn, "  Warning: pre-push provenance record not written: %v\n", installer.lastProvenanceErr)
			}
		}
	case errors.Is(err, ErrUserHookExists):
		_, _ = fmt.Fprintln(out, "  Note: existing pre-push hook preserved (no MoAI-ADK marker found)")
	case errors.Is(err, errPrePushBackupFailed):
		// No backup could be written, so the hook was left exactly as found —
		// nothing was installed and nothing was lost.
		_, _ = fmt.Fprintf(warn, "  Warning: pre-push hook left unchanged (backup could not be written): %v\n", err)
	default:
		if installer.lastBackupPath != "" {
			// The backup succeeded but the replacement write failed, so no
			// replacement notice (the hook was not replaced). Name the orphan
			// backup so the artifact beside the unchanged hook is explained
			// rather than mysterious.
			_, _ = fmt.Fprintf(warn, "  Warning: pre-push hook install failed: %v (previous hook backed up at %s)\n", err, installer.lastBackupPath)
		} else {
			_, _ = fmt.Fprintf(out, "  Warning: pre-push hook install failed: %v\n", err)
		}
	}
}

// fileHasMoaiMarker reads the first 3 lines of the given file and returns true
// if any of them contain the pre-push MoAI-ADK marker string.
//
// It is a thin wrapper over fileHasMarker preserving the pre-push installer's
// behaviour byte-for-byte; callers checking a different marker use fileHasMarker.
func fileHasMoaiMarker(path string) (bool, error) {
	return fileHasMarker(path, moaiPrePushMarker)
}

// fileHasMarker reads the first 3 lines of the given file and returns true if
// any of them contain the supplied marker string. Shared by the pre-push and
// pre-commit installers so each hook keeps its own marker while reusing the
// first-3-lines detection logic.
func fileHasMarker(path, marker string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() && lineCount < 3 {
		if strings.Contains(scanner.Text(), marker) {
			return true, nil
		}
		lineCount++
	}
	return false, scanner.Err()
}
