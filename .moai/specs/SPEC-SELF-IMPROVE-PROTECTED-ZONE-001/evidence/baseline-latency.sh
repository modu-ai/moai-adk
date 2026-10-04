#!/usr/bin/env bash
# baseline-latency.sh — wall time of `moai hook pre-tool` for a self-improvement-identity
# Write and for a non-identity Write, N runs each. Reports min / median / max in
# milliseconds per class. This is a baseline-first measurement: it is run once on the
# pre-implementation tree and once after M2, and the two are compared (AC-SIPZ-010).
#
# Usage (from the repository root):
#   bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/baseline-latency.sh <moai-binary> [N]
#
# The measurement includes process start-up, which dominates the hook cost; the guard's own
# contribution is the DIFFERENCE between the two builds, not the absolute number.
set -u
BIN="${1:?usage: baseline-latency.sh <moai-binary> [N]}"
N="${2:-15}"
ROOT="$(cd "$(dirname "$0")/../../../.." && pwd -P)"
cd "$ROOT" || exit 2

ms_now() { python3 -c 'import time; print(int(time.time()*1000))'; }

measure() { # <label> <agent_type|-> <path>
  local label="$1" agent="$2" path="$3" i t0 t1 out agentjson="" samples=""
  [ "$agent" != "-" ] && agentjson=",\"agent_type\":\"$agent\""
  for i in $(seq 1 "$N"); do
    t0="$(ms_now)"
    printf '%s' "{\"session_id\":\"lat\",\"hook_event_name\":\"PreToolUse\",\"cwd\":\"$ROOT\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"$path\",\"content\":\"x\"}$agentjson}" \
      | CLAUDE_PROJECT_DIR="$ROOT" "$BIN" hook pre-tool >/dev/null 2>&1
    t1="$(ms_now)"
    samples="$samples $((t1 - t0))"
  done
  printf '%s\n' "$samples" | tr ' ' '\n' | grep -v '^$' | sort -n | awk -v l="$label" '
    { a[NR]=$1 } END { printf "%s\tn=%d\tmin=%d\tmedian=%d\tmax=%d\n", l, NR, a[1], a[int((NR+1)/2)], a[NR] }'
}

measure "identity-write-outside-zone" harness-learner "internal/hook/pre_tool.go"
measure "identity-write-inside-zone"  harness-learner ".claude/hooks/moai/x.sh"
measure "non-identity-write"          -               "internal/hook/pre_tool.go"
