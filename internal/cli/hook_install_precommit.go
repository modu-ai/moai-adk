package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// moaiPreCommitMarker is the identifier written near the top of the pre-commit
// hook file. Its presence signals that the hook was installed by MoAI-ADK and
// can be safely overwritten on subsequent installs.
const moaiPreCommitMarker = "# MoAI-ADK pre-commit hook"

// moaiPreCommitProvenanceName is the pre-commit provenance sidecar basename;
// see hookTier.provenanceName (hook_install_shared.go) for the rationale.
const moaiPreCommitProvenanceName = ".moai-pre-commit.sha256"

// preCommitBackupPrefix is the pre-commit backup basename prefix; see
// hookTier.backupPrefix (hook_install_shared.go) for the rationale.
const preCommitBackupPrefix = "pre-commit.bak."

// errPreCommitBackupFailed wraps a backup-write failure so the optional
// wrapper can turn it into a warning while leaving the hook untouched
// (REQ-PCP-010 sub-case (a)). It never reaches the caller as a failure.
var errPreCommitBackupFailed = errors.New("pre-commit backup failed")

// preCommitHookContent is the canonical content of the pre-commit hook. It
// runs the fast subset (gofmt -l + go vet on staged Go files) followed by the
// heavy quality gate via `moai gate` (vet + lint + test, 16-language detection).
//
// The heavy gate runs in the user's shell — outside Claude Code's 5s PreToolUse
// hook budget — eliminating the census C-2 silent-drop defect. This is the
// relocation surface for SPEC-PRETOOL-GATE-MOVE-001.
//
// MUST stay byte-identical with internal/template/templates/.git_hooks/pre-commit.
// TestPreCommitTemplateMatchesConstant enforces this; do not edit one without the other.
const preCommitHookContent = `#!/bin/sh
# MoAI-ADK pre-commit hook — fast subset (gofmt + go vet) + heavy gate (moai gate)
# Bypass via: SKIP_MOAI_PRECOMMIT=1 git commit
# Heavy gate (vet + lint + test, 16-language detection) runs in your shell, outside the 5s hook budget.
set -eu

if [ "${SKIP_MOAI_PRECOMMIT:-0}" = "1" ]; then
    printf '[pre-commit] SKIP_MOAI_PRECOMMIT=1 -- bypass requested\n' >&2
    exit 0
fi

# Staged Go files (Added / Copied / Modified; deletions excluded via ACM).
STAGED_GO="$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)"

# --- Fast subset: gofmt + go vet on staged Go files (sub-second; skipped when none staged) ---
if [ -n "$STAGED_GO" ]; then
    # gofmt format check. Skipped when gofmt is not on PATH (non-Go environment).
    if command -v gofmt >/dev/null 2>&1; then
        NEED_FMT="$(
            printf '%s\n' "$STAGED_GO" | while IFS= read -r f; do
                [ -n "$f" ] || continue
                gofmt -l "$f" 2>/dev/null || true
            done
        )"
        if [ -n "$NEED_FMT" ]; then
            printf '\n[pre-commit] FAILED: the following staged files need formatting:\n%s\n' "$NEED_FMT" >&2
            printf '[pre-commit] Hint: gofmt -w <files> && git add <files>\n' >&2
            printf '[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit\n' >&2
            exit 1
        fi
    fi

    # go vet on the affected packages. Skipped when go is not on PATH (non-Go environment).
    if command -v go >/dev/null 2>&1; then
        PKGS="$(
            printf '%s\n' "$STAGED_GO" | while IFS= read -r f; do
                [ -n "$f" ] || continue
                printf './%s\n' "$(dirname "$f")"
            done | sort -u
        )"
        if [ -n "$PKGS" ]; then
            # Optional Go build tags (.moai/config/build-tags, first non-comment
            # non-blank line) so projects requiring non-default tags (e.g. goolm)
            # are vetted under them.
            BT_TAGS=""
            if [ -f .moai/config/build-tags ]; then
                _bt_line="$(sed -e 's/#.*//' .moai/config/build-tags | awk 'NF{print; exit}')" || true
                [ -n "$_bt_line" ] && BT_TAGS="-tags=$_bt_line"
            fi
            # shellcheck disable=SC2086
            if ! go vet $BT_TAGS $PKGS >/dev/null 2>&1; then
                printf '\n[pre-commit] FAILED: go vet reported issues in the staged packages.\n' >&2
                printf '[pre-commit] Hint: run go vet on the affected packages, fix, then re-commit.\n' >&2
                printf '[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit\n' >&2
                exit 1
            fi
        fi
    fi
fi

# --- Heavy gate: vet + lint + test via 'moai gate' (16-language toolchain detection) ---
# Runs in the user's shell, outside Claude Code's 5s PreToolUse hook budget.
# Skipped when moai is not on PATH so non-moai downstream projects pass silently.
if command -v moai >/dev/null 2>&1; then
    if ! moai gate; then
        printf '\n[pre-commit] FAILED: moai gate reported errors above.\n' >&2
        printf '[pre-commit] Hint: address the reported issues, then re-commit.\n' >&2
        printf '[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit\n' >&2
        exit 1
    fi
fi

exit 0
`

