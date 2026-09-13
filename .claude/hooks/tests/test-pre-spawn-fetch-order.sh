#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
RULE="$ROOT/.claude/rules/moai/core/agent-common-protocol.md"
SECTION="$(sed -n '/### Pre-Spawn Sync Check (Multi-Session Race Mitigation)/,/^Interpretation matrix (git divergence)/p' "$RULE")"
LANE_A="$(printf '%s\n' "$SECTION" | sed -n '/# Lane A — ordered/,/moai session list/p')"

fetch_line="$(printf '%s\n' "$LANE_A" | grep -n 'git fetch origin main 2>&1' | head -1 | cut -d: -f1)"
rev_line="$(printf '%s\n' "$LANE_A" | grep -n '^git rev-list --count --left-right origin/main...HEAD' | head -1 | cut -d: -f1)"

test -n "$fetch_line"
test -n "$rev_line"
test "$fetch_line" -lt "$rev_line"
grep -Fq 'MUST finish and its exit status be observed before' <<<"$SECTION"
grep -Fq 'may run concurrently with Lane A' <<<"$SECTION"
grep -Fq 'fetch_status=$?' <<<"$SECTION"
grep -Fq 'pre-spawn sync blocked: fetch origin/main failed' <<<"$SECTION"

printf 'PASS: pre-spawn fetch completion gates the divergence read while session discovery remains independent\n'
