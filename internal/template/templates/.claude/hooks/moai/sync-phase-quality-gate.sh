#!/bin/bash
# Hook: sync-phase-quality-gate
# Purpose: Fast sync-phase quality gate (compile/vet + dependency manifest audit)
# Trigger: Stop event when the current session's HEAD is a sync-phase commit
#
# Scope: the hook runs ONLY fast structural checks (compile/vet) that finish well
# within the Stop timeout. Heavy lint (golangci-lint) and the full test suite are
# deliberately NOT run here — they cannot finish within a turn-end Stop timeout and
# belong in CI. Each language's fast check lives inside its own case branch; absent
# tools are skipped gracefully; projects with no recognized language marker pass
# the gate silently.
#
# Behavior: advisory (non-blocking) by DEFAULT. A failing check emits only
# {"systemMessage": ...} — a warning that does NOT block the turn. The blocking
# path (exit 2 + {"decision":"block"}) is dormant and enabled only when
# MOAI_SYNC_GATE_BLOCKING=1 is explicitly set. This split matters because, per
# Claude Code Stop-hook semantics, stdout carrying {"decision":"block"} blocks the
# turn REGARDLESS of exit code — so an advisory run must never emit that field.
#
# Once-per-commit: a given sync commit is gated at most ONCE. The gated HEAD SHA is
# recorded in .moai/state/sync-quality-gate.last and the hook short-circuits on any
# later turn whose HEAD is unchanged, so the gate does not re-run every turn-end.
#
# Manual smoke test:
#   echo '{}' | bash .claude/hooks/moai/sync-phase-quality-gate.sh
# Expected: empty stdout (silent pass) on skip/allow; on an advisory warning a Stop
# JSON {"systemMessage":...}; on a blocking failure (MOAI_SYNC_GATE_BLOCKING=1) a
# Stop JSON {"decision":"block","reason":...,"systemMessage":...}. The per-check
# detail is written to .moai/logs/sync-quality-gate.log, not stdout (Stop JSON-schema
# rejects unknown fields and non-{approve,block} decision values).
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

# Once-per-commit sentinel: gate a given sync commit at most ONCE. Without this the
# Stop hook re-fires on every subsequent turn-end while HEAD is still the sync commit
# (the last-commit-subject trigger stays matched until a newer non-sync commit lands),
# re-running the toolchain each turn. Record the gated HEAD SHA and short-circuit when
# it is unchanged. The SHA is recorded BEFORE the checks run, so a slow/killed run
# still counts as gated and cannot re-trigger a per-turn re-run.
HEAD_SHA=$(git rev-parse HEAD 2>/dev/null || echo "")
STATE_DIR="${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state"
SENTINEL_FILE="$STATE_DIR/sync-quality-gate.last"
if [ -n "$HEAD_SHA" ] && [ -f "$SENTINEL_FILE" ] && [ "$(cat "$SENTINEL_FILE" 2>/dev/null)" = "$HEAD_SHA" ]; then
    # This commit was already gated in a prior turn — silent pass, no re-run.
    exit 0
fi
if [ -n "$HEAD_SHA" ]; then
    mkdir -p "$STATE_DIR"
    echo "$HEAD_SHA" > "$SENTINEL_FILE"
fi

# Per-check result scratch dir
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Default per-check results: 0 = pass/skipped, used when a step does not run for the
# detected language. command -v guards every tool invocation so an absent toolchain
# is skipped gracefully (exit 0, recorded as skipped) rather than failing.
echo "0" > "$TMPDIR/c1.exit"; echo "not run for $LANG" > "$TMPDIR/c1.log"
echo "0" > "$TMPDIR/c2.exit"; echo "not run for $LANG" > "$TMPDIR/c2.log"

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

# Fast structural checks only. Two slots per language: c1 (vet/lint) + c2 (build).
# Heavy lint (golangci-lint) and the full test suite are intentionally NOT run here —
# they cannot finish within the Stop timeout and belong in CI.
C1_LABEL="(none)"
C2_LABEL="(none)"

case "$LANG" in
    go)
        C1_LABEL="go vet"; C2_LABEL="go build"
        run_step go c1 go vet ./...
        run_step go c2 go build ./...
        ;;
    node)
        C1_LABEL="eslint"
        run_step eslint c1 eslint .
        ;;
    python)
        C1_LABEL="ruff"
        run_step ruff c1 ruff check .
        ;;
    rust)
        C1_LABEL="cargo check"
        run_step cargo c1 cargo check
        ;;
esac

# Dependency manifest audit: flag if a dependency manifest was modified in the
# sync-phase commit (unexpected for a docs sync). Informational only — it does NOT
# drive the block decision. Language-specific manifest set.
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

C1_EXIT=$(cat "$TMPDIR/c1.exit")
C2_EXIT=$(cat "$TMPDIR/c2.exit")

# Decision
DECISION="allow"
BLOCKED_REASON=""
if [ "$C1_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="$C1_LABEL failed"
elif [ "$C2_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="$C2_LABEL failed"
fi

# Resolve the mode once (set -e safe) for both stdout and the audit log.
if [ "${MOAI_SYNC_GATE_BLOCKING:-0}" = "1" ]; then MODE="blocking"; else MODE="advisory"; fi

# Emit a Stop-schema-compliant response.
#
# Advisory (default): a failing check emits ONLY {"systemMessage": ...} — a
# non-blocking warning. Per Claude Code Stop-hook semantics, stdout carrying
# {"decision":"block"} blocks the turn REGARDLESS of exit code, so the advisory
# path MUST NOT emit a "decision" field.
#
# Blocking (opt-in, MOAI_SYNC_GATE_BLOCKING=1): a failing check emits
# {"decision":"block", ...} and exits 2 below — this blocks the turn.
#
# On allow, stdout is intentionally empty (silent pass); the audit log records detail.
if [ "$DECISION" = "block" ]; then
    if [ "$MODE" = "blocking" ]; then
        printf '{"decision":"block","reason":"%s","systemMessage":"sync-phase quality gate BLOCKED: %s (%s=%s %s=%s deps_modified=%s). Detail: .moai/logs/sync-quality-gate.log"}\n' \
            "$BLOCKED_REASON" "$BLOCKED_REASON" "$C1_LABEL" "$C1_EXIT" "$C2_LABEL" "$C2_EXIT" "$DEPS_MODIFIED"
    else
        printf '{"systemMessage":"sync-phase quality gate WARNING (advisory, not blocking): %s (%s=%s %s=%s deps_modified=%s). Heavy lint/tests run in CI. Detail: .moai/logs/sync-quality-gate.log"}\n' \
            "$BLOCKED_REASON" "$C1_LABEL" "$C1_EXIT" "$C2_LABEL" "$C2_EXIT" "$DEPS_MODIFIED"
    fi
fi

mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] language=$LANG mode=$MODE decision=$DECISION $C1_LABEL=$C1_EXIT $C2_LABEL=$C2_EXIT deps_modified=$DEPS_MODIFIED head=$HEAD_SHA" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/sync-quality-gate.log"

# Blocking exit-2 path is dormant and enabled only when MOAI_SYNC_GATE_BLOCKING=1.
# With the flag unset (the shipped default), a "block" decision surfaces only as an
# advisory systemMessage above and the hook exits 0 so the Stop event is not blocked.
if [ "$DECISION" = "block" ] && [ "$MODE" = "blocking" ]; then
    exit 2
fi
exit 0
