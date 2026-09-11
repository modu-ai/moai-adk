#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SCRIPT="$ROOT/.claude/hooks/moai/trace-ledger.sh"
RULE="$ROOT/.claude/rules/moai/workflow/trace-ledger-contract.md"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
LEDGER="$TMP_DIR/workflow-trace.jsonl"

grep -Fq '.claude/skills/moai/workflows/sync/**' "$RULE"
while IFS= read -r workflow_file; do
  grep -Fq '.claude/hooks/moai/trace-ledger.sh record' "$workflow_file"
  if grep -Fq 'Emits one line per Phase entry/exit' "$workflow_file"; then
    echo "stale prose trace claim: $workflow_file" >&2
    exit 1
  fi
done < <(rg -l 'TRACE PROBE: activation hint only' "$ROOT/.claude/skills/moai/workflows")

bash "$SCRIPT" record "$LEDGER" run.phase0 stage_start root sha256:abc S-cold miss none 0
bash "$SCRIPT" record "$LEDGER" run.tool tool_end root sha256:abc S-cold miss none 17
bash "$SCRIPT" record "$LEDGER" run.phase0 stage_end root sha256:abc S-cold miss timeout 31

[[ "$(wc -l < "$LEDGER" | tr -d ' ')" == "3" ]]
jq -e -s '
  length == 3 and
  .[0].event == "stage_start" and
  .[0].cohort == "S-cold" and
  .[1].event == "tool_end" and
  .[1].duration_ms == 17 and
  .[2].retry_cause == "timeout"
' "$LEDGER" >/dev/null

SUMMARY="$(bash "$SCRIPT" summary "$LEDGER")"
jq -e '
  .records == 3 and
  .calls == 1 and
  .p50_duration_ms == 17 and
  .p95_duration_ms == 17 and
  .cache_misses == 3 and
  .retries == 1 and
  .cohorts == ["S-cold"]
' <<<"$SUMMARY" >/dev/null

if bash "$SCRIPT" record "$LEDGER" 'bad stage' stage_start root sha256:abc S-cold miss none 0 >/dev/null 2>&1; then
  echo 'unsafe stage must be rejected' >&2
  exit 1
fi

printf 'PASS: workflow trace ledger records identity, cache, retry, duration, and p50/p95 evidence\n'
