package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// moaiPreCommitMarker is the identifier written near the top of the pre-commit
// hook file. Its presence signals that the hook was installed by MoAI-ADK and
// can be safely overwritten on subsequent installs.
const moaiPreCommitMarker = "# MoAI-ADK pre-commit hook"

// moaiPreCommitProvenanceName is the sidecar recording the SHA-256 of the hook
// content this tool last wrote. It lives beside the hook in .git/hooks/, so it
// shares the hook's lifetime exactly: wiping .git/hooks/ takes both.
//
// The record is the third operand of the attribution in REQ-PCP-014. Without
// it only "installed" and "incoming" exist, and those two cannot separate "the
// user changed it" from "we changed it" — a routine version bump then reads as
// a user patch for every user on every release.
const moaiPreCommitProvenanceName = ".moai-pre-commit.sha256"

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

// preCommitClass is the attribution verdict for an existing marker-bearing hook.
type preCommitClass int

const (
	// preCommitUnmodified: the hook is as MoAI last wrote it. Any difference
	// against the incoming content is an upstream version bump, so the
	// overwrite is quiet (REQ-PCP-002).
	preCommitUnmodified preCommitClass = iota
	// preCommitUserModified: the hook was edited after MoAI wrote it. This is
	// the loss-bearing case; M2 hangs the backup and the notice off it.
	preCommitUserModified
)

func (c preCommitClass) String() string {
	if c == preCommitUserModified {
		return "user-modified"
	}
	return "unmodified"
}

// preCommitBasis names which operands produced the verdict — the third label of
// the spec's three-way classification, kept separate from the verdict because
// "undecidable-legacy" describes how the answer was reached, not what it was.
type preCommitBasis int

const (
	// preCommitBasisRecord: a usable provenance record existed, so the verdict
	// comes from installed-vs-recorded — the three-way comparison.
	preCommitBasisRecord preCommitBasis = iota
	// preCommitBasisUndecidableLegacy: no usable record existed, so attribution
	// was impossible and the verdict falls back to installed-vs-incoming, with
	// any difference read as a user edit (REQ-PCP-005). Deliberately the noisy
	// direction: a hand-patched legacy hook is the most likely thing to be
	// found without a record.
	preCommitBasisUndecidableLegacy
)

func (b preCommitBasis) String() string {
	if b == preCommitBasisUndecidableLegacy {
		return "undecidable-legacy"
	}
	return "record"
}

// preCommitAttribution is one classification of an existing marker-bearing hook.
type preCommitAttribution struct {
	Class preCommitClass
	Basis preCommitBasis
}

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
	lastAttribution *preCommitAttribution

	// lastProvenanceErr holds a provenance-write failure from the most recent
	// run. The write happens after a hook write that already succeeded, so it
	// must not fail the caller (REQ-PCP-010); it is recorded rather than
	// discarded so M2 can warn about it.
	lastProvenanceErr error
}

// NewPreCommitInstaller creates a PreCommitInstaller for the given repository root.
func NewPreCommitInstaller(repoRoot string) *PreCommitInstaller {
	return &PreCommitInstaller{repoRoot: repoRoot, content: preCommitHookContent}
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
		attribution := classifyPreCommitHook(installed, readPreCommitProvenance(hookDir), []byte(p.content))
		p.lastAttribution = &attribution
	}

	if err := os.WriteFile(hookPath, []byte(p.content), 0o755); err != nil {
		return fmt.Errorf("write pre-commit hook: %w", err)
	}

	// Record what was just written, so the next run can attribute a difference
	// (REQ-PCP-001). A failure here follows a hook write that already
	// succeeded, so it must not fail the caller (REQ-PCP-010): the missing
	// record self-heals on the next run, which finds installed == incoming and
	// re-stamps.
	p.lastProvenanceErr = writePreCommitProvenance(hookDir, p.content)

	return nil
}

// classifyPreCommitHook decides whether an existing marker-bearing hook was
// edited by the user, from three operands: the installed bytes, the digest this
// tool last recorded writing (empty when absent or unusable), and the incoming
// bytes.
//
// A two-way comparison of installed against incoming cannot make this call: an
// upstream version bump produces the same signal as a user patch, so a two-way
// design warns every user on every release that touches the hook (REQ-PCP-014).
func classifyPreCommitHook(installed []byte, recordedDigest string, incoming []byte) preCommitAttribution {
	if recordedDigest == "" {
		// No usable record: attribution is impossible, so fall back to
		// installed-vs-incoming and read any difference as a user edit.
		if bytes.Equal(installed, incoming) {
			return preCommitAttribution{Class: preCommitUnmodified, Basis: preCommitBasisUndecidableLegacy}
		}
		return preCommitAttribution{Class: preCommitUserModified, Basis: preCommitBasisUndecidableLegacy}
	}

	if digestOfBytes(installed) == recordedDigest {
		return preCommitAttribution{Class: preCommitUnmodified, Basis: preCommitBasisRecord}
	}
	return preCommitAttribution{Class: preCommitUserModified, Basis: preCommitBasisRecord}
}

// readPreCommitProvenance returns the recorded digest, or "" when no usable
// record exists. A missing, unreadable or malformed record is treated as
// absent, which routes the caller to the deliberately noisy legacy path
// (REQ-PCP-005).
func readPreCommitProvenance(hookDir string) string {
	raw, err := os.ReadFile(filepath.Join(hookDir, moaiPreCommitProvenanceName))
	if err != nil {
		return ""
	}
	digest := strings.TrimSpace(string(raw))
	if len(digest) != hex.EncodedLen(sha256.Size) {
		return ""
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return ""
	}
	return digest
}

// writePreCommitProvenance records the digest of the content just written.
func writePreCommitProvenance(hookDir, content string) error {
	path := filepath.Join(hookDir, moaiPreCommitProvenanceName)
	if err := os.WriteFile(path, []byte(digestOfBytes([]byte(content))+"\n"), 0o644); err != nil {
		return fmt.Errorf("write pre-commit provenance record: %w", err)
	}
	return nil
}

func digestOfBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
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
