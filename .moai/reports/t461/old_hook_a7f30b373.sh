#!/bin/sh
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