// PreCommitInstaller installs the MoAI-ADK pre-commit git hook. It mirrors
// PrePushInstaller's marker-based semantics for the commit tier.
type PreCommitInstaller struct {
	// repoRoot is the root of the git repository.
	repoRoot string

	// content is the hook body this installer writes — the "incoming" operand
	// of the attribution. Defaults to preCommitHookContent; overridden in tests
	// to construct a version bump without editing the shipped constant.
	content string

	// lastAttribution is the classification of the most recent install run
	// against an existing marker-bearing hook, or nil when the run classified
	// nothing (no existing hook, a non-MoAI hook, or skip=true).
	lastAttribution *hookAttribution

	// lastProvenanceErr holds a provenance-write failure from the most recent
	// run. The write happens after a hook write that already succeeded, so it
	// must not fail the caller (REQ-PCP-010); it is recorded rather than
	// discarded so M2 can warn about it.
	lastProvenanceErr error

	// lastBackupPath is the backup copy written for a user-modified hook in the
	// most recent run, or "" when no backup was taken. The wrapper reads it to
	// emit the disclosure notice (REQ-PCP-004).
	lastBackupPath string

	// now supplies the timestamp for backup names. A seam so tests can force
	// the same stamp twice and construct REQ-PCP-009's occupied-path Given.
	now func() time.Time
}

// NewPreCommitInstaller creates a PreCommitInstaller for the given repository root.
func NewPreCommitInstaller(repoRoot string) *PreCommitInstaller {
	return &PreCommitInstaller{repoRoot: repoRoot, content: preCommitHookContent, now: time.Now}
}

// InstallPreCommitHook writes the pre-commit hook to .git/hooks/pre-commit with
// mode 0755.
//
// Behaviour (mirror of InstallPrePushHook):
//   - skip=true: no-op, returns nil.
//   - File does not exist: creates it.
//   - File exists with MoAI-ADK marker (first 3 lines): overwrites safely.
//   - File exists WITHOUT marker: returns ErrUserHookExists without modifying the file.
//
// The installer is config-agnostic: it never reads the
// git_strategy.<mode>.hooks.pre_commit field. Materializing that field's
// declared intent is achieved by making a real hook exist and run the fast
// subset at commit time; a runtime severity dial is a separate follow-up.
func (p *PreCommitInstaller) InstallPreCommitHook(skip bool) error {
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

	hookPath := filepath.Join(hookDir, "pre-commit")

	// Check if the file already exists.
	if _, err := os.Stat(hookPath); err == nil {
		// File exists — inspect first 3 lines for the MoAI pre-commit marker.
		hasMarker, err := fileHasMarker(hookPath, moaiPreCommitMarker)
		if err != nil {
			return fmt.Errorf("read existing hook: %w", err)
		}
		if !hasMarker {
			return ErrUserHookExists
		}
		// MoAI hook found — classify it before overwriting. The verdict is
		// recorded, not acted on: M1 decides, M2 hangs the backup and the
		// notice off the decision.
		installed, err := os.ReadFile(hookPath)
		if err != nil {
			return fmt.Errorf("read existing hook: %w", err)
		}
		attribution := classifyHook(installed, readHookProvenance(hookDir, preCommitTier), []byte(p.content))
		p.lastAttribution = &attribution

		// REQ-PCP-003: a user-modified hook is copied aside BEFORE the
		// replacement is written, so the patch is recoverable the moment the
		// loss-bearing overwrite happens.
		if attribution.Class == hookUserModified {
			backupPath, err := backupHook(hookDir, installed, p.now(), preCommitTier)
			if err != nil {
				// REQ-PCP-010 sub-case (a): the backup sits before the hook
				// write, so a failure here means the patch cannot be made
				// recoverable — the hook is left untouched and the caller is
				// not failed; the wrapper turns this into a warning. Overwriting
				// anyway would destroy the patch with no recoverable copy,
				// the exact outcome REQ-PCP-006 forbids.
				return fmt.Errorf("%w: %w", errPreCommitBackupFailed, err)
			}
			p.lastBackupPath = backupPath
		}
	}

	if err := os.WriteFile(hookPath, []byte(p.content), 0o755); err != nil {
		return fmt.Errorf("write pre-commit hook: %w", err)
	}

	// Record what was just written, so the next run can attribute a difference
	// (REQ-PCP-001). A failure here follows a hook write that already
	// succeeded, so it must not fail the caller (REQ-PCP-010): the missing
	// record self-heals on the next run, which finds installed == incoming and
	// re-stamps.
	p.lastProvenanceErr = writeHookProvenance(hookDir, p.content, preCommitTier)

	return nil
}

