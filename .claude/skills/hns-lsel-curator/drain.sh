#!/usr/bin/env bash
# drain.sh — the mechanical drain engine for hns-lsel-curator (SPEC-LSEL-LOCAL-EVOLUTION-001 M1).
#
# Companion-offset drain of a lessons-inbox.jsonl. The frozen Go applier is NOT touched
# (REQ-LSEL-003 — bypass, never unfreeze). This script writes ONLY to user-owned state
# under --state-dir (the `.moai/state/lsel/` evolvable surface) and NEVER to memory/.
#
# Pipeline (REQ-LSEL-009 + AC-LSEL-009/010):
#   1. Read companion offset from <state-dir>/drain-offset.json (seed 0 if absent).
#   2. Slice the inbox from offset onwards (append-only inbox is preserved — no mutation).
#   3. Drain-side severity filter (AC-LSEL-010): discard Bash-timeout/sandbox noise BEFORE
#      clustering — tool_failure:Bash:UnknownFailure (the opaque ~65% timeout/sandbox
#      bucket), tool_failure:Bash:SandboxViolation (env constraint, not a code defect),
#      and any *:TimeoutError (Bash + MCP timeouts). The filter is drain-side because
#      internal/hook/failure_observer.go is OUTSIDE the six loop-writable surfaces, so
#      the loop cannot edit the writer (plan.md §F.1 [DECISION RESOLVED]).
#   4. Cluster survivors by event_key with frequency count + first/last seen + samples.
#   5. Discard singletons (frequency < 2) — single-occurrence noise per the constitution
#      Lessons Protocol drain paragraph.
#   6. Score each survivor with a Generative-Agents-style 1-10 importance gate
#      (importance = min(10, frequency); frequency as proxy, the model augments in M2+).
#   7. Emit candidates to <state-dir>/clusters.json; advance the companion offset.
#
# M1 invariant: candidates are STAGED in clusters.json only. No feedback_*.md is written
# to memory/ — APPLY lands in M3 (plan.md §F.1).
#
# Usage:
#   drain.sh --inbox <path-to-lessons-inbox.jsonl> --state-dir <path-to-lsel-state>
set -euo pipefail

INBOX=""
STATE_DIR=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --inbox)     INBOX="$2"; shift 2;;
    --state-dir) STATE_DIR="$2"; shift 2;;
    -h|--help)
      sed -n '2,30p' "$0"; exit 0;;
    *) echo "drain.sh: unknown argument: $1" >&2; exit 2;;
  esac
done

[[ -n "$INBOX" && -n "$STATE_DIR" ]] || { echo "usage: drain.sh --inbox <path> --state-dir <dir>" >&2; exit 2; }
[[ -f "$INBOX" ]]     || { echo "drain.sh: inbox not found: $INBOX" >&2; exit 2; }
command -v jq >/dev/null || { echo "drain.sh: jq is required" >&2; exit 2; }

mkdir -p "$STATE_DIR"
OFFSET_FILE="$STATE_DIR/drain-offset.json"
CLUSTERS_FILE="$STATE_DIR/clusters.json"
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# --- step 1: companion offset (seed 0 if absent) ---
if [[ -f "$OFFSET_FILE" ]]; then
  OFFSET_BEFORE="$(jq -r '(.offset // 0) | tonumber' "$OFFSET_FILE")"
else
  OFFSET_BEFORE=0
fi

# --- step 2: total inbox lines ---
TOTAL_LINES="$(wc -l <"$INBOX" | tr -d ' ')"
if [[ -z "$TOTAL_LINES" ]]; then TOTAL_LINES=0; fi

# --- empty-delta no-op: offset already at or past the inbox tail ---
if [[ "$TOTAL_LINES" -le "$OFFSET_BEFORE" ]]; then
  jq -n --arg ts "$NOW" --argjson off "$OFFSET_BEFORE" '{
    drained_at: $ts,
    offset_before: $off,
    offset_after: $off,
    total_read: 0,
    noise_discarded: 0,
    singletons_discarded: 0,
    candidates: []
  }' >"$CLUSTERS_FILE"
  echo "drain: no-op (offset=$OFFSET_BEFORE, inbox=$TOTAL_LINES lines) — nothing new to drain" >&2
  exit 0
fi

# --- steps 3-7: slice → filter → cluster → singleton-gate → importance → emit ---
# Slicing uses 1-indexed tail: stubs at lines (OFFSET_BEFORE+1)..TOTAL_LINES.
# The inbox file itself is NEVER mutated (append-only preserved; offset marks consumed stubs).
tail -n +"$((OFFSET_BEFORE + 1))" "$INBOX" \
| jq -s --arg ts "$NOW" \
       --argjson off_before "$OFFSET_BEFORE" \
       --argjson off_after  "$TOTAL_LINES" '
    . as $all
    | ($all | length) as $total_read
    | def is_noise: (
        (.event_key == "tool_failure:Bash:UnknownFailure")
        or (.event_key == "tool_failure:Bash:SandboxViolation")
        or (.event_key | endswith(":TimeoutError"))
      );
      ($all | map(select(is_noise)) | length) as $noise_count
    | ($all | map(select(is_noise | not))) as $signal
    | ($signal
        | group_by(.event_key)
        | map({
            event_key:    .[0].event_key,
            frequency:    length,
            first_seen:   (map(.timestamp) | min),
            last_seen:    (map(.timestamp) | max),
            sample_summaries: (map(.summary) | .[0:3]),
            source:       (map(.source // null) | .[0])
          })
        | . as $clusters
        | ($clusters | map(select(.frequency < 2)) | length) as $singleton_count
        | ($clusters
            | map(select(.frequency >= 2)
                | .importance = ([10, .frequency] | min)
              )
          ) as $candidates
        | {singletons: $singleton_count, candidates: $candidates}
      ) as $result
    | {
        drained_at:         $ts,
        offset_before:      $off_before,
        offset_after:       $off_after,
        total_read:         $total_read,
        noise_discarded:    $noise_count,
        singletons_discarded: $result.singletons,
        candidates:         $result.candidates
      }
  ' >"$CLUSTERS_FILE"

# --- advance the companion offset (the consumed-stub marker) ---
jq -n --argjson off "$TOTAL_LINES" --arg ts "$NOW" \
  '{offset: $off, updated: $ts}' >"$OFFSET_FILE"

# --- human-readable summary to stderr (stdout stays clean for callers) ---
jq -r '"drain: read \(.total_read) stubs (offset \(.offset_before)→\(.offset_after)) — discarded \(.noise_discarded) noise + \(.singletons_discarded) singletons; \(.candidates | length) candidate cluster(s) emitted to clusters.json"' \
  "$CLUSTERS_FILE" >&2

exit 0
