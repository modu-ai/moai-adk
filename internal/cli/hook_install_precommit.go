package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// moaiPreCommitMarker is the identifier written near the top of the pre-commit
// hook file. Its presence signals that the hook was installed by MoAI-ADK and
// can be safely overwritten on subsequent installs.
const moaiPreCommitMarker = "# MoAI-ADK pre-commit hook"

// preCommitHookContent is the canonical content of the fast-subset pre-commit
// hook. It runs gofmt -l + go vet on staged Go files only (the fast commit tier);
// the full local CI mirror stays in the pre-push hook (the push tier).
//
// MUST stay byte-identical with internal/template/templates/.git_hooks/pre-commit.
// TestPreCommitTemplateMatchesConstant enforces this; do not edit one without the other.
const preCommitHookContent = `#!/bin/sh
# MoAI-ADK pre-commit hook — fast subset (gofmt -l + go vet on staged Go files)
# Bypass via: SKIP_MOAI_PRECOMMIT=1 git commit
# Full local CI (fmt + lint + build + suite) is the pre-push hook's job — not duplicated here.
set -eu

if [ "${SKIP_MOAI_PRECOMMIT:-0}" = "1" ]; then
    printf '[pre-commit] SKIP_MOAI_PRECOMMIT=1 -- bypass requested\n' >&2
    exit 0
fi

# Staged Go files (Added / Copied / Modified; deletions excluded via ACM).
STAGED_GO="$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)"
[ -z "$STAGED_GO" ] && exit 0

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
        # shellcheck disable=SC2086
        if ! go vet $PKGS >/dev/null 2>&1; then
            printf '\n[pre-commit] FAILED: go vet reported issues in the staged packages.\n' >&2
            printf '[pre-commit] Hint: run go vet on the affected packages, fix, then re-commit.\n' >&2
            printf '[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit\n' >&2
            exit 1
        fi
    fi
fi

exit 0
`

// PreCommitInstaller installs the MoAI-ADK pre-commit git hook. It mirrors
// PrePushInstaller's marker-based semantics for the commit tier.
type PreCommitInstaller struct {
	// repoRoot is the root of the git repository.
	repoRoot string
}

// NewPreCommitInstaller creates a PreCommitInstaller for the given repository root.
func NewPreCommitInstaller(repoRoot string) *PreCommitInstaller {
	return &PreCommitInstaller{repoRoot: repoRoot}
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
		// MoAI hook found — overwrite.
	}

	if err := os.WriteFile(hookPath, []byte(preCommitHookContent), 0o755); err != nil {
		return fmt.Errorf("write pre-commit hook: %w", err)
	}

	return nil
}

// installPreCommitHookOptional installs the pre-commit hook into projectRoot's
// .git/hooks/ unless skip is true. Friendly, non-fatal: prints status to out,
// returns nothing. Used by `moai init` and `moai update` to install the hook
// consistently alongside the pre-push hook.
//
// If a non-MoAI user hook is present, this function preserves it and prints a
// note. Other errors are reported as warnings; project init/update is never
// blocked by hook installation failures.
func installPreCommitHookOptional(projectRoot string, skip bool, out io.Writer) {
	installer := NewPreCommitInstaller(projectRoot)
	err := installer.InstallPreCommitHook(skip)
	switch {
	case err == nil:
		if !skip {
			_, _ = fmt.Fprintln(out, "  Pre-commit hook installed (.git/hooks/pre-commit)")
		}
	case errors.Is(err, ErrUserHookExists):
		_, _ = fmt.Fprintln(out, "  Note: existing pre-commit hook preserved (no MoAI-ADK marker found)")
	default:
		_, _ = fmt.Fprintf(out, "  Warning: pre-commit hook install failed: %v\n", err)
	}
}
