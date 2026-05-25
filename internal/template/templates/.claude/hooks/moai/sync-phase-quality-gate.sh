#!/bin/bash
# Hook: sync-phase-quality-gate
# Purpose: Enforce sync-phase quality gate (lint + test + coverage delta + dependency manifest audit)
# Trigger: Stop event when current session contains sync-phase commit
# Origin: SPEC-V3R6-AGENT-TEAM-REBUILD-001 M4 (2026-05-25)
# REQs: REQ-ATR-009 (replaces manager-quality spawn in workflows/sync.md);
#       REQ-ATR-014 (dependency manifest audit)
#
# Manual smoke test:
#   echo '{}' | bash .claude/hooks/moai/sync-phase-quality-gate.sh
# Expected: structured JSON with "decision":"skip" or "decision":"allow"/"block"
# and per-check results in "verifications" array.

set -e

# Opt-out flag
if [ "$1" = "--skip-hook" ]; then
    echo "{\"skipped\": true, \"reason\": \"--skip-hook flag\"}" >&2
    mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] skipped via --skip-hook" \
        >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log"
    exit 0
fi

# Detect sync-phase via last commit subject
LAST_COMMIT_SUBJECT=$(git log -1 --format='%s' 2>/dev/null || echo "")
case "$LAST_COMMIT_SUBJECT" in
    *"docs("*"): sync-phase"*|*"chore("*"): sync-phase"*|*"docs: sync"*|*"chore: sync"*)
        ;;
    *)
        # Not a sync-phase commit — skip
        echo "{\"hook\":\"sync-phase-quality-gate\",\"decision\":\"skip\",\"reason\":\"not a sync-phase commit\",\"last_subject\":\"$LAST_COMMIT_SUBJECT\"}"
        exit 0
        ;;
esac

# Detect Go file changes in HEAD commit; skip if 0 .go file delta (markdown-only sync)
GO_DELTA=$(git diff --name-only HEAD~1..HEAD 2>/dev/null | grep -cE '\.go$' || echo "0")
if [ "$GO_DELTA" -eq 0 ]; then
    echo "{\"hook\":\"sync-phase-quality-gate\",\"decision\":\"skip\",\"reason\":\"markdown-only sync (0 .go file delta)\"}"
    exit 0
fi

# Parallel verification batch
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# 1. go vet
(go vet ./... > "$TMPDIR/vet.log" 2>&1; echo $? > "$TMPDIR/vet.exit") &
VET_PID=$!

# 2. golangci-lint (if available)
if command -v golangci-lint >/dev/null 2>&1; then
    (golangci-lint run --timeout=2m > "$TMPDIR/lint.log" 2>&1; echo $? > "$TMPDIR/lint.exit") &
    LINT_PID=$!
else
    echo "0" > "$TMPDIR/lint.exit"
    echo "golangci-lint unavailable; skipped" > "$TMPDIR/lint.log"
    LINT_PID=""
fi

# 3. go test
(go test ./... > "$TMPDIR/test.log" 2>&1; echo $? > "$TMPDIR/test.exit") &
TEST_PID=$!

# 4. Dependency manifest audit (REQ-ATR-014)
(git diff HEAD~1..HEAD -- go.mod go.sum > "$TMPDIR/deps.diff" 2>&1; echo $? > "$TMPDIR/deps.exit") &
DEPS_PID=$!

# Wait for parallel jobs
wait $VET_PID $TEST_PID $DEPS_PID
[ -n "$LINT_PID" ] && wait $LINT_PID || true

VET_EXIT=$(cat "$TMPDIR/vet.exit")
LINT_EXIT=$(cat "$TMPDIR/lint.exit")
TEST_EXIT=$(cat "$TMPDIR/test.exit")
DEPS_EXIT=$(cat "$TMPDIR/deps.exit")

# Dependency audit: flag if go.mod or go.sum modified in sync-phase commit (unexpected)
DEPS_MODIFIED=0
if [ -s "$TMPDIR/deps.diff" ]; then
    DEPS_MODIFIED=1
fi

# Coverage delta: simplified (full coverage diff requires baseline file)
# For M4 baseline implementation, just record current coverage; coverage regression check
# is a deferred enhancement.
COVERAGE_LOG="$TMPDIR/coverage.log"
go test -cover ./... 2>/dev/null | grep -E "^ok\s" | awk '{sum += $5; count++} END {if (count > 0) printf "%.1f", sum/count; else print "0.0"}' > "$COVERAGE_LOG" 2>/dev/null || echo "0.0" > "$COVERAGE_LOG"
COVERAGE=$(cat "$COVERAGE_LOG")

# Decision
DECISION="allow"
BLOCKED_REASON=""
if [ "$VET_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="go vet failed"
elif [ "$LINT_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="golangci-lint failed"
elif [ "$TEST_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="go test failed"
fi

cat <<EOF
{
  "hook": "sync-phase-quality-gate",
  "decision": "$DECISION",
  "blocked_reason": "$BLOCKED_REASON",
  "verifications": [
    {"check": "go vet", "exit": $VET_EXIT},
    {"check": "golangci-lint", "exit": $LINT_EXIT},
    {"check": "go test", "exit": $TEST_EXIT},
    {"check": "dependency manifest audit", "go_mod_or_sum_modified": $DEPS_MODIFIED},
    {"check": "coverage (advisory)", "average_coverage_pct": $COVERAGE}
  ]
}
EOF

mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] decision=$DECISION vet=$VET_EXIT lint=$LINT_EXIT test=$TEST_EXIT deps_modified=$DEPS_MODIFIED coverage=$COVERAGE" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/sync-quality-gate.log"

if [ "$DECISION" = "block" ]; then
    exit 2
fi
exit 0
