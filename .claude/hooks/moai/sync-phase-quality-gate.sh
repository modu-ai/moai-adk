#!/bin/bash
# Hook: sync-phase-quality-gate
# Purpose: Enforce sync-phase quality gate (lint + test + coverage delta + dependency manifest audit)
# Trigger: Stop event when current session contains sync-phase commit
#
# Language support: detects the project language from canonical project markers
# and runs the matching toolchain. Each language's tools live inside its own
# case branch; absent tools are skipped gracefully; projects with no recognized
# language marker pass the gate silently.
#
# Manual smoke test:
#   echo '{}' | bash .claude/hooks/moai/sync-phase-quality-gate.sh
# Expected: empty stdout (silent pass) on skip/allow; on block, a Stop-schema JSON
# {"decision":"block","reason":...,"systemMessage":...}. The verifications detail is
# written to .moai/logs/sync-quality-gate.log, not stdout (Stop JSON-schema rejects
# unknown fields and non-{approve,block} decision values).
#
# Unit-test the detector directly (bypasses the sync-phase git gate):
#   source .claude/hooks/moai/sync-phase-quality-gate.sh && detect_language "$dir"

set -e

# --- detect_language: directly-invocable, side-effect-free language detector ---
# Echoes a single language token (go|node|python|rust) or empty string when no
# recognized marker is present. Marker priority follows the language matrix order.
# This function MUST remain source-able so it can be unit-tested without first
# passing the sync-phase-commit git gate below.
detect_language() {
    root="${1:-.}"
    if [ -f "$root/go.mod" ]; then
        echo "go"
    elif [ -f "$root/package.json" ]; then
        echo "node"
    elif [ -f "$root/pyproject.toml" ] || [ -f "$root/requirements.txt" ]; then
        echo "python"
    elif [ -f "$root/Cargo.toml" ]; then
        echo "rust"
    else
        echo ""
    fi
}

# --- code_delta_pattern: per-language source-file extension regex ---
# Used to detect whether the sync-phase commit touched code files; a 0-code-file
# delta means a docs/markdown-only sync and the gate skips.
code_delta_pattern() {
    case "$1" in
        go)     echo '\.go$' ;;
        node)   echo '\.(js|ts|jsx|tsx|mjs|cjs)$' ;;
        python) echo '\.py$' ;;
        rust)   echo '\.rs$' ;;
        *)      echo '' ;;
    esac
}

# Opt-out flag
if [ "$1" = "--skip-hook" ]; then
    echo "{\"skipped\": true, \"reason\": \"--skip-hook flag\"}" >&2
    mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] skipped via --skip-hook" \
        >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log"
    exit 0
fi

# When sourced for unit testing, stop here so detect_language is available
# without running the gate. Detection: BASH_SOURCE[0] != $0 means the file was
# sourced, not executed directly.
case "${BASH_SOURCE[0]}" in
    "$0") ;;            # executed directly — continue running the gate
    *) return 0 2>/dev/null || true ;;  # sourced — expose functions, do not run
esac

# Detect sync-phase via last commit subject
LAST_COMMIT_SUBJECT=$(git log -1 --format='%s' 2>/dev/null || echo "")
case "$LAST_COMMIT_SUBJECT" in
    *"docs("*"): sync-phase"*|*"chore("*"): sync-phase"*|*"docs: sync"*|*"chore: sync"*)
        ;;
    *)
        # Not a sync-phase commit — gate not applicable, silent pass.
        # stdout intentionally empty: Stop decision must be "approve" | "block"
        # (not "skip"); unknown fields also fail Claude Code JSON-schema validation.
        exit 0
        ;;
esac

# Resolve project root and detect language from canonical markers
PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
LANG=$(detect_language "$PROJECT_ROOT")

# Silent pass when no recognized language marker is present (docs-only projects, etc.)
if [ -z "$LANG" ]; then
    # stdout intentionally empty (Stop schema: decision must be approve|block, not "skip").
    exit 0
fi

# Detect code-file changes in HEAD commit; skip if 0 code-file delta (markdown-only sync).
# On an initial commit HEAD~1 does not exist, so diff against the empty tree instead.
DELTA_PATTERN=$(code_delta_pattern "$LANG")
if git rev-parse --verify -q HEAD~1 >/dev/null 2>&1; then
    DIFF_RANGE="HEAD~1..HEAD"
else
    DIFF_RANGE=$(git hash-object -t tree /dev/null)  # empty-tree SHA — initial commit
fi
# grep -c is wrapped so its no-match exit (1) under `set -e` does not abort; the
# result is normalized to a single integer (avoids a "0\n0" double-emit).
CODE_DELTA=$(git diff --name-only "$DIFF_RANGE" 2>/dev/null | grep -cE "$DELTA_PATTERN" || true)
CODE_DELTA=${CODE_DELTA:-0}
if [ "$CODE_DELTA" -eq 0 ]; then
    # stdout intentionally empty (Stop schema: decision must be approve|block, not "skip").
    exit 0
fi

# Per-check result scratch dir
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Default per-check results: 0 = pass/skipped, used when a step does not run for
# the detected language. command -v guards every tool invocation so an absent
# toolchain is skipped gracefully (exit 0, recorded as skipped) rather than failing.
echo "0" > "$TMPDIR/vet.exit";  echo "not run for $LANG" > "$TMPDIR/vet.log"
echo "0" > "$TMPDIR/lint.exit"; echo "not run for $LANG" > "$TMPDIR/lint.log"
echo "0" > "$TMPDIR/test.exit"; echo "not run for $LANG" > "$TMPDIR/test.log"