// installPreCommitHookOptional installs the pre-commit hook into projectRoot's
// .git/hooks/ unless skip is true. Friendly, non-fatal: prints progress to out,
// warnings to warn (a writer distinct from out — both callers bind it to the
// command's stderr, REQ-PCP-004), returns nothing. Used by `moai init` and
// `moai update` to install the hook consistently alongside the pre-push hook.
//
// If a non-MoAI user hook is present, this function preserves it and prints a
// note. Other errors are reported as warnings; project init/update is never
// blocked by hook installation failures.
func installPreCommitHookOptional(projectRoot string, skip bool, out, warn io.Writer) {
	installer := NewPreCommitInstaller(projectRoot)
	err := installer.InstallPreCommitHook(skip)
	switch {
	case err == nil:
		if !skip {
			_, _ = fmt.Fprintln(out, "  Pre-commit hook installed (.git/hooks/pre-commit)")
			if installer.lastBackupPath != "" {
				// REQ-PCP-004: the notice carries exactly two elements — the
				// backup path and the fact of the replacement — on the warning
				// writer, never the progress writer (which `moai update` binds
				// to stdout, where a redirected run swallows a data-loss
				// notice whole). It names nothing else; in particular it does
				// not name pre-commit.local, a facility this SPEC does not ship.
				_, _ = fmt.Fprintf(warn, "  Warning: user-modified pre-commit hook was replaced; previous hook backed up at %s\n", installer.lastBackupPath)
			}
			if installer.lastProvenanceErr != nil {
				// REQ-PCP-010 sub-case (b): the hook write already succeeded,
				// so the replacement stands; the missing record self-heals on
				// the next run (REQ-PCP-005).
				_, _ = fmt.Fprintf(warn, "  Warning: pre-commit provenance record not written: %v\n", installer.lastProvenanceErr)
			}
		}
	case errors.Is(err, ErrUserHookExists):
		_, _ = fmt.Fprintln(out, "  Note: existing pre-commit hook preserved (no MoAI-ADK marker found)")
	case errors.Is(err, errPreCommitBackupFailed):
		// REQ-PCP-010 sub-case (a): no backup could be written, so the hook was
		// left exactly as found — nothing was installed and nothing was lost.
		_, _ = fmt.Fprintf(warn, "  Warning: pre-commit hook left unchanged (backup could not be written): %v\n", err)
	default:
		if installer.lastBackupPath != "" {
			// The backup succeeded but the replacement write failed, so no
			// replacement notice (the hook was not replaced). Name the orphan
			// backup so the artifact beside the unchanged hook is explained
			// rather than mysterious (acceptance.md §D.1 edge table).
			_, _ = fmt.Fprintf(warn, "  Warning: pre-commit hook install failed: %v (previous hook backed up at %s)\n", err, installer.lastBackupPath)
		} else {
			_, _ = fmt.Fprintf(out, "  Warning: pre-commit hook install failed: %v\n", err)
		}
	}
}
