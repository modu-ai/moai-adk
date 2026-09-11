#!/usr/bin/env bash
# Record and summarize executable workflow trace evidence.
#
# The workflow comments are activation hints only. This helper is the evidence
# surface: one validated JSON object per append, never a prose trace claim.
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
usage:
  trace-ledger.sh record <ledger> <stage> <event> <parent_id> <input_hash> <cohort> <cache_hit> <retry_cause> <duration_ms>
  trace-ledger.sh summary <ledger>
USAGE
}

fail() {
  printf 'trace-ledger: %s\n' "$1" >&2
  exit 2
}

safe_token() {
  case "$1" in
    ''|*[!A-Za-z0-9_.:/-]*) return 1 ;;
    *) return 0 ;;
  esac
}

record() {
  [[ $# -eq 9 ]] || { usage; exit 2; }
  local ledger=$1 stage=$2 event=$3 parent_id=$4 input_hash=$5 cohort=$6 cache_hit=$7 retry_cause=$8 duration_ms=$9

  [[ -n "$ledger" ]] || fail 'ledger path is required'
  safe_token "$stage" || fail 'stage contains an unsupported character'
  case "$event" in stage_start|stage_end|tool_start|tool_end) ;; *) fail 'event must be stage_start, stage_end, tool_start, or tool_end' ;; esac
  safe_token "$parent_id" || fail 'parent_id contains an unsupported character'
  safe_token "$input_hash" || fail 'input_hash contains an unsupported character'
  safe_token "$cohort" || fail 'cohort contains an unsupported character'
  case "$cache_hit" in hit|miss|unknown) ;; *) fail 'cache_hit must be hit, miss, or unknown' ;; esac
  safe_token "$retry_cause" || fail 'retry_cause contains an unsupported character'
  [[ "$duration_ms" =~ ^[0-9]+$ ]] || fail 'duration_ms must be a non-negative integer'

  local parent_dir
  parent_dir=$(dirname "$ledger")
  umask 077
  mkdir -p "$parent_dir"
  printf '{"schema_version":1,"ts":"%s","stage":"%s","event":"%s","parent_id":"%s","input_hash":"%s","cohort":"%s","cache_hit":"%s","retry_cause":"%s","duration_ms":%s}\n' \
    "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$stage" "$event" "$parent_id" "$input_hash" "$cohort" "$cache_hit" "$retry_cause" "$duration_ms" >> "$ledger"
}

summary() {
  [[ $# -eq 1 ]] || { usage; exit 2; }
  local ledger=$1
  [[ -f "$ledger" ]] || fail "ledger not found: $ledger"
  command -v jq >/dev/null 2>&1 || fail 'summary requires jq'
  jq -s '
    . as $all
    | ($all | map(select(.event == "tool_end" and (.duration_ms | type == "number")))) as $tools
    | ($tools | sort_by(.duration_ms)) as $sorted
    | {
        schema_version: 1,
        records: ($all | length),
        calls: ($tools | length),
        p50_duration_ms: (if ($sorted | length) == 0 then null else $sorted[(((($sorted | length) * 0.50) | ceil) - 1)].duration_ms end),
        p95_duration_ms: (if ($sorted | length) == 0 then null else $sorted[(((($sorted | length) * 0.95) | ceil) - 1)].duration_ms end),
        cache_hits: ($all | map(select(.cache_hit == "hit")) | length),
        cache_misses: ($all | map(select(.cache_hit == "miss")) | length),
        retries: ($all | map(select(.retry_cause != "none" and .retry_cause != "")) | length),
        cohorts: ($all | map(.cohort) | unique)
      }
  ' "$ledger"
}

[[ $# -ge 1 ]] || { usage; exit 2; }
case "$1" in
  record)
    shift
    record "$@"
    ;;
  summary)
    shift
    summary "$@"
    ;;
  *)
    usage
    exit 2
    ;;
esac