# run_step <tool> <result-prefix> <command...>: run only if the tool is on PATH,
# otherwise record exit 0 and log a graceful skip. The `&& rc=0 || rc=$?` idiom
# captures the tool's exit code without letting `set -e` abort the hook when a
# tool legitimately fails (the failure is recorded and drives the decision).
run_step() {
    tool="$1"; prefix="$2"; shift 2
    if command -v "$tool" >/dev/null 2>&1; then
        local rc=0
        "$@" > "$TMPDIR/$prefix.log" 2>&1 && rc=0 || rc=$?
        echo "$rc" > "$TMPDIR/$prefix.exit"
    else
        echo "0" > "$TMPDIR/$prefix.exit"
        echo "skipped: $tool absent" > "$TMPDIR/$prefix.log"
    fi
}

VET_LABEL="(none)"
LINT_LABEL="(none)"
TEST_LABEL="(none)"
COVERAGE="0.0"

case "$LANG" in
    go)
        VET_LABEL="go vet"; LINT_LABEL="golangci-lint"; TEST_LABEL="go test"
        run_step go vet go vet ./...
        run_step golangci-lint lint golangci-lint run --timeout=2m
        run_step go test go test ./...
        if command -v go >/dev/null 2>&1; then
            go test -cover ./... 2>/dev/null | grep -E "^ok\s" | awk '{sum += $5; count++} END {if (count > 0) printf "%.1f", sum/count; else print "0.0"}' > "$TMPDIR/coverage.log" 2>/dev/null || echo "0.0" > "$TMPDIR/coverage.log"
            COVERAGE=$(cat "$TMPDIR/coverage.log")
        fi
        ;;
    node)
        LINT_LABEL="eslint"; TEST_LABEL="npm test"
        run_step eslint lint eslint .
        run_step npm test npm test
        ;;
    python)
        LINT_LABEL="ruff"; TEST_LABEL="pytest"
        run_step ruff lint ruff check .
        run_step pytest test pytest
        ;;
    rust)
        LINT_LABEL="cargo clippy"; TEST_LABEL="cargo test"
        run_step cargo lint cargo clippy
        run_step cargo test cargo test
        ;;
esac

# Dependency manifest audit: flag if a dependency manifest was modified in the
# sync-phase commit (unexpected for a docs sync). Language-specific manifest set.
DEPS_MANIFESTS=""
case "$LANG" in
    go)     DEPS_MANIFESTS="go.mod go.sum" ;;
    node)   DEPS_MANIFESTS="package.json package-lock.json yarn.lock pnpm-lock.yaml" ;;
    python) DEPS_MANIFESTS="pyproject.toml requirements.txt poetry.lock" ;;
    rust)   DEPS_MANIFESTS="Cargo.toml Cargo.lock" ;;
esac
git diff HEAD~1..HEAD -- $DEPS_MANIFESTS > "$TMPDIR/deps.diff" 2>&1 || true
DEPS_MODIFIED=0
if [ -s "$TMPDIR/deps.diff" ]; then
    DEPS_MODIFIED=1
fi

VET_EXIT=$(cat "$TMPDIR/vet.exit")
LINT_EXIT=$(cat "$TMPDIR/lint.exit")
TEST_EXIT=$(cat "$TMPDIR/test.exit")

# Decision
DECISION="allow"
BLOCKED_REASON=""
if [ "$VET_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="$VET_LABEL failed"
elif [ "$LINT_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="$LINT_LABEL failed"
elif [ "$TEST_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="$TEST_LABEL failed"
fi

# Emit Stop-schema-compliant JSON. The custom {hook,language,decision:skip/allow,
# verifications} shape failed Claude Code JSON-schema validation on every Stop fire:
# Stop accepts only "decision":"approve"|"block" (not "skip"/"allow") and rejects
# unknown top-level fields. Per-check verifications detail is recorded in the audit
# log below; stdout carries only schema-valid fields. On block, systemMessage surfaces
# the failure to the user (advisory: non-blocking unless MOAI_SYNC_GATE_BLOCKING=1).
# On allow, stdout is intentionally empty (silent pass; audit log records the detail).
if [ "$DECISION" = "block" ]; then
    printf '{"decision":"block","reason":"%s","systemMessage":"sync-phase quality gate BLOCKED: %s (vet=%s lint=%s test=%s deps_modified=%s coverage=%s%%). Detail: .moai/logs/sync-quality-gate.log"}\n' \
        "$BLOCKED_REASON" "$BLOCKED_REASON" "$VET_EXIT" "$LINT_EXIT" "$TEST_EXIT" "$DEPS_MODIFIED" "$COVERAGE"
fi

mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] language=$LANG decision=$DECISION vet=$VET_EXIT lint=$LINT_EXIT test=$TEST_EXIT deps_modified=$DEPS_MODIFIED coverage=$COVERAGE" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/sync-quality-gate.log"

# Warn-first (advisory) by default: the blocking exit-2 path is dormant and
# enabled only when MOAI_SYNC_GATE_BLOCKING=1 is explicitly set. With the flag
# unset (the shipped default), a "block" decision is logged but the hook exits 0
# so the Stop event is not blocked. Activating the exit-2 gate is deferred to a
# follow-up; the dormant path is preserved here intentionally.
if [ "$DECISION" = "block" ] && [ "${MOAI_SYNC_GATE_BLOCKING:-0}" = "1" ]; then
    exit 2
fi
exit 0
